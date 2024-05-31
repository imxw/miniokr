// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"

	"github.com/imxw/miniokr/internal/pkg/model"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

func convertToObjective(data []*larkbitable.AppTableRecord, of *v1.ObjectiveField, sortBy string, orderBy string) []model.Objective {
	var objectives []model.Objective
	for _, record := range data {
		if record.RecordId == nil {
			continue // Skip record without a valid RecordId
		}
		fields := record.Fields

		// 安全提取 Title 字段
		title, ok := fields[of.Title].([]interface{})
		if !ok {
			title = []interface{}{} // 默认为空切片，避免 panic
		}

		objective := model.Objective{
			ID:               OPrefix + *record.RecordId,
			Owner:            extractText(fields[of.Owner].([]interface{})),
			Date:             extractString(fields, of.Date),
			Weight:           extractFloatToInt(fields, of.Weight),
			KrsIds:           extractRecordIDs(fields, of.KeyResultIDs, KrPrefix),
			CreatedTime:      safeInt64(record.CreatedTime),
			LastModifiedTime: safeInt64(record.LastModifiedTime),
			//Title:            extractText(fields[of.Title].([]interface{})),
			Title: extractText(title),
		}
		objectives = append(objectives, objective)
	}

	sortObjectives(objectives, sortBy, orderBy)
	return objectives
}

func convertToKeyResult(data []*larkbitable.AppTableRecord, of *v1.KeyResultField, sortBy string, orderBy string) []model.KeyResult {
	var krs []model.KeyResult
	for _, record := range data {
		if record.RecordId == nil {
			continue
		}
		fields := record.Fields

		// 安全提取 Criteria 字段
		criteria, ok := fields[of.Criteria].([]interface{})
		if !ok {
			criteria = []interface{}{} // 默认为空切片，避免 panic
		}

		// 安全提取 Title 字段
		title, ok := fields[of.Title].([]interface{})
		if !ok {
			title = []interface{}{} // 默认为空切片，避免 panic
		}

		// 安全提取 Reason 字段
		reason, ok := fields[of.Reason].([]interface{})
		if !ok {
			reason = []interface{}{} // 默认为空切片，避免 panic
		}

		// 安全提取 Objective ID
		objectiveIDs := extractRecordIDs(fields, of.ObjectiveID, OPrefix)
		var objectiveID string
		if len(objectiveIDs) > 0 {
			objectiveID = objectiveIDs[0]
		}

		kr := model.KeyResult{
			ID:               KrPrefix + *record.RecordId,
			Owner:            extractText(fields[of.Owner].([]interface{})),
			Date:             extractString(fields, of.Date),
			Weight:           extractFloatToInt(fields, of.Weight),
			Completed:        extractString(fields, of.Completed),
			ObjectiveID:      objectiveID,
			Criteria:         extractAllTexts(criteria),
			SelfRating:       extractRating(fields, of.SelfRating),
			Reason:           extractAllTexts(reason),
			CreatedTime:      safeInt64(record.CreatedTime),
			LastModifiedTime: safeInt64(record.LastModifiedTime),
			Title:            extractText(title),
			LeaderRating:     extractRating(fields, of.LeaderRating),
			Department:       extractFirstDepartment(fields, of.Department),
			Leader:           extractLeaderName(fields, of.Leader),
		}

		krs = append(krs, kr)
	}
	sortKeyResults(krs, sortBy, orderBy)
	return krs
}

func extractText(value []interface{}) string {
	for _, item := range value {
		if m, ok := item.(map[string]interface{}); ok {
			if text, ok := m["text"].(string); ok {
				return text
			}
		}
	}
	return ""
}

func extractRating(fields map[string]interface{}, key string) *int {
	if value, ok := fields[key].(float64); ok {
		rating := int(value*100) / 100
		return &rating
	}

	return nil
}

func extractAllTexts(items []interface{}) string {
	var texts []string
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			if text, ok := m["text"].(string); ok {
				texts = append(texts, text)
			}
		}
	}
	// 直接连接，不添加分隔符
	return strings.Join(texts, "")
}

