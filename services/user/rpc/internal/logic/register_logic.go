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

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register a new user account
func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	// 1. Validate input parameters
	if err := l.validateRegisterParams(in); err != nil {
		return nil, err
	}

	// 2. Check if username already exists
	existingUser, err := l.svcCtx.UserModel.FindOneByUsername(in.Username)
	if err != nil && err != model.ErrNotFound {
		l.Logger.Errorf("Failed to check username existence: %v", err)
		return nil, errorx.ErrDatabase
	}
	if existingUser != nil {
		return nil, errorx.NewCodeError(2001, "Username already exists")
	}

	// 3. Check if email already exists
	existingEmail, err := l.svcCtx.UserModel.FindOneByEmail(in.Email)
	if err != nil && err != model.ErrNotFound {
		l.Logger.Errorf("Failed to check email existence: %v", err)
		return nil, errorx.ErrDatabase
	}
	if existingEmail != nil {
		return nil, errorx.NewCodeError(2006, "Email already exists")
	}

	// 4. Generate password salt and hash password
	salt := utils.GenerateSalt()
	hashedPassword := utils.HashPassword(in.Password, salt)

	// 5. Prepare user data
	now := time.Now().Unix()
	newUser := &model.User{
		Username:  in.Username,
		Password:  hashedPassword,
		Salt:      salt,
		Email:     in.Email,
		Phone:     in.Phone,
		Avatar:    "", // Default empty avatar
		Status:    1,  // 1 = active
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 6. Insert user into database
	_, err = l.svcCtx.UserModel.Insert(newUser)
	if err != nil {
		l.Logger.Errorf("Failed to insert user: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 7. Generate JWT token for automatic login
	token, err := l.generateToken(newUser.Id)
	if err != nil {
		l.Logger.Errorf("Failed to generate token: %v", err)
		return nil, errorx.NewCodeError(2007, "Failed to generate token")
	}

	l.Logger.Infof("User registered successfully: user_id=%d, username=%s", newUser.Id, newUser.Username)

	return &user.RegisterResponse{
		UserId: newUser.Id,
		Token:  token,
	}, nil
}

// validateRegisterParams validates registration input parameters
func (l *RegisterLogic) validateRegisterParams(in *user.RegisterRequest) error {
	// Validate username
	if len(strings.TrimSpace(in.Username)) == 0 {
		return errorx.NewCodeError(1001, "Username cannot be empty")
	}
	if len(in.Username) < 3 || len(in.Username) > 50 {
		return errorx.NewCodeError(1001, "Username must be 3-50 characters")
	}

	// Validate password
	if len(in.Password) < 6 {
		return errorx.NewCodeError(1001, "Password must be at least 6 characters")
	}

	// Validate email
	if len(strings.TrimSpace(in.Email)) == 0 {
		return errorx.NewCodeError(1001, "Email cannot be empty")
	}
	if !strings.Contains(in.Email, "@") || !strings.Contains(in.Email, ".") {
		return errorx.NewCodeError(1001, "Invalid email format")
	}

	return nil
}

// generateToken generates JWT token for user
func (l *RegisterLogic) generateToken(userId int64) (string, error) {
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
