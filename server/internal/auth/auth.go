// Package auth 负责 JWT 签发、解析和 Gin 中间件鉴权。后台端和眼镜端共用同一套用户模型，
// 但通过不同 scope 隔离权限；眼镜端 token 还携带设备 ID，便于集中拦截已撤销或丢失设备。
package auth

import (
	"net/http"
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Scope string

const (
	ScopeAdmin   Scope = "admin"
	ScopeGlasses Scope = "glasses"
)

type Claims struct {
	UserID   uint64  `json:"user_id"`
	DeviceID *uint64 `json:"device_id,omitempty"`
	Scope    Scope   `json:"scope"`
	JWTID    string  `json:"jti"`
	jwt.RegisteredClaims
}

type Service struct {
	db        *gorm.DB
	secret    []byte
	accessTTL time.Duration
}

// NewService 创建认证服务，注入数据库、JWT 密钥和访问令牌有效期。
func NewService(db *gorm.DB, secret string, accessTTL time.Duration) *Service {
	return &Service{db: db, secret: []byte(secret), accessTTL: accessTTL}
}

// Login 根据用户名和登录范围签发访问令牌，眼镜端登录必须携带设备 ID。
func (s *Service) Login(username string, scope Scope, deviceID *uint64) (string, database.User, error) {
	var user database.User
	if err := s.db.Where("username = ? AND status = ?", username, "active").First(&user).Error; err != nil {
		return "", user, httperr.New(httperr.AuthForbidden, "user is not allowed")
	}
	if scope == ScopeGlasses && deviceID == nil {
		return "", user, httperr.New(httperr.DeviceRevoked, "device is required")
	}
	token, err := s.IssueAccessToken(user.ID, deviceID, scope)
	return token, user, err
}

// OrganizationName 查询指定组织编码对应的单位名称。
func (s *Service) OrganizationName(orgCode string) (string, error) {
	if orgCode == "" {
		return "", nil
	}
	var org database.Organization
	if err := s.db.Select("name").Where("code = ?", orgCode).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return org.Name, nil
}

// IssueAccessToken 为指定用户、设备和 scope 生成 JWT 访问令牌。
func (s *Service) IssueAccessToken(userID uint64, deviceID *uint64, scope Scope) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:   userID,
		DeviceID: deviceID,
		Scope:    scope,
		JWTID:    uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// Parse 校验 JWT 并解析业务 Claims，同时检查眼镜设备是否仍然有效。
func (s *Service) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, httperr.New(httperr.AuthTokenExpired, "token expired or invalid")
	}
	if claims.DeviceID != nil {
		var device database.Device
		if err := s.db.First(&device, *claims.DeviceID).Error; err != nil {
			return nil, httperr.New(httperr.DeviceRevoked, "device no longer exists")
		}
		switch device.Status {
		case "revoked":
			return nil, httperr.New(httperr.DeviceRevoked, "device revoked")
		case "lost_disabled":
			return nil, httperr.New(httperr.DeviceDisabledLost, "device disabled as lost")
		}
	}
	return claims, nil
}

// Middleware 构造 Gin 鉴权中间件，校验 Bearer Token 和接口所需 scope。
func Middleware(service *Service, scope Scope) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httperr.Respond(c, httperr.New(httperr.AuthTokenExpired, "missing bearer token"))
			c.Abort()
			return
		}
		claims, err := service.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httperr.Respond(c, err)
			c.Abort()
			return
		}
		if claims.Scope != scope {
			httperr.Respond(c, httperr.New(httperr.AuthForbidden, "token scope is not allowed"))
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		if claims.DeviceID != nil {
			c.Set("device_id", *claims.DeviceID)
		}
		c.Next()
	}
}

// UserID 从 Gin 上下文中读取已认证用户 ID，缺失表示中间件未正确执行。
func UserID(c *gin.Context) uint64 {
	value, exists := c.Get("user_id")
	if !exists {
		panic(http.ErrNoCookie)
	}
	return value.(uint64)
}

// DeviceID 从 Gin 上下文中读取眼镜设备 ID，后台请求通常为空。
func DeviceID(c *gin.Context) *uint64 {
	value, exists := c.Get("device_id")
	if !exists {
		return nil
	}
	id := value.(uint64)
	return &id
}
