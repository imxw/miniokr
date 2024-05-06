// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package field

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/stretchr/testify/assert"

	"github.com/imxw/miniokr/internal/pkg/bitable/token"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

var appID, appSecret, appToken string

// 从环境变量获取配置信息
func getLarkClient(t *testing.T) *lark.Client {
	appID = os.Getenv("APP_ID")
	appSecret = os.Getenv("APP_SECRET")
	appToken = os.Getenv("APP_TOKEN")
	if appID == "" || appSecret == "" || appToken == "" {
		t.Skip("APP_ID or APP_SECRET or APP_TOKEN  is not set. Skipping integration test.")
	}
	return lark.NewClient(appID, appSecret)
}

func TestGetNewTokenIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 初始化 Lark 客户端
	client := getLarkClient(t)

	// 创建服务
	lts := &token.LarkTokenService{Client: client}

	// 调用服务
	resp, err := lts.GetNewToken(ctx, appID, appSecret)
	assert.NoError(t, err, "获取Token时不应有错误发生")
	assert.NotEmpty(t, resp.Data.TenantAccessToken, "获取的Token不应为空")

	t.Logf("Received Token: %s\n", resp.Data.TenantAccessToken)
}

func TestGetFieldMapping(t *testing.T) {
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	ctx := context.Background()

	client := getLarkClient(t)
	tokenService := &token.LarkTokenService{Client: client}
	store := token.NewTokenStorage()
	tm := token.NewManager(tokenService, store, nil, appID, appSecret)
	err := tm.Initialize(ctx)
	if err != nil {
		return
	}

	fm := NewManager(client, appToken, tm)

	OTableID := "tblMm4wh9xmwRFqk"
	//KRTableID := "tblyRhpeJMvFK2PS"
	//OFieldPath := getFieldMappingFilePath(OTableID)
	fm.LoadOrRefreshFieldMapping(ctx, OTableID)

	//_, err := os.Stat(OFieldPath)
	//t.Log(OFieldPath)
	//if os.IsNotExist(err) {
	//	// 如果文件不存在，你可以这样断言
	//	assert.Error(t, err, "File should exist")
	//} else {
	//	// 如果文件存在，断言没有错误
	//	//assert.NoError(t, err, "File should exist and be accessible")
	//	printFileContents(OTableID)
	//}
	//printFileContents(OFieldPath)

	//fm.GetFieldMapping(ctx, OTableID)
}

func printFileContents(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist.")
		} else {
			fmt.Println("Error opening file:", err)
		}
		return
	}
	defer file.Close()

	// 读取文件内容并打印
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println("File contents:")
	fmt.Println(string(data))
}

//	type Objective struct {
//		Title  string `json:"title" field:"fldp1iQXFv"`
//		Owner  string `json:"owner" field:"fld8vFDsXz"`
//		Date   string `json:"date" field:"fldc36J6LW"`
//		Weight string `json:"weight" field:"fldgooDzO7"`
//	}
//
//	type KeyResult struct {
//		Title            string `json:"title" field:"fldrxgL9LV"`
//		Owner            string `json:"owner" field:"fldjEZmY3S"`
//		Date             string `json:"date" field:"fldnMxJlKj"`
//		Weight           string `json:"weight" field:"fld5HXzwmN"`
//		Completed        string `json:"completed" field:"fldG2nSDTZ"`
//		SelfRating       string `json:"selfRating" field:"fldvjvtxRr"`
//		Criteria         string `json:"criteria" field:"fldPGxpg2b"`
//		UnfinishedReason string `json:"unfinishedReason" field:"fld6OsYad8"`
//	}
//
// //// mapFields maps data from a map with field IDs as keys to the Objective struct using field tags.
// //func mapFields(obj interface{}, data map[string]interface{}) error {
// //	t := reflect.TypeOf(obj).Elem()
// //	v := reflect.ValueOf(obj).Elem()
// //
// //	for i := 0; i < t.NumField(); i++ {
// //		field := t.Field(i)
// //		fieldID := field.Tag.Get("field")
// //		if value, ok := data[fieldID]; ok {
// //			if strVal, ok := value.(string); ok {
// //				v.Field(i).SetString(strVal)
// //			} else {
// //				return fmt.Errorf("expected string value for field '%s', but got different type", field.Name)
// //			}
// //		}
// //	}
// //	return nil
// //}
func TestFieldMap(t *testing.T) {

	ctx := context.Background()

	client := getLarkClient(t)
	tokenService := &token.LarkTokenService{Client: client}
	store := token.NewTokenStorage()
	tm := token.NewManager(tokenService, store, nil, appID, appSecret)
	err := tm.Initialize(ctx)
	if err != nil {
		return
	}

	fm := NewManager(client, appToken, tm)

	OTableID := "tblMm4wh9xmwRFqk"
	//KRTableID := "tblyRhpeJMvFK2PS"
	OMapping, err := fm.GetFieldMapping(ctx, OTableID)
	if err != nil {
		return
	}

	//var o1 = make(map[string]interface{}, len(OMapping))
	var OfieldMapping = make(map[string]string, len(OMapping))
	for _, x := range OMapping {
		OfieldMapping[x.FieldID] = x.FieldName
	}

	var obj v1.ObjectiveField
	//var obj KeyResult
	err = mapFields(&obj, OfieldMapping)
	if err != nil {
		fmt.Println("Error mapping fields:", err)
		return
	}

	resultJSON, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("Error marshalling result to JSON:", err)
		return
	}

	fmt.Println(string(resultJSON))
}
