package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) GetCount(ctx context.Context, token string) (int, error) {
	args := m.Called(ctx, token)
	return args.Int(0), args.Error(1)
}

func (m *MockDatabase) IncrementCount(ctx context.Context, token string, duration time.Duration) error {
	return m.Called(ctx, token, duration).Error(0)
}

func (m *MockDatabase) UpdateExpiration(ctx context.Context, token string, duration time.Duration) error {
	return m.Called(ctx, token, duration).Error(0)
}

func TestAddTokenConf(t *testing.T) {
	db := &MockDatabase{}
	rl := NewRateLimiter(db)

	rl.AddTokenConf("apiToken", 100, 10*time.Second)

	assert.Equal(t, 100, rl.Conf["apiToken"].Limit)
	assert.Equal(t, 10*time.Second, rl.Conf["apiToken"].Interval)
}

func TestCheckRateLimitCount_BelowLimit(t *testing.T) {
	db := &MockDatabase{}
	rl := NewRateLimiter(db)
	rl.AddTokenConf("apiToken", 5, 10*time.Second)

	db.On("GetCount", mock.Anything, "apiToken").Return(3, nil)

	ok, err := rl.CheckRateLimitCount(context.Background(), "apiToken", false)

	assert.NoError(t, err)
	assert.True(t, ok)
	db.AssertExpectations(t)
}

func TestCheckRateLimitCount_AboveLimit(t *testing.T) {
	db := &MockDatabase{}
	rl := NewRateLimiter(db)
	rl.AddTokenConf("apiToken", 5, 10*time.Second)

	db.On("GetCount", mock.Anything, "apiToken").Return(5, nil)

	ok, err := rl.CheckRateLimitCount(context.Background(), "apiToken", false)

	assert.NoError(t, err)
	assert.False(t, ok)
	db.AssertExpectations(t)
}

func TestIncreaseRateLimitCount(t *testing.T) {
	db := &MockDatabase{}
	rl := NewRateLimiter(db)
	rl.AddTokenConf("apiToken", 5, 10*time.Second)

	db.On("IncrementCount", mock.Anything, "apiToken", 10*time.Second).Return(nil)

	err := rl.IncreaseRateLimitCount(context.Background(), "apiToken", false)
	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestUpdateExpiration(t *testing.T) {
	db := &MockDatabase{}
	rl := NewRateLimiter(db)
	rl.AddTokenConf("apiToken", 5, 10*time.Second)

	db.On("UpdateExpiration", mock.Anything, "apiToken", 10*time.Second).Return(nil)

	err := rl.UpdateExpiration(context.Background(), "apiToken", false)
	assert.NoError(t, err)
	db.AssertExpectations(t)
}
