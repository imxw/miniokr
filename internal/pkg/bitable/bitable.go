// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package bitable

import (
	"context"
	"errors"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"

	"github.com/imxw/miniokr/internal/pkg/bitable/token"
	"github.com/imxw/miniokr/internal/pkg/log"
)

var ErrInvalidUser = errors.New("用户无效,请联系管理员")

type RecordManager struct {
	Client        *lark.Client
	AppToken      string
	tokenProvider token.Provider
}

func NewRecordManager(client *lark.Client, appToken string, tp token.Provider) *RecordManager {
	return &RecordManager{
		Client:        client,
		AppToken:      appToken,
		tokenProvider: tp,
	}
}

func (r *RecordManager) CreateRecord(ctx context.Context, tableID string, fields map[string]interface{}) (string, error) {
	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(r.AppToken).
		TableId(tableID).
		AppTableRecord(larkbitable.NewAppTableRecordBuilder().
			Fields(fields).Build()).Build()

	// fmt.Println("查看当前fields", larkcore.Prettify(fields))
	t, err := r.tokenProvider.EnsureValidToken(ctx)
	if err != nil {
		log.C(ctx).Errorw("获取token失败", "error", err)
		return "", errors.New("获取token失败")
	}
	resp, err := r.Client.Bitable.AppTableRecord.Create(context.Background(), req, larkcore.WithTenantAccessToken(t))
	if err != nil {
		log.C(ctx).Errorw("failed to create record", "error", err, "tableID", tableID)
		return "", fmt.Errorf("failed to create record: %v", err)
	}

	if !resp.Success() {
		log.C(ctx).Errorw("failed to create record", "errMsg", resp.Msg, "tableID", tableID)
		return "", fmt.Errorf("failed to create record: code=%d, msg=%s, requestId=%s\n", resp.Code, resp.Msg, resp.RequestId())
	}

	return *resp.Data.Record.RecordId, nil
}

func (r *RecordManager) UpdateRecord(ctx context.Context, tableID string, recordID string, fields map[string]interface{}) error {
	t, err := r.tokenProvider.EnsureValidToken(ctx)
	if err != nil {
		log.C(ctx).Errorw("获取token失败", "error", err)
		return errors.New("获取token失败")
	}

	// fmt.Println("查看当前fields", recordID, larkcore.Prettify(fields))
	req := larkbitable.NewUpdateAppTableRecordReqBuilder().
		AppToken(r.AppToken).
		TableId(tableID).
		RecordId(recordID).
		AppTableRecord(larkbitable.NewAppTableRecordBuilder().Fields(fields).Build()).Build()
	resp, err := r.Client.Bitable.AppTableRecord.Update(ctx, req, larkcore.WithTenantAccessToken(t))
	if err != nil {
		log.C(ctx).Errorw("failed to update record", "error", err, "tableID", tableID, "recordID", recordID)
		return fmt.Errorf("failed to update record: %v", err)
	}
	if !resp.Success() {
		log.C(ctx).Errorw("failed to update record", "errMsg", resp.Msg, "tableID", tableID, "recordID", recordID)
		return fmt.Errorf("failed to update record: code=%d, msg=%s, requestId=%s\n", resp.Code, resp.Msg, resp.RequestId())
	}
	return nil
}

func (r *RecordManager) DeleteRecord(ctx context.Context, tableID, recordID string) error {
	t, err := r.tokenProvider.EnsureValidToken(ctx)
	if err != nil {
		log.C(ctx).Errorw("获取token失败", "error", err)
		return errors.New("获取token失败")
	}

	// fmt.Println("查看传入的id: ", recordID)
	req := larkbitable.NewDeleteAppTableRecordReqBuilder().
		AppToken(r.AppToken).
		TableId(tableID).
		RecordId(recordID).Build()

	resp, err := r.Client.Bitable.AppTableRecord.Delete(ctx, req, larkcore.WithTenantAccessToken(t))
	if err != nil {
		log.C(ctx).Errorw("failed to delete record", "error", err, "recordId", recordID, "tableId", tableID)
		return fmt.Errorf("failed to delete record: %v", err)
	}
	if !resp.Success() {
		log.C(ctx).Errorw("failed to delete record", "error", resp.CodeError.Error(), "recordId", recordID, "tableId", tableID)
		return fmt.Errorf("failed to delete record: %s", resp.CodeError.Error())
	}
	return nil
}

func (r *RecordManager) BatchDeleteRecord(ctx context.Context, tableID string, records []string) error {
	t, err := r.tokenProvider.EnsureValidToken(ctx)
	if err != nil {
		log.C(ctx).Errorw("获取token失败", "error", err)
		return errors.New("获取token失败")
	}

	req := larkbitable.NewBatchDeleteAppTableRecordReqBuilder().
		AppToken(r.AppToken).
		TableId(tableID).
		Body(larkbitable.NewBatchDeleteAppTableRecordReqBodyBuilder().
			Records(records).
			Build()).Build()

	resp, err := r.Client.Bitable.AppTableRecord.BatchDelete(ctx, req, larkcore.WithTenantAccessToken(t))
	if err != nil {
		log.C(ctx).Errorw("failed to delete records", "error", err, "tableId", tableID, "records", records)
		return fmt.Errorf("failed to delete records: %v", err)
	}
	if !resp.Success() {
		log.C(ctx).Errorw("failed to delete records", "error", resp.CodeError.Error(), "tableId", tableID, "records", records)
		return fmt.Errorf("failed to delete records: %s", resp.CodeError.Error())
	}
	return nil
}

func (r *RecordManager) SearchRecord(ctx context.Context, tableID string, fieldNames []string, filter *larkbitable.FilterInfo) ([]*larkbitable.AppTableRecord, error) {
	t, err := r.tokenProvider.EnsureValidToken(ctx)
	if err != nil {
		log.C(ctx).Errorw("获取token失败", "error", err)
		return nil, errors.New("获取token失败")
	}

	// 创建请求对象
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(r.AppToken).
		TableId(tableID).
		PageSize(100).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			FieldNames(fieldNames).
			Filter(filter).
			AutomaticFields(true).
			Build()).
		Build()
	// 发起请求
	resp, err := r.Client.Bitable.AppTableRecord.Search(context.Background(), req, larkcore.WithTenantAccessToken(t))

	// 处理错误
	if err != nil {
		// TODO: 优化错误处理
		log.C(ctx).Errorw("查询错误", "error", err)
		return nil, err
	}
	// 服务端错误处理
	if !resp.Success() {

		if resp.Code == 1254018 || resp.Msg == "InvalidFilter" {

			return nil, ErrInvalidUser
		}

		log.C(ctx).Errorw("failed to list record", "errMsg", resp.Msg, "tableID", tableID)

		return nil, fmt.Errorf("failed to list record: code=%d, msg=%s, requestId=%s\n", resp.Code, resp.Msg, resp.RequestId())
	}

	return resp.Data.Items, nil
}

func (r *RecordManager) SearchRecordByUser(ctx context.Context, tableID string, username string, fieldNames []string) ([]*larkbitable.AppTableRecord, error) {

	OC := []*larkbitable.Condition{
		larkbitable.NewConditionBuilder().FieldName("员工姓名").Operator(`is`).Value([]string{username}).Build(),
	}

	// if len(months) != 0 {
	// 	OC = append(OC, larkbitable.NewConditionBuilder().FieldName("考核月份").Operator(`contains`).Value(months).Build())
	// }

	filter := larkbitable.NewFilterInfoBuilder().
		Conjunction(`and`).
		Conditions(OC).
		Build()

	return r.SearchRecord(ctx, tableID, fieldNames, filter)

}
