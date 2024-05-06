// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package token

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/imxw/miniokr/internal/pkg/log"
)

const TokenFilePath = "tenant_access_token.json"

type Provider interface {
	EnsureValidToken(ctx context.Context) (string, error)
}

type Service interface {
	GetNewToken(ctx context.Context, appID string, appSecret string) (*Response, error)
}

type Storage interface {
	LoadTokenFromFile(path string) (Data, error)
	SaveTokenToFile(data Data, path string) error
}

type Manager struct {
	Service                Service
	AppID                  string
	AppSecret              string
	rwMutex                sync.RWMutex
	cachedToken            string
	expiryTime             time.Time
	Storage                Storage
	Clock                  Clock
	TokenRefresherInterval time.Duration
}

type Data struct {
	TenantAccessToken string    `json:"tenant_access_token"`
	Expire            int       `json:"expire"`
	SavedAt           time.Time `json:"saved_at"`
}

type Response struct {
	Code int    `json:"code"` // API 响应码
	Msg  string `json:"msg"`  // API 响应消息
	Data
}

func NewManager(service Service, storage Storage, clock Clock, appID, appSecret string) *Manager {
	m := &Manager{
		Service:                service,
		AppID:                  appID,
		AppSecret:              appSecret,
		cachedToken:            "",
		expiryTime:             time.Time{},
		Storage:                storage,
		TokenRefresherInterval: 1 * time.Hour,
		Clock:                  clock,
	}
	return m
}

func (m *Manager) Initialize(ctx context.Context) error {
	return m.initializeTokenCache(ctx)
}

func (m *Manager) initializeTokenCache(ctx context.Context) error {
	tokenData, err := m.Storage.LoadTokenFromFile(TokenFilePath)

	// 如果文件读取失败或token已过期，尝试刷新token
	if err != nil || tokenData.SavedAt.IsZero() || time.Now().After(tokenData.SavedAt.Add(time.Duration(tokenData.Expire)*time.Second)) {
		if err != nil {
			log.Errorw("Error loading token from file, attempting to refresh", "err", err)
		} else {
			log.Errorw("Loaded token is expired or has invalid SavedAt, attempting to refresh", "SavedAt", tokenData.SavedAt)
		}
		_, refreshErr := m.refreshToken(ctx)
		if refreshErr != nil {
			log.Errorw("Failed to refresh token", "error", refreshErr)
			return refreshErr
		}
		log.Infow("Token refreshed successfully.")
		return nil
	}

	// 文件存在且令牌有效，设置缓存
	m.cachedToken = tokenData.TenantAccessToken
	m.expiryTime = tokenData.SavedAt.Add(time.Duration(tokenData.Expire) * time.Second)
	log.Infow("Token loaded and cache initialized successfully.")
	return nil
}

func (m *Manager) EnsureValidToken(ctx context.Context) (string, error) {
	m.rwMutex.RLock()
	// 检查令牌是否有效
	if time.Now().Before(m.expiryTime) {
		token := m.cachedToken
		m.rwMutex.RUnlock()
		return token, nil
	}
	m.rwMutex.RUnlock() // 释放读锁，因为可能需要写操作

	m.rwMutex.Lock()         // 获取写锁进行令牌刷新
	defer m.rwMutex.Unlock() // 确保在退出方法前释放写锁

	// 再次检查，因为其他线程可能已经更新了令牌
	if time.Now().Before(m.expiryTime) {
		return m.cachedToken, nil
	}

	// 从文件中重新加载，以处理系统重启或令牌文件外部更新的情况
	tokenData, err := m.Storage.LoadTokenFromFile(TokenFilePath)
	if err != nil || time.Now().After(tokenData.SavedAt.Add(time.Duration(tokenData.Expire)*time.Second)) {
		// 文件加载失败或令牌过期，尝试刷新令牌
		return m.refreshToken(ctx)
	}

	// 更新内存中的缓存
	m.cachedToken = tokenData.TenantAccessToken
	m.expiryTime = tokenData.SavedAt.Add(time.Duration(tokenData.Expire) * time.Second)

	return m.cachedToken, nil
}

func (m *Manager) refreshToken(ctx context.Context) (string, error) {
	// 从Token服务获取新的令牌响应
	resp, err := m.Service.GetNewToken(ctx, m.AppID, m.AppSecret)
	if err != nil {
		return "", err
	}

	// 从响应中提取令牌数据并更新内存中的缓存
	tokenData := Data{
		TenantAccessToken: resp.Data.TenantAccessToken,
		Expire:            resp.Data.Expire, // 直接保存剩余秒数
		SavedAt:           time.Now(),
	}
	m.cachedToken = tokenData.TenantAccessToken
	m.expiryTime = tokenData.SavedAt.Add(time.Duration(tokenData.Expire) * time.Second)

	// 将新令牌保存到文件
	if err := m.Storage.SaveTokenToFile(tokenData, TokenFilePath); err != nil {
		return "", fmt.Errorf("保存Token到文件失败: %v", err)
	}

	return tokenData.TenantAccessToken, nil
}

func (m *Manager) StartTokenRefresher(ctx context.Context) {
	tick := m.Clock.NewTicker(m.TokenRefresherInterval) // 使用 Clock 接口创建 Ticker
	// tick := time.NewTicker(m.TokenRefresherInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infow("Context cancelled, stopping refresher", "now", time.Now())
			return
		case t := <-tick.C():
			log.Infow("Tick at ", "tick", t)
			_, err := m.EnsureValidToken(ctx)
			if err != nil {
				log.Infow("Error refreshing token", "error", err)
				// 可以选择重试或发送告警
			}
		}
	}
}
