package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TokenTypeAccess 访问令牌类型。
	TokenTypeAccess = "access"
	// TokenTypeRefresh 刷新令牌类型。
	TokenTypeRefresh = "refresh"
)

// Claims 自定义 JWT Claims。
type Claims struct {
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenService 统一管理 JWT 的签发与解析。
type TokenService struct {
	secret []byte
}

// NewTokenService 创建 TokenService。
func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: []byte(secret)}
}

// CreateAccessToken 签发 72 小时访问令牌。
func (s *TokenService) CreateAccessToken(userID uint) (string, *Claims, error) {
	return s.createToken(userID, TokenTypeAccess, 72*time.Hour)
}

// CreateRefreshToken 签发 7 天刷新令牌。
func (s *TokenService) CreateRefreshToken(userID uint) (string, *Claims, error) {
	return s.createToken(userID, TokenTypeRefresh, 7*24*time.Hour)
}

func (s *TokenService) createToken(userID uint, tokenType string, ttl time.Duration) (string, *Claims, error) {
	now := time.Now()
	jti, err := randomJTI()
	if err != nil {
		return "", nil, err
	}

	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "user_auth",
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", nil, err
	}
	return tokenString, claims, nil
}

// ParseToken 解析并校验 JWT。
func (s *TokenService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.UserID == 0 || claims.ID == "" || claims.TokenType == "" || claims.ExpiresAt == nil {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func randomJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