func extractString(fields map[string]interface{}, key string) string {
	if value, ok := fields[key].(string); ok {
		return value
	}
	return ""
}

// extractRecordIDs 从给定的字段数据中提取 record_id 列表。
func extractRecordIDs(fields map[string]interface{}, key, prefix string) []string {
	var recordIDs []string
	if nestedMap, ok := fields[key].(map[string]interface{}); ok {
		if ids, ok := nestedMap["link_record_ids"].([]interface{}); ok {
			for _, id := range ids {
				if strID, ok := id.(string); ok {
					recordIDs = append(recordIDs, prefix+strID)
				}
			}
		}
	}
	return recordIDs
}

func extractFloatToInt(fields map[string]interface{}, key string) int {
	if value, ok := fields[key].(float64); ok {
		return int(value * 100)
	}
	return 0
}

func extractFirstDepartment(fields map[string]interface{}, key string) string {
	if nestedMap, ok := fields[key].(map[string]interface{}); ok {
		if ds, ok := nestedMap["value"].([]interface{}); ok {
			if len(ds) > 0 {
				if strD, ok := ds[0].(string); ok {
					return strD
				}
			}
		}
	}
	return "" // 如果遇到任何问题，返回空字符串
}

func extractLeaderName(data map[string]interface{}, key string) string {
	// 获取 key 对应的值，假设它是一个 map 类型
	if nestedMap, ok := data[key].(map[string]interface{}); ok {
		// 获取 "value" 对应的切片，假设它是 []interface{} 类型
		if values, ok := nestedMap["value"].([]interface{}); ok {
			// 检查切片至少有一个元素
			if len(values) > 0 {
				// 假设第一个元素是一个 map 类型，且包含 "name" 键
				if leaderInfo, ok := values[0].(map[string]interface{}); ok {
					// 从 leaderInfo 中获取 "name"，假设它是 string 类型
					if name, ok := leaderInfo["name"].(string); ok {
						return name
					}
				}
			}
		}
	}
	return "" // 如果任何步骤失败，返回空字符串
}

func safeInt64(ptr *int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return 0 // 默认值，可根据业务需要调整
}

// Dynamic sorting function, reused by both conversion functions
func sortObjectives(objectives []model.Objective, sortBy string, orderBy string) {
	sort.Slice(objectives, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			numI := extractNumberFromString(objectives[i].Title)
			numJ := extractNumberFromString(objectives[j].Title)
			less = numI < numJ
		case "createdTime":
			less = objectives[i].CreatedTime < objectives[j].CreatedTime
		case "lastModifiedTime":
			less = objectives[i].LastModifiedTime < objectives[j].LastModifiedTime
		default:
			less = objectives[i].CreatedTime < objectives[j].CreatedTime // 默认按创建时间排序
		}

		if orderBy == "desc" {
			return !less // 降序：返回反向结果
		}
		return less // 升序
	})
}

func sortKeyResults(krs []model.KeyResult, sortBy string, orderBy string) {
	sort.Slice(krs, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			// 提取标题中的数字进行排序
			numI := extractNumberFromString(krs[i].Title)
			numJ := extractNumberFromString(krs[j].Title)
			less = numI < numJ
		case "createtime":
			// 根据创建时间排序
			less = krs[i].CreatedTime < krs[j].CreatedTime
		case "updatetime":
			// 根据最后修改时间排序
			less = krs[i].LastModifiedTime < krs[j].LastModifiedTime
		default:
			// 默认排序依据创建时间
			less = krs[i].CreatedTime < krs[j].CreatedTime
		}

		if orderBy == "desc" {
			return !less // 如果是降序，返回反向结果
		}
		return less // 默认为升序
	})
}

// extractNumberFromString 提取字符串中的第一个连续数字序列
func extractNumberFromString(s string) int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindStringSubmatch(s)
	if len(matches) == 0 {
		return -1 // 如果没有找到数字，返回-1作为错误标识
	}
	num, err := strconv.Atoi(matches[0])
	if err != nil {
		return -1 // 转换错误处理
	}
	return num
}
