// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealTicker(t *testing.T) {
	// 定义测试的持续时间和间隔
	d := 100 * time.Millisecond
	duration := 500 * time.Millisecond

	// 初始化 realClock 和 Ticker
	clock := realClock{}
	ticker := clock.NewTicker(d)
	defer ticker.Stop() // 确保 ticker 在测试后停止

	// 创建一个计时器来限制测试的最长运行时间
	timeout := time.After(duration)
	// 用于接收 tick 事件，验证是否按预期间隔发生
	tickC := make(chan time.Time)

	go func() {
		for tm := range ticker.C() {
			tickC <- tm
		}
	}()

	// 记录接收到的 tick 数量
	ticksReceived := 0
	for {
		select {
		case <-timeout:
			if ticksReceived < 3 { // 根据持续时间和间隔，预期至少收到3个tick
				t.Errorf("Expected at least 3 ticks, got %d", ticksReceived)
			}
			return
		case tm := <-tickC:
			ticksReceived++
			t.Logf("Tick at %v", tm)
		}
	}
}

func TestRealTickerWithTestify(t *testing.T) {
	// 定义测试的持续时间和间隔
	d := 100 * time.Millisecond
	duration := 500 * time.Millisecond

	// 初始化 realClock 和 Ticker
	clock := realClock{}
	ticker := clock.NewTicker(d)
	defer ticker.Stop() // 确保 ticker 在测试后停止

	// 创建一个计时器来限制测试的最长运行时间
	timeout := time.After(duration)
	// 用于接收 tick 事件，验证是否按预期间隔发生
	tickC := make(chan time.Time)

	go func() {
		for tm := range ticker.C() {
			tickC <- tm
		}
	}()

	// 记录接收到的 tick 数量
	ticksReceived := 0
	for {
		select {
		case <-timeout:
			require.True(t, ticksReceived >= 3, "Should receive at least 3 ticks")
			return
		case tm := <-tickC:
			ticksReceived++
			assert.NotZero(t, tm, "The tick timestamp should not be zero")
		}
	}
}
