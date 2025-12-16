package logic

import (
	"context"
	"strings"
	"time"

	"letsgo/common/errorx"
	"letsgo/common/utils"
	"letsgo/services/user/model"
	"letsgo/services/user/rpc/internal/svc"
	"letsgo/services/user/rpc/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login and get JWT token
func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	// 1. Validate input parameters
	if len(strings.TrimSpace(in.Username)) == 0 {
		return nil, errorx.NewCodeError(1001, "Username cannot be empty")
	}
	if len(strings.TrimSpace(in.Password)) == 0 {
		return nil, errorx.NewCodeError(1001, "Password cannot be empty")
	}

	// 2. Find user by username
	existingUser, err := l.svcCtx.UserModel.FindOneByUsername(in.Username)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrUserNotFound
		}
		l.Logger.Errorf("Failed to find user: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 3. Check if user is disabled
	if existingUser.Status != 1 {
		return nil, errorx.NewCodeError(2008, "User account is disabled")
	}

	// 4. Verify password
	hashedPassword := utils.HashPassword(in.Password, existingUser.Salt)
	if hashedPassword != existingUser.Password {
		return nil, errorx.ErrWrongPassword
	}

	// 5. Generate JWT token
	token, err := l.generateToken(existingUser.Id)
	if err != nil {
		l.Logger.Errorf("Failed to generate token: %v", err)
		return nil, errorx.NewCodeError(2007, "Failed to generate token")
	}

	l.Logger.Infof("User logged in successfully: user_id=%d, username=%s", existingUser.Id, existingUser.Username)

	return &user.LoginResponse{
		UserId: existingUser.Id,
		Token:  token,
	}, nil
}

// generateToken generates JWT token for user
func (l *LoginLogic) generateToken(userId int64) (string, error) {
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.AuthConf.AccessExpire
	accessSecret := l.svcCtx.Config.AuthConf.AccessSecret

	claims := make(jwt.MapClaims)
	claims["exp"] = now + accessExpire
	claims["iat"] = now
	claims["userId"] = userId

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(accessSecret))
}
