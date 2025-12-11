package logic

import (
	"context"

	"letsgo/services/user/rpc/internal/svc"
	"letsgo/services/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Verify JWT token validity
func (l *VerifyTokenLogic) VerifyToken(in *user.VerifyTokenRequest) (*user.VerifyTokenResponse, error) {
	// todo: add your logic here and delete this line

	return &user.VerifyTokenResponse{}, nil
}
