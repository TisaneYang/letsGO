package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"letsgo/common/errorx"
	"letsgo/services/user/model"
	"letsgo/services/user/rpc/internal/svc"
	"letsgo/services/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update user profile
func (l *UpdateProfileLogic) UpdateProfile(in *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	// 1. Validate user ID
	if in.UserId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid user ID")
	}

	// 2. Check if user exists
	existingUser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrUserNotFound
		}
		l.Logger.Errorf("Failed to find user: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 3. Validate and prepare updates
	hasChanges := false

	// Update email if provided and different
	if in.Email != "" && in.Email != existingUser.Email {
		// Validate email format
		if !strings.Contains(in.Email, "@") || !strings.Contains(in.Email, ".") {
			return nil, errorx.NewCodeError(1001, "Invalid email format")
		}

		// Check if email already used by another user
		existingEmail, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
		if err != nil && err != model.ErrNotFound {
			l.Logger.Errorf("Failed to check email: %v", err)
			return nil, errorx.ErrDatabase
		}
		if existingEmail != nil && existingEmail.Id != in.UserId {
			return nil, errorx.NewCodeError(2006, "Email already in use")
		}

		existingUser.Email = in.Email
		hasChanges = true
	}

	// Update phone if provided and different
	if in.Phone != "" && in.Phone != existingUser.Phone {
		existingUser.Phone = in.Phone
		hasChanges = true
	}

	// Update avatar if provided and different
	if in.Avatar != "" && in.Avatar != existingUser.Avatar {
		existingUser.Avatar = in.Avatar
		hasChanges = true
	}

	// 4. If no changes, return success immediately
	if !hasChanges {
		return &user.UpdateProfileResponse{Success: true}, nil
	}

	// 5. Update timestamp
	existingUser.UpdatedAt = time.Now().Unix()

	// 6. Update database
	err = l.svcCtx.UserModel.Update(l.ctx, existingUser)
	if err != nil {
		l.Logger.Errorf("Failed to update user profile: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 7. Invalidate cache
	cacheKey := fmt.Sprintf("user:info:%d", in.UserId)
	_, err = l.svcCtx.Redis.Del(cacheKey)
	if err != nil {
		l.Logger.Errorf("Failed to invalidate cache: %v", err)
		// Continue even if cache invalidation fails
	}

	l.Logger.Infof("User profile updated successfully: user_id=%d", in.UserId)

	return &user.UpdateProfileResponse{Success: true}, nil
}
