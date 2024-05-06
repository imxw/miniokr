// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package field

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/xwlearn/miniokr/internal/pkg/bitable/field"
	v1 "github.com/xwlearn/miniokr/pkg/api/miniokr/v1"
)

type FeishuFieldService struct {
	OTableID, KrTableID string
	FieldManager        *field.Manager
}

// NewFeishufieldService 创建一个新的FeishuFieldService实例
func NewFeishufieldService(OTableID, KrTableID string, manager *field.Manager) (*FeishuFieldService, error) {
	if OTableID == "" {
		return nil, errors.New("OTableID cannot be empty")
	}
	if KrTableID == "" {
		return nil, errors.New("KrTableID cannot be empty")
	}
	if manager == nil {
		return nil, errors.New("manager cannot be nil")
	}

	return &FeishuFieldService{
		OTableID:     OTableID,
		KrTableID:    KrTableID,
		FieldManager: manager,
	}, nil
}

func (f *FeishuFieldService) GetFieldDefinitions(ctx context.Context) (v1.FieldMappingsResponse, error) {
	var objectiveFields v1.ObjectiveField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, f.OTableID, &objectiveFields); err != nil {
		return v1.FieldMappingsResponse{}, err
	}

	var keyResultFields v1.KeyResultField
	if err := f.FieldManager.GetFriendlyFieldMapping(ctx, f.KrTableID, &keyResultFields); err != nil {
		return v1.FieldMappingsResponse{}, err

	}

	response := v1.FieldMappingsResponse{
		Objective: objectiveFields,
		KeyResult: keyResultFields,
	}

	return response, nil
}

func (f *FeishuFieldService) GetValidDates(ctx context.Context) ([]string, error) {

	dates, err := f.FieldManager.ExtractFieldOptions(ctx, f.OTableID, "考核月份")
	if err != nil {
		return nil, fmt.Errorf("获取考核月份列表失败: %v", err)
	}
	// 解析日期并排序
	var dateTimes []time.Time
	for _, dateStr := range dates {
		date, err := parseChineseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("解析日期失败 '%s': %v", dateStr, err)
		}
		dateTimes = append(dateTimes, date)
	}

	sort.Slice(dateTimes, func(i, j int) bool {
		return dateTimes[i].Before(dateTimes[j])
	})

	// 将排序后的时间对象转换回字符串
	var sortedDates []string
	for _, date := range dateTimes {
		sortedDates = append(sortedDates, date.Format("2006年1月"))
	}

	return sortedDates, nil
}

func (f *FeishuFieldService) GetValidUsers(ctx context.Context) ([]string, error) {

	users, err := f.FieldManager.ExtractFieldOptions(ctx, f.OTableID, "员工姓名")
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %v", err)
	}
	var validUsers []string
	for _, user := range users {
		if validateChineseName(user) {
			validUsers = append(validUsers, user)
		}
	}

	return validUsers, nil
}

// parseChineseDate 将中文日期格式（"2024年2月"）解析为time.Time
func parseChineseDate(dateStr string) (time.Time, error) {
	// 替换中文日期中的年和月，以符合Parse的要求
	replacer := strings.NewReplacer("年", "-", "月", "")
	dateStr = replacer.Replace(dateStr)

	// 解析日期，这里假设所有日期都是月的第一天
	return time.Parse("2006-1", dateStr)
}

// validateChineseName 校验中文用户名是否符合要求
func validateChineseName(name string) bool {
	// 正则表达式匹配中文字符且长度为1到4
	match, _ := regexp.MatchString("^[\u4e00-\u9fa5]{1,4}$", name)
	return match
}
