// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package token

import (
	"encoding/json"
	"fmt"
	"os"
)

func NewTokenStorage() Storage {
	return &fileStorage{}
}

type fileStorage struct{}

func (f *fileStorage) LoadTokenFromFile(path string) (Data, error) {
	return loadTokenFromFile(path)
}

func (f *fileStorage) SaveTokenToFile(data Data, path string) error {
	return saveTokenToFile(data, path)
}

func saveTokenToFile(data Data, filePath string) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, d, 0644)
}

func loadTokenFromFile(filePath string) (Data, error) {
	var data Data
	d, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, os.ErrNotExist
		}
		return data, err
	}
	err = json.Unmarshal(d, &data)
	if err != nil {
		return data, fmt.Errorf("failed to unmarshal token data: %v", err)
	}
	return data, nil
}
