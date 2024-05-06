// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package main

import (
	"os"

	_ "go.uber.org/automaxprocs"

	"github.com/imxw/miniokr/internal/miniokr"
)

// Go 程序的默认入口函数(主函数).
func main() {
	command := miniokr.NewMiniOkrCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
