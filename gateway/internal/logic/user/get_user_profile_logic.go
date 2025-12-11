// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/user/rpc/user_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user profile - Returns current user information
func NewGetUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileLogic {
	return &GetUserProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserProfileLogic) GetUserProfile() (resp *types.UserProfileResp, err error) {
	userResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user_client.GetUserInfoRequest{
		UserId: l.ctx.Value("user_id").(int64),
	})
	if err != nil {
		return nil, err
	}

	return &types.UserProfileResp{
		UserId:    userResp.UserId,
		Username:  userResp.Username,
		Email:     userResp.Email,
		Phone:     userResp.Phone,
		Avatar:    userResp.Avatar,
		CreatedAt: userResp.CreatedAt,
	}, nil
}
