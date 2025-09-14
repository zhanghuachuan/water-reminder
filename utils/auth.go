package utils

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/types"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 使用 utils.GetDB() 替代直接使用 DB 变量

var jwtSecret = []byte("your_jwt_secret_key")

type User = types.User

// CreateUser 创建新用户
func CreateUser(username, email, password string) (*User, error) {
	if email == "" || password == "" || username == "" {
		return nil, errors.New("all fields are required")
	}

	// 检查邮箱是否已存在
	var existingUser User
	if err := database.GetDB().Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:       uuid.New().String(),
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
	}

	// 保存到数据库
	if err := database.GetDB().Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser 从数据库验证用户凭据
func AuthenticateUser(email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	var user User
	if err := database.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 返回用户信息（不包含密码）
	return &User{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}, nil
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		userID, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 获取完整用户信息
		var user types.User
		if err := database.GetDB().Where("id = ?", userID).First(&user).Error; err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// 将用户信息和ID存入上下文
		ctx := context.WithValue(r.Context(), "user", &user)
		ctx = context.WithValue(ctx, "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateJWT 验证JWT令牌并返回详细的错误信息
// 1. 验证JWT签名有效性
func ValidateJWT(tokenString string) (string, error) {
	// 1. 验证JWT基本格式和签名
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return "", errors.New("malformed token")
			case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return "", errors.New("invalid token signature")
			default:
				return "", errors.New("invalid token")
			}
		}
		return "", errors.New("could not parse token")
	}

	// 2. 验证claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", errors.New("invalid user id in token")
	}
	return userID, nil
}
