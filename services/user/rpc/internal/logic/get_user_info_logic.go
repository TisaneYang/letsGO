package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/user/model"
	"letsgo/services/user/rpc/internal/svc"
	"letsgo/services/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get user information by user ID
func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoRequest) (*user.GetUserInfoResponse, error) {
	// 1. Validate user ID
	if in.UserId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid user ID")
	}

	// 2. Try to get from Redis cache first
	cacheKey := fmt.Sprintf("user:info:%d", in.UserId)
	cachedData, err := l.svcCtx.Redis.Get(cacheKey)
	if err == nil && cachedData != "" {
		// Cache hit - unmarshal and return
		var cachedUser user.GetUserInfoResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedUser); err == nil {
			l.Logger.Infof("User info retrieved from cache: user_id=%d", in.UserId)
			return &cachedUser, nil
		}
	}

	// 3. Cache miss - query from database
	existingUser, err := l.svcCtx.UserModel.FindOne(in.UserId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrUserNotFound
		}
		l.Logger.Errorf("Failed to find user: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 4. Prepare response
	response := &user.GetUserInfoResponse{
		UserId:    existingUser.Id,
		Username:  existingUser.Username,
		Email:     existingUser.Email,
		Phone:     existingUser.Phone,
		Avatar:    existingUser.Avatar,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
	}

	// 5. Cache the result in Redis (TTL: 1 hour)
	jsonData, err := json.Marshal(response)
	if err == nil {
		err = l.svcCtx.Redis.Setex(cacheKey, string(jsonData), 3600)
		if err != nil {
			l.Logger.Errorf("Failed to cache user info: %v", err)
			// Continue even if caching fails
		}
	}

	l.Logger.Infof("User info retrieved from database: user_id=%d", in.UserId)

	return response, nil
}
