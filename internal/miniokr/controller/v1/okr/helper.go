// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"fmt"
	"strings"
	"time"
)

func formatRecentMonths() []string {
	now := time.Now()
	location := now.Location()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	lastMonth := currentMonth.AddDate(0, -1, 0)
	nextMonth := currentMonth.AddDate(0, 1, 0)

	layout := "2006年1月"
	return []string{
		lastMonth.Format(layout),
		currentMonth.Format(layout),
		nextMonth.Format(layout),
	}
}

func intersectStrings(sliceA, sliceB []string) []string {
	set := make(map[string]struct{}) // 使用空结构体节省内存
	intersection := []string{}

	// 将 sliceA 中的所有元素存储到集合 set 中
	for _, item := range sliceA {
		set[item] = struct{}{}
	}

	// 检查 sliceB 中的每个元素，如果它也在 set 中，则添加到交集中
	for _, item := range sliceB {
		if _, found := set[item]; found {
			intersection = append(intersection, item)
			delete(set, item) // 从 set 中移除已添加到交集的元素，防止重复添加
		}
	}

	return intersection
}

func trimIDPrefix(id string) string {
	if strings.HasPrefix(id, "o-") {
		return strings.TrimPrefix(id, "o-")
	}
	if strings.HasPrefix(id, "kr-") {
		return strings.TrimPrefix(id, "kr-")
	}
	return id
}

// standardizeMonthFormat converts a validated date string to "2006年1月".
func standardizeMonthFormat(input string) (string, error) {
	// 定义可能接受的日期格式
	layouts := []string{
		"2006-1", "2006-01", // 处理"年-月"格式
		"2006/1", "2006/01", // 处理"年/月"格式
		"2006年1月", // 处理"年年月月"格式
	}

	var parsedTime time.Time
	var err error
	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, input)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("日期格式转换错误: %v", err)
	}

	return parsedTime.Format("2006年1月"), nil
}
