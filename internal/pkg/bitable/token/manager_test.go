// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package token

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService Mock Service
type MockService struct {
	mock.Mock
}

func (m *MockService) GetNewToken(ctx context.Context, appID, appSecret string) (*Response, error) {
	args := m.Called(ctx, appID, appSecret)
	return args.Get(0).(*Response), args.Error(1)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) LoadTokenFromFile(path string) (Data, error) {
	args := m.Called(path)
	return args.Get(0).(Data), args.Error(1)
}

func (m *MockStorage) SaveTokenToFile(data Data, path string) error {
	args := m.Called(data, path)
	return args.Error(0)
}

// MockTicker 模拟一个定时器
type MockTicker struct {
	mock.Mock
	CChan chan time.Time
}

func (m *MockTicker) C() <-chan time.Time {
	return m.CChan
}

func (m *MockTicker) Stop() {
	m.Called()
	close(m.CChan)
}

// MockClock 模拟一个时钟
type MockClock struct {
	mock.Mock
}

func (m *MockClock) NewTicker(d time.Duration) Ticker {
	args := m.Called(d)
	return args.Get(0).(Ticker)
}

// TestInitializeTokenCache_LoadError 文件加载失败，触发刷新令牌
func TestInitializeTokenCache_LoadError(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	mockService := new(MockService)
	manager := NewManager(mockService, mockStorage, nil, "appID", "appSecret")

	// 模拟文件读取错误
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(Data{}, errors.New("read error")).Once()

	// 模拟成功的token刷新
	refreshedTokenData := Data{TenantAccessToken: "new_token", Expire: 3600, SavedAt: time.Now()}
	mockService.On("GetNewToken", ctx, "appID", "appSecret").Return(&Response{Data: refreshedTokenData}, nil).Once()

	// 模拟保存新token的行为
	mockStorage.On("SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath).Return(nil).Once()

	// 执行初始化来触发token刷新逻辑
	err := manager.Initialize(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "new_token", manager.cachedToken)
	// 验证expiryTime是否被正确设置
	expiryTime := refreshedTokenData.SavedAt.Add(time.Duration(refreshedTokenData.Expire) * time.Second)
	assert.WithinDuration(t, expiryTime, manager.expiryTime, time.Second)
	// 确认模拟行为被调用
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

// TestInitializeTokenCache_ExpiredToken 令牌文件存在但已过期，触发更新
func TestInitializeTokenCache_ExpiredToken(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	mockService := new(MockService)
	appID := "test_app_id"
	appSecret := "test_app_secret"
	manager := NewManager(mockService, mockStorage, nil, appID, appSecret)

	currentTime := time.Now() // 使用一个固定的当前时间

	// 模拟已过期的token数据
	expiredSavedAt := currentTime.Add(-2 * time.Hour) // 2小时前保存的token
	expiredTokenData := Data{TenantAccessToken: "expired_token", Expire: 3600, SavedAt: expiredSavedAt}
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(expiredTokenData, nil)

	// 模拟成功的token刷新
	newTokenData := Data{TenantAccessToken: "new_token", Expire: 3600, SavedAt: currentTime}
	mockService.On("GetNewToken", mock.Anything, appID, appSecret).Return(&Response{Data: newTokenData}, nil)

	// 模拟保存新token的行为，因为 refreshToken 会尝试保存新的令牌
	mockStorage.On("SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath).Return(nil).Once()

	// 执行初始化来触发token刷新逻辑
	err := manager.Initialize(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "new_token", manager.cachedToken)
	assert.WithinDuration(t, currentTime.Add(time.Second*time.Duration(newTokenData.Expire)), manager.expiryTime, time.Second)

	// 确认模拟行为被调用
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

// TestInitializeTokenCache_TokenValidNoRefreshNeeded 令牌文件且有效，不需要刷新
func TestInitializeTokenCache_TokenValidNoRefreshNeeded(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	mockService := new(MockService)
	appID := "test_app_id"
	appSecret := "test_app_secret"
	manager := NewManager(mockService, mockStorage, nil, appID, appSecret)

	// 设置一个未来的时间以模拟有效的token
	validSavedAt := time.Now().Add(-30 * time.Minute) // 30分钟前保存的token
	validExpiry := 2 * time.Hour                      // 有效期为2小时
	validTokenData := Data{
		TenantAccessToken: "valid_token",
		Expire:            int(validExpiry.Seconds()), // 转换为秒
		SavedAt:           validSavedAt,
	}

	// 模拟文件存在且token有效
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(validTokenData, nil).Once()

	// 执行初始化，应该不会触发刷新逻辑
	err := manager.Initialize(ctx)
	assert.NoError(t, err)

	// 确保cachedToken是有效的，且没有尝试刷新token
	assert.Equal(t, "valid_token", manager.cachedToken)
	assert.WithinDuration(t, validSavedAt.Add(validExpiry), manager.expiryTime, time.Second)

	// GetNewToken 和 SaveTokenToFile 不应该被调用
	mockService.AssertNotCalled(t, "GetNewToken", mock.Anything, appID, appSecret)
	mockStorage.AssertNotCalled(t, "SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath)

	// 确认模拟行为被调用
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

// 刷新失败
func TestInitializeTokenCache_LoadError_RefreshError(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	mockService := new(MockService)
	manager := NewManager(mockService, mockStorage, nil, "appID", "appSecret")

	// 模拟文件读取错误
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(Data{}, errors.New("read error")).Once()

	// 模拟失败的token刷新
	mockService.On("GetNewToken", ctx, "appID", "appSecret").Return((*Response)(nil), errors.New("token刷新失败")).Once()

	// 执行初始化来触发token刷新逻辑
	err := manager.Initialize(ctx)
	assert.Error(t, err)
	// 确认模拟行为被调用
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

// 保存失败
func TestInitializeTokenCache_LoadError_SaveError(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	mockService := new(MockService)
	manager := NewManager(mockService, mockStorage, nil, "appID", "appSecret")

	// 模拟文件读取错误
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(Data{}, errors.New("read error")).Once()

	// 模拟成功的token刷新
	refreshedTokenData := Data{TenantAccessToken: "new_token", Expire: 3600, SavedAt: time.Now()}
	mockService.On("GetNewToken", ctx, "appID", "appSecret").Return(&Response{Data: refreshedTokenData}, nil).Once()

	// 模拟保存新token的行为
	mockStorage.On("SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath).Return(errors.New("保存token到文件失败")).Once()

	// 执行初始化来触发token刷新逻辑
	err := manager.Initialize(ctx)
	assert.Error(t, err)
	// 确认模拟行为被调用
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

/*
令牌有效：令牌在缓存中且未过期，应返回现有令牌。
令牌过期，重载成功：缓存中的令牌已过期，从文件加载新令牌成功，且令牌有效。
文件令牌过期，刷新成功：文件中的令牌也过期，需要调用刷新逻辑，刷新后令牌有效。
文件加载失败：尝试从文件加载令牌时发生错误，需要处理错误。
令牌刷新失败：令牌需要刷新，但刷新逻辑失败。
*/

// TestEnsureValidToken_CachedValid 测试缓存中的令牌有效的情况
func TestEnsureValidToken_CachedValid(t *testing.T) {
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	manager := NewManager(nil, mockStorage, mockClock, "appID", "appSecret")
	now := time.Now()
	manager.cachedToken = "valid_token"
	manager.expiryTime = now.Add(time.Hour) // 令牌在一小时后过期

	token, err := manager.EnsureValidToken(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "valid_token", token)
	mockStorage.AssertNotCalled(t, "LoadTokenFromFile")
}

func TestEnsureValidToken_TokenExpired_ReloadFromFileSuccess(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockService)
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	appID := "test_app_id"
	appSecret := "test_app_secret"

	manager := NewManager(mockService, mockStorage, mockClock, appID, appSecret)

	// 设置缓存中的令牌为过期令牌
	manager.cachedToken = "expired_token"
	manager.expiryTime = time.Now().Add(-1 * time.Hour) // 设置为1小时前过期

	// 模拟从文件加载新的有效令牌
	validTokenData := Data{
		TenantAccessToken: "new_valid_token",
		Expire:            3600, // 新令牌的有效期1小时
		SavedAt:           time.Now(),
	}
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(validTokenData, nil).Once()

	// 调用EnsureValidToken并验证返回的新令牌
	actualToken, err := manager.EnsureValidToken(ctx)

	// 验证结果
	assert.NoError(t, err, "应该没有错误发生")
	assert.Equal(t, validTokenData.TenantAccessToken, actualToken, "应该返回新加载的有效令牌")
	newExpiryTime := validTokenData.SavedAt.Add(time.Duration(validTokenData.Expire) * time.Second)
	assert.WithinDuration(t, newExpiryTime, manager.expiryTime, time.Second, "新令牌的过期时间应该在未来")

	// 确认模拟行为被调用
	mockStorage.AssertCalled(t, "LoadTokenFromFile", TokenFilePath)

	// 确认所有预期的模拟行为都已经满足
	mockStorage.AssertExpectations(t)
}

func TestEnsureValidToken_TokenExpired_ReloadSuccess(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockService)
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	appID := "test_app_id"
	appSecret := "test_app_secret"

	currentTime := time.Now()
	manager := NewManager(mockService, mockStorage, mockClock, appID, appSecret)

	// 设置已过期的令牌
	expiredTokenData := Data{
		TenantAccessToken: "expired_token",
		Expire:            3600,                            // 令牌有效期1小时
		SavedAt:           currentTime.Add(-2 * time.Hour), // 设置为2小时前保存的令牌
	}
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(expiredTokenData, nil).Once()

	// 模拟刷新令牌成功
	newTokenData := Data{
		TenantAccessToken: "new_valid_token",
		Expire:            3600, // 新令牌的有效期1小时
		SavedAt:           currentTime,
	}
	mockService.On("GetNewToken", ctx, appID, appSecret).Return(&Response{Data: newTokenData}, nil).Once()

	// 模拟保存新令牌到文件
	mockStorage.On("SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath).Return(nil).Once()

	// 执行 EnsureValidToken
	actualToken, err := manager.EnsureValidToken(ctx)

	// 验证结果
	assert.NoError(t, err, "应该没有错误发生")
	assert.Equal(t, newTokenData.TenantAccessToken, actualToken, "应该返回新加载的有效令牌")
	newExpiryTime := newTokenData.SavedAt.Add(time.Second * time.Duration(newTokenData.Expire))
	assert.WithinDuration(t, newExpiryTime, manager.expiryTime, time.Second, "新令牌的过期时间应该在未来")

	// 确认 LoadTokenFromFile 和 SaveTokenToFile 被调用
	mockStorage.AssertCalled(t, "LoadTokenFromFile", TokenFilePath)
	mockStorage.AssertCalled(t, "SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath)
	mockService.AssertCalled(t, "GetNewToken", ctx, appID, appSecret)

	// 确认模拟行为被调用
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestEnsureValidToken_LoadTokenFromFileError(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockService)
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	appID := "test_app_id"
	appSecret := "test_app_secret"

	manager := NewManager(mockService, mockStorage, mockClock, appID, appSecret)

	// 模拟从文件加载令牌时发生错误
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(Data{}, errors.New("file read error")).Once()

	// 因为是文件加载错误，我们不期望调用 GetNewToken 方法
	// 所以不需要模拟 Service 的 GetNewToken 方法

	// 调用 EnsureValidToken 并验证返回的错误
	actualToken, err := manager.EnsureValidToken(ctx)

	// 验证结果
	assert.Error(t, err, "应该返回一个错误，因为文件加载失败")
	assert.Empty(t, actualToken, "在文件加载失败的情况下，应该没有返回令牌")

	// 确认模拟行为被调用
	mockStorage.AssertCalled(t, "LoadTokenFromFile", TokenFilePath)
	mockService.AssertNotCalled(t, "GetNewToken", mock.Anything, mock.Anything, mock.Anything)

	// 确认所有预期的模拟行为都已经满足
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestEnsureValidToken_RefreshTokenFails(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockService)
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	appID := "test_app_id"
	appSecret := "test_app_secret"

	manager := NewManager(mockService, mockStorage, mockClock, appID, appSecret)

	// 创建已过期的令牌数据
	expiredTokenData := Data{
		TenantAccessToken: "expired_token",
		Expire:            3600,                           // 令牌有效期1小时
		SavedAt:           time.Now().Add(-2 * time.Hour), // 设置为2小时前保存的令牌，已过期
	}

	// 模拟从文件加载操作，返回已过期的令牌数据
	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(expiredTokenData, nil).Once()

	// 模拟刷新令牌失败
	mockService.On("GetNewToken", mock.Anything, appID, appSecret).Return((*Response)(nil), errors.New("refresh failed")).Once()

	// 调用 EnsureValidToken 并验证返回的错误
	actualToken, err := manager.EnsureValidToken(ctx)

	// 验证结果
	assert.Error(t, err, "应该返回一个错误，因为刷新令牌失败")
	assert.Empty(t, actualToken, "在刷新令牌失败的情况下，应该没有返回令牌")

	// 确认模拟行为被调用
	mockStorage.AssertCalled(t, "LoadTokenFromFile", TokenFilePath)
	mockService.AssertCalled(t, "GetNewToken", mock.Anything, appID, appSecret)

	// 确认所有预期的模拟行为都已经满足
	mockStorage.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestStartTokenRefresher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockService := new(MockService)
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	mockTicker := new(MockTicker)
	mockTicker.CChan = make(chan time.Time, 1) // 创建带缓冲的通道

	manager := NewManager(mockService, mockStorage, mockClock, "appID", "appSecret")
	manager.TokenRefresherInterval = 100 * time.Millisecond

	// 设定 Mock 对象的预期行为
	mockClock.On("NewTicker", mock.AnythingOfType("time.Duration")).Return(mockTicker)
	mockTicker.On("Stop").Once()

	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(Data{}, nil) // 假设总是成功加载
	mockStorage.On("SaveTokenToFile", mock.AnythingOfType("Data"), TokenFilePath).Return(nil).Once()
	mockService.On("GetNewToken", mock.Anything, "appID", "appSecret").Return(&Response{
		Data: Data{
			TenantAccessToken: "new_valid_token",
			Expire:            3600,
		},
	}, nil).Once() // 假设获取新令牌总是成功

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered in f %v", r)
			}
		}()
		manager.StartTokenRefresher(ctx)
	}()

	// 触发定时器
	mockTicker.CChan <- time.Now()

	// 给足够时间处理通道消息
	time.Sleep(time.Millisecond * 100)

	// 触发停止，观察是否有阻塞
	cancel()
	time.Sleep(time.Millisecond * 100) // 确保所有goroutine已经停止

	// 验证所有预期调用都已满足
	mockService.AssertExpectations(t)
	mockTicker.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestStartTokenRefresher_faild(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockService := new(MockService)
	mockStorage := new(MockStorage)
	mockClock := new(MockClock)
	mockTicker := new(MockTicker)
	mockTicker.CChan = make(chan time.Time, 1) // 创建带缓冲的通道

	manager := NewManager(mockService, mockStorage, mockClock, "appID", "appSecret")
	manager.TokenRefresherInterval = 100 * time.Millisecond

	// 设定 Mock 对象的预期行为
	mockClock.On("NewTicker", mock.AnythingOfType("time.Duration")).Return(mockTicker)
	mockTicker.On("Stop").Once()

	mockStorage.On("LoadTokenFromFile", TokenFilePath).Return(Data{}, nil)                                                      // 假设总是成功加载
	mockService.On("GetNewToken", mock.Anything, "appID", "appSecret").Return((*Response)(nil), errors.New("获取token失败")).Once() // 获取新令牌失败

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered in f %v", r)
			}
		}()
		manager.StartTokenRefresher(ctx)
	}()

	// 触发定时器
	mockTicker.CChan <- time.Now()

	// 给足够时间处理通道消息
	time.Sleep(time.Millisecond * 100)

	// 触发停止，观察是否有阻塞
	cancel()
	time.Sleep(time.Millisecond * 100) // 确保所有goroutine已经停止

	// 验证所有预期调用都已满足
	mockService.AssertExpectations(t)
	mockTicker.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
