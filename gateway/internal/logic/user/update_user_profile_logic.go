// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update user profile - Updates user information
func NewUpdateUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserProfileLogic {
	return &UpdateUserProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserProfileLogic) UpdateUserProfile(req *types.UpdateProfileReq) (resp *types.UpdateProfileResp, err error) {
	// todo: add your logic here and delete this line
	userResp, err := l.svcCtx.UserRpc.UpdateProfile(l.ctx, &user.UpdateProfileRequest{
		UserId: l.ctx.Value("user_id").(int64),
		Email:  req.Email,
		Phone:  req.Phone,
		Avatar: req.Avatar,
	})

	if err != nil {
		return nil, err
	}

	return &types.UpdateProfileResp{
		Success: userResp.Success,
	}, nil
}
