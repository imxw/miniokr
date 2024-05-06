// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package field

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"

	"github.com/imxw/miniokr/internal/pkg/bitable/token"
	"github.com/imxw/miniokr/internal/pkg/log"
)

type Manager struct {
	Client        *lark.Client
	AppToken      string
	tokenProvider token.Provider
	fieldCache    map[string][]Field
	lastUpdate    time.Time
	mutex         sync.RWMutex
}

type Option struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Field struct {
	FieldID   string   `json:"field_id"`
	FieldName string   `json:"field_name"`
	Options   []Option `json:"options,omitempty"` // omitempty表示如果Options为空，则在JSON中省略该字段
}

type MappingData struct {
	FieldMapping []Field `json:"field_mapping"`
	LastUpdate   int64   `json:"last_update"`
}

const UpdateInterval = 24 * time.Hour // 定义更新间隔

// 根据表ID生成字段映射文件的路径
func getFieldMappingFilePath(tableID string) string {
	return "field_mapping_" + tableID + ".json"
}

func NewManager(client *lark.Client, appToken string, tokenProvider token.Provider) *Manager {
	m := &Manager{
		Client:        client,
		AppToken:      appToken,
		tokenProvider: tokenProvider,
		fieldCache:    make(map[string][]Field, 0),
		lastUpdate:    time.Time{},
	}
	return m
}

func (m *Manager) LoadOrRefreshFieldMapping(ctx context.Context, tableID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	filePath := getFieldMappingFilePath(tableID)
	mappingData, err := LoadFieldMappingFromFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			m.RefreshAndSaveFieldMapping(ctx, tableID)
		} else {
			fmt.Printf("Error loading field mapping: %v\n", err)
		}
	} else if time.Since(time.Unix(mappingData.LastUpdate, 0)) > UpdateInterval {
		m.RefreshAndSaveFieldMapping(ctx, tableID)
	} else {
		m.fieldCache[tableID] = mappingData.FieldMapping
		m.lastUpdate = time.Unix(mappingData.LastUpdate, 0)
	}
}

// GetFieldMapping 从本地文件加载字段映射，如果不存在或过期则从API刷新
func (m *Manager) GetFieldMapping(ctx context.Context, tableID string) ([]Field, error) {
	m.mutex.RLock() // 对检查操作加读锁
	// 检查内存中的缓存是否有效
	if fields, ok := m.fieldCache[tableID]; ok && time.Since(m.lastUpdate) < UpdateInterval {
		cacheCopy := make([]Field, len(fields))
		copy(cacheCopy, fields) // 创建缓存的副本，避免外部修改
		m.mutex.RUnlock()       // 读取完成，释放读锁
		return cacheCopy, nil
	}
	m.mutex.RUnlock() // 读取完成，释放读锁

	m.mutex.Lock() // 对更新操作加写锁
	// 再次检查缓存，防止在获取锁的过程中缓存已被更新
	if fields, ok := m.fieldCache[tableID]; ok && time.Since(m.lastUpdate) < UpdateInterval {
		cacheCopy := make([]Field, len(fields))
		copy(cacheCopy, fields)
		m.mutex.Unlock()
		return cacheCopy, nil
	}

	// 从文件中加载字段映射数据
	filePath := getFieldMappingFilePath(tableID)
	mappingData, err := LoadFieldMappingFromFile(filePath)
	if err != nil || time.Since(time.Unix(mappingData.LastUpdate, 0)) > UpdateInterval {
		m.mutex.Unlock() // 释放写锁前启动刷新
		return m.RefreshAndSaveFieldMapping(ctx, tableID)
	}

	// 更新内存缓存并返回结果
	m.fieldCache[tableID] = mappingData.FieldMapping
	m.lastUpdate = time.Unix(mappingData.LastUpdate, 0)
	m.mutex.Unlock() // 更新完成，释放写锁
	return mappingData.FieldMapping, nil
}

