// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package token

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaveTokenToFile(t *testing.T) {
	// 创建一个临时文件来测试
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // 清理测试文件

	// 测试正常情况
	data := Data{
		TenantAccessToken: "test_access_token",
		Expire:            3600,
		SavedAt:           time.Now(),
	}
	store := NewTokenStorage()
	err = store.SaveTokenToFile(data, tmpfile.Name())
	assert.NoError(t, err, "should not have an error when saving valid data")

	// 读回文件检查是否正确写入
	content, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err, "should read file without error")
	var readData Data
	err = json.Unmarshal(content, &readData)
	assert.NoError(t, err, "should unmarshal without error")
	assert.Equal(t, data.TenantAccessToken, readData.TenantAccessToken)
	assert.Equal(t, data.Expire, readData.Expire)
	assert.WithinDuration(t, data.SavedAt, readData.SavedAt, time.Second, "read data should match written data within one second")
}

func TestLoadTokenFromFile(t *testing.T) {
	// 创建并写入测试数据到临时文件
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // 清理测试文件

	data := Data{
		TenantAccessToken: "test_access_token",
		Expire:            3600,
		SavedAt:           time.Now().Truncate(time.Millisecond), // 确保时间比较不因微秒差异失败
	}
	content, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}
	if err := os.WriteFile(tmpfile.Name(), content, 0644); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	store := NewTokenStorage()

	// 测试正常情况
	loadedData, err := store.LoadTokenFromFile(tmpfile.Name())
	assert.NoError(t, err, "should not have an error when loading valid file")
	assert.Equal(t, data.TenantAccessToken, loadedData.TenantAccessToken)
	assert.Equal(t, data.Expire, loadedData.Expire)
	assert.WithinDuration(t, data.SavedAt, loadedData.SavedAt, time.Second, "loaded data should match the original data within one second")

	// 测试文件不存在的情况
	_, err = store.LoadTokenFromFile("non_existent_file.json")
	assert.ErrorIs(t, err, os.ErrNotExist, "should return os.ErrNotExist for non-existent files")

	// 测试 JSON 解析错误
	if err := os.WriteFile(tmpfile.Name(), []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid json to temp file: %v", err)
	}
	_, err = store.LoadTokenFromFile(tmpfile.Name())
	assert.Error(t, err, "should have an error due to invalid JSON")
}
