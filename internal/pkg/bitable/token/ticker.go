// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package token

import "time"

type Clock interface {
	NewTicker(d time.Duration) Ticker
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// realTicker 包装 time.Ticker，使其满足 Ticker 接口
type realTicker struct {
	*time.Ticker
}

// C 提供对 time.Ticker 的 C 字段的访问，符合自定义 Ticker 接口
func (rt *realTicker) C() <-chan time.Time {
	return rt.Ticker.C
}

// Stop 封装对 time.Ticker 的 Stop 方法调用
func (rt *realTicker) Stop() {
	rt.Ticker.Stop()
}

// realClock 代表实际的时钟，使用标准库的 time.Ticker
type realClock struct{}

// NewTicker 创建一个新的 Ticker，返回包装过的 realTicker
func (rc *realClock) NewTicker(d time.Duration) Ticker {
	return &realTicker{Ticker: time.NewTicker(d)}
}

// NewRealClock 创建一个新的 realClock 实例
func NewRealClock() Clock {
	return &realClock{}
}