// fieldMapping 获取飞书多维表格中字段映射
func (m *Manager) fieldMapping(ctx context.Context, tableID string) ([]Field, error) {
	// Ensure a valid token is available
	t, err := m.tokenProvider.EnsureValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("error ensuring valid token: %v", err)
	}

	// Build the request with provided table ID and app token
	req := larkbitable.NewListAppTableFieldReqBuilder().
		AppToken(m.AppToken).
		TableId(tableID).
		Build()

	// List fields using the bitable app table field API
	resp, err := m.Client.Bitable.AppTableField.List(ctx, req, larkcore.WithTenantAccessToken(t))
	if err != nil {
		return nil, fmt.Errorf("failed to list fields: %v", err)
	}

	// Check the success of the response
	if !resp.Success() {
		return nil, fmt.Errorf("failed to list fields: %s", larkcore.Prettify(resp))
	}

	var fields []Field
	for _, item := range resp.Data.Items {
		// Skip computed fields
		if *item.Type == 19 || *item.Type == 20 {
			continue
		}

		// Ensure both field ID and name are present
		if item.FieldId != nil && item.FieldName != nil {
			field := Field{
				FieldID:   *item.FieldId,
				FieldName: *item.FieldName,
			}

			// Handle options if they are present
			if item.Property != nil && item.Property.Options != nil {
				for _, opt := range item.Property.Options {
					if opt.Id != nil && opt.Name != nil {
						field.Options = append(field.Options, Option{
							ID:   *opt.Id,
							Name: *opt.Name,
						})
					}
				}
			}

			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (m *Manager) RefreshAndSaveFieldMapping(ctx context.Context, tableID string) ([]Field, error) {
	fieldMapping, err := m.fieldMapping(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("error refreshing field mapping: %v", err)
	}

	// 保存字段映射到文件
	mappingData := MappingData{
		FieldMapping: fieldMapping,
		LastUpdate:   time.Now().Unix(),
	}
	filePath := getFieldMappingFilePath(tableID)
	if err = SaveFieldMappingToFile(mappingData, filePath); err != nil {
		return nil, fmt.Errorf("error saving field mapping to file: %v", err)
	}

	// 更新内存缓存
	m.fieldCache[tableID] = fieldMapping
	m.lastUpdate = time.Now()

	return fieldMapping, nil
}

// GetFriendlyFieldMapping 获取友好的字段映射
func (m *Manager) GetFriendlyFieldMapping(ctx context.Context, tableID string, obj interface{}) error {

	fielding, err := m.GetFieldMapping(ctx, tableID)
	if err != nil {
		return err
	}

	var friendlyMapping = make(map[string]string, len(fielding))

	for _, x := range fielding {
		friendlyMapping[x.FieldID] = x.FieldName
	}
	err = mapFields(obj, friendlyMapping)
	if err != nil {
		log.C(ctx).Errorw("Error mapping fields", "error", err)
	}

	return nil
}

// ExtractFieldOptions 根据字段名从字段映射中提取选项
func (m *Manager) ExtractFieldOptions(ctx context.Context, tableId, fieldName string) ([]string, error) {
	fieldMappings, err := m.GetFieldMapping(ctx, tableId)
	if err != nil {
		return nil, err
	}

	// 遍历字段映射查找匹配的字段名并返回其选项名称
	for _, field := range fieldMappings {
		if field.FieldName == fieldName {
			var names []string
			for _, option := range field.Options {
				names = append(names, option.Name)
			}
			return names, nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found", fieldName)
}

// mapFields maps data from a map with field IDs as keys to the Objective struct using field tags.
func mapFields(obj interface{}, data map[string]string) error {
	t := reflect.TypeOf(obj).Elem()
	v := reflect.ValueOf(obj).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldID := field.Tag.Get("field")
		if value, ok := data[fieldID]; ok {
			v.Field(i).SetString(value) // Directly set the string since data is map[string]string
		} else {
			// Consider whether you want to report missing fields as an error
			fmt.Printf("No data provided for field '%s' (field ID: '%s') - %+v\n", field.Name, fieldID, obj)
		}
	}
	return nil
}

//func (m *Manager) StartFieldSync(ctx context.Context) {
//	ticker := time.NewTicker(UpdateInterval)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			_, err := m.RefreshAndSaveFieldMapping(ctx,)
//			if err != nil {
//				return
//			}
//		case <-ctx.Done():
//			return
//		}
//	}
//}
