// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package field

import (
	"encoding/json"
	"os"
)

func SaveFieldMappingToFile(mappingData MappingData, filePath string) error {
	data, err := json.Marshal(mappingData)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func LoadFieldMappingFromFile(filePath string) (MappingData, error) {
	var mappingData MappingData
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return mappingData, nil // 文件不存在，返回空FieldMappingData
		}
		return mappingData, err
	}
	err = json.Unmarshal(data, &mappingData)
	return mappingData, err
}
