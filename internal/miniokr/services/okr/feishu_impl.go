// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"context"
	"errors"
	"fmt"

	"github.com/imxw/miniokr/internal/pkg/bitable"
	"github.com/imxw/miniokr/internal/pkg/bitable/field"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/internal/pkg/model"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

const KrPrefix = "kr-"
const OPrefix = "o-"

type FeishuOkrService struct {
	OTableID, KrTableID string
	FieldManager        *field.Manager
	RecordManager       *bitable.RecordManager
}

func NewFeishuOkrService(OTableID, KrTableID string, fm *field.Manager, rm *bitable.RecordManager) (*FeishuOkrService, error) {
	if OTableID == "" {
		return nil, errors.New("OTableID cannot be empty")
	}
	if KrTableID == "" {
		return nil, errors.New("KrTableID cannot be empty")
	}
	if fm == nil || rm == nil {
		return nil, errors.New("fm or rm cannot be nil")
	}

	return &FeishuOkrService{
		OTableID:      OTableID,
		KrTableID:     KrTableID,
		FieldManager:  fm,
		RecordManager: rm,
	}, nil
}

func (f *FeishuOkrService) ListObjectivesByOwner(ctx context.Context, username string, sortBy string, orderBy string) ([]model.Objective, error) {

	tableID := f.OTableID

	var friendlyMapping v1.ObjectiveField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, tableID, &friendlyMapping); err != nil {
		return nil, err
	}

	oResp, err := f.RecordManager.SearchRecordByUser(ctx, tableID, username, nil)
	if err != nil {
		return nil, err
	}

	return convertToObjective(oResp, &friendlyMapping, sortBy, orderBy), nil

}
func (f *FeishuOkrService) ListKeyResultsByOwner(ctx context.Context, username string, sortBy string, orderBy string) ([]model.KeyResult, error) {

	tableID := f.KrTableID

	var friendlyMapping v1.KeyResultField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, tableID, &friendlyMapping); err != nil {
		return nil, err
	}

	KrResp, err := f.RecordManager.SearchRecordByUser(ctx, tableID, username, nil)
	if err != nil {
		return nil, err
	}
	return convertToKeyResult(KrResp, &friendlyMapping, sortBy, orderBy), nil
}

func (f *FeishuOkrService) CreateObjective(ctx context.Context, objective model.Objective) (string, error) {

	tableID := f.OTableID

	var friendlyMapping v1.ObjectiveField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, tableID, &friendlyMapping); err != nil {
		return "", err
	}

	fields := make(map[string]interface{})
	fields[friendlyMapping.Title] = objective.Title
	fields[friendlyMapping.Date] = objective.Date
	fields[friendlyMapping.Owner] = objective.Owner
	fields[friendlyMapping.Weight] = float32(objective.Weight) / 100.0

	return f.RecordManager.CreateRecord(ctx, tableID, fields)

}
func (f *FeishuOkrService) UpdateObjective(ctx context.Context, objective model.Objective) error {
	tableID := f.OTableID

	var friendlyMapping v1.ObjectiveField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, tableID, &friendlyMapping); err != nil {
		return err
	}
	fields := make(map[string]interface{})
	fields[friendlyMapping.Title] = objective.Title
	fields[friendlyMapping.Date] = objective.Date
	fields[friendlyMapping.Owner] = objective.Owner
	fields[friendlyMapping.Weight] = float32(objective.Weight) / 100

	return f.RecordManager.UpdateRecord(ctx, tableID, objective.ID, fields)
}
func (f *FeishuOkrService) DeleteObjectiveByID(ctx context.Context, oid string, krids []string) error {
	tableID := f.OTableID
	krTableID := f.KrTableID

	var err error

	// 删除目标记录
	err = f.RecordManager.DeleteRecord(ctx, tableID, oid)
	if err != nil {
		return fmt.Errorf("failed to delete objective: %v", err)
	}
	// 如果有关键结果 ID，则批量删除
	if len(krids) > 0 {
		err = f.RecordManager.BatchDeleteRecord(ctx, krTableID, krids)
		if err != nil {
			return fmt.Errorf("failed to delete key results: %v", err)
		}
	}

	return nil

}
func (f *FeishuOkrService) CreateKeyResult(ctx context.Context, keyResult model.KeyResult) (string, error) {
	tableID := f.KrTableID

	var friendlyMapping v1.KeyResultField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, tableID, &friendlyMapping); err != nil {
		return "", err
	}

	fields := make(map[string]interface{})
	fields[friendlyMapping.Title] = keyResult.Title
	fields[friendlyMapping.Date] = keyResult.Date
	fields[friendlyMapping.Owner] = keyResult.Owner
	fields[friendlyMapping.Weight] = float32(keyResult.Weight) / 100.0
	fields[friendlyMapping.Completed] = keyResult.Completed
	fields[friendlyMapping.SelfRating] = keyResult.SelfRating
	fields[friendlyMapping.Reason] = keyResult.Reason
	fields[friendlyMapping.Criteria] = keyResult.Criteria

	fields[friendlyMapping.ObjectiveID] = []string{keyResult.ObjectiveID}
	return f.RecordManager.CreateRecord(ctx, tableID, fields)
}
func (f *FeishuOkrService) UpdateKeyResult(ctx context.Context, keyResult model.KeyResult) error {

	log.C(ctx).Debugw("Service UpdateKeyResult called")
	tableID := f.KrTableID

	var friendlyMapping v1.KeyResultField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, tableID, &friendlyMapping); err != nil {
		return err
	}
	fields := make(map[string]interface{})
	fields[friendlyMapping.Title] = keyResult.Title
	fields[friendlyMapping.Date] = keyResult.Date
	fields[friendlyMapping.Owner] = keyResult.Owner
	fields[friendlyMapping.Weight] = float32(keyResult.Weight) / 100
	fields[friendlyMapping.Completed] = keyResult.Completed
	fields[friendlyMapping.SelfRating] = keyResult.SelfRating
	fields[friendlyMapping.Reason] = keyResult.Reason
	fields[friendlyMapping.Criteria] = keyResult.Criteria

	if keyResult.LeaderRating != nil {
		fields[friendlyMapping.LeaderRating] = keyResult.LeaderRating
	}

	// TODO: 单独处理
	if keyResult.ObjectiveID != "" {

		fields[friendlyMapping.ObjectiveID] = []string{keyResult.ObjectiveID}
	}

	return f.RecordManager.UpdateRecord(ctx, tableID, keyResult.ID, fields)
}
func (f *FeishuOkrService) DeleteKeyResultByID(ctx context.Context, id string) error {
	tableID := f.KrTableID
	return f.RecordManager.DeleteRecord(ctx, tableID, id)

}
