package logic

import (
	"context"
	"strings"

	"letsgo/common/errorx"
	"letsgo/services/user/rpc/internal/svc"
	"letsgo/services/user/rpc/user"

	"github.com/golang-jwt/jwt/v4"
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
	// 1. Validate input
	if len(strings.TrimSpace(in.Token)) == 0 {
		return &user.VerifyTokenResponse{
			Valid:  false,
			UserId: 0,
		}, nil
	}

	// 2. Parse and validate JWT token
	token, err := jwt.Parse(in.Token, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errorx.ErrTokenInvalid
		}
		return []byte(l.svcCtx.Config.AuthConf.AccessSecret), nil
	})

	// 3. Check for parsing errors
	if err != nil {
		l.Logger.Errorf("Failed to parse token: %v", err)
		return &user.VerifyTokenResponse{
			Valid:  false,
			UserId: 0,
		}, nil
	}

	// 4. Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract user ID
		userId, ok := claims["userId"].(float64)
		if !ok {
			l.Logger.Error("Invalid userId in token claims")
			return &user.VerifyTokenResponse{
				Valid:  false,
				UserId: 0,
			}, nil
		}

		// Extract expiration time
		exp, ok := claims["exp"].(float64)
		if !ok {
			l.Logger.Error("Invalid exp in token claims")
			return &user.VerifyTokenResponse{
				Valid:  false,
				UserId: 0,
			}, nil
		}

		l.Logger.Infof("Token verified successfully: user_id=%d", int64(userId))

		return &user.VerifyTokenResponse{
			Valid:    true,
			UserId:   int64(userId),
			ExpireAt: int64(exp),
		}, nil
	}

	// 5. Token is invalid
	return &user.VerifyTokenResponse{
		Valid:  false,
		UserId: 0,
	}, nil
}
