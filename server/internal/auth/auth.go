// Package auth 负责 JWT 签发、解析和 Gin 中间件鉴权。后台端和眼镜端共用同一套用户模型，
// 但通过不同 scope 隔离权限；眼镜端 token 还携带设备 ID，便于集中拦截已撤销或丢失设备。
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	db         *gorm.DB
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewService 创建认证服务，注入数据库、JWT 密钥和访问令牌有效期。
func NewService(db *gorm.DB, secret string, accessTTL time.Duration) *Service {
	return NewServiceWithRefresh(db, secret, accessTTL, 30*24*time.Hour)
}

// NewServiceWithRefresh 创建支持 refresh token 的认证服务。
func NewServiceWithRefresh(db *gorm.DB, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// Login 根据用户名、密码和登录范围签发访问令牌，眼镜端登录必须携带设备 ID。
func (s *Service) Login(username, password string, scope Scope, deviceID *uint64) (string, database.User, error) {
	var user database.User
	if err := s.db.Where("username = ? AND status = ?", username, "active").First(&user).Error; err != nil {
		return "", user, httperr.New(httperr.AuthForbidden, "用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", user, httperr.New(httperr.AuthForbidden, "用户名或密码错误")
	}
	if scope == ScopeGlasses && deviceID == nil {
		return "", user, httperr.New(httperr.DeviceRevoked, "device is required")
	}
	token, err := s.IssueAccessToken(user.ID, deviceID, scope)
	return token, user, err
}

// TokenPair 是眼镜端登录/刷新返回的令牌组合。
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int64
	RefreshExpiresIn int64
	Session          database.DeviceSession
}

// LoginWithRefresh 校验用户名密码和眼镜设备，并签发 access/refresh token。
func (s *Service) LoginWithRefresh(username, password string, deviceID uint64) (TokenPair, database.User, error) {
	var user database.User
	if err := s.db.Where("username = ? AND status = ?", username, "active").First(&user).Error; err != nil {
		return TokenPair{}, user, httperr.New(httperr.AuthForbidden, "用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return TokenPair{}, user, httperr.New(httperr.AuthForbidden, "用户名或密码错误")
	}
	var device database.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		return TokenPair{}, user, httperr.New(httperr.DeviceRevoked, "device no longer exists")
	}
	if err := ensureDeviceAllowed(device, user.ID); err != nil {
		return TokenPair{}, user, err
	}
	pair, err := s.IssueTokenPair(user.ID, deviceID)
	return pair, user, err
}

// IssueTokenPair 为眼镜端签发 access token 和 refresh token，并记录设备会话。
func (s *Service) IssueTokenPair(userID, deviceID uint64) (TokenPair, error) {
	access, err := s.IssueAccessToken(userID, &deviceID, ScopeGlasses)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, jti, err := randomToken()
	if err != nil {
		return TokenPair{}, err
	}
	now := time.Now().UTC()
	session := database.DeviceSession{DeviceID: deviceID, UserID: userID, RefreshJTI: jti, Status: "active", RefreshUntil: now.Add(s.refreshTTL)}
	if err := s.db.Create(&session).Error; err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(s.accessTTL.Seconds()), RefreshExpiresIn: int64(s.refreshTTL.Seconds()), Session: session}, nil
}

// RefreshAccessToken 校验 refresh token 并签发新的 access token。
func (s *Service) RefreshAccessToken(refreshToken string, deviceID uint64) (TokenPair, error) {
	jti, err := refreshJTI(refreshToken)
	if err != nil {
		return TokenPair{}, httperr.New(httperr.AuthTokenExpired, "refresh token invalid")
	}
	var session database.DeviceSession
	if err := s.db.Where("refresh_jti = ? AND status = ?", jti, "active").First(&session).Error; err != nil {
		return TokenPair{}, httperr.New(httperr.AuthTokenExpired, "refresh token expired or revoked")
	}
	if deviceID != 0 && session.DeviceID != deviceID {
		return TokenPair{}, httperr.New(httperr.AuthForbidden, "refresh token device mismatch")
	}
	if time.Now().UTC().After(session.RefreshUntil) {
		return TokenPair{}, httperr.New(httperr.AuthTokenExpired, "refresh token expired")
	}
	var device database.Device
	if err := s.db.First(&device, session.DeviceID).Error; err != nil {
		return TokenPair{}, httperr.New(httperr.DeviceRevoked, "device no longer exists")
	}
	if err := ensureDeviceAllowed(device, session.UserID); err != nil {
		return TokenPair{}, err
	}
	access, err := s.IssueAccessToken(session.UserID, &session.DeviceID, ScopeGlasses)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refreshToken, ExpiresIn: int64(s.accessTTL.Seconds()), RefreshExpiresIn: int64(time.Until(session.RefreshUntil).Seconds()), Session: session}, nil
}

// Logout 撤销当前设备会话。refreshToken 为空时撤销当前用户和设备的全部 active session。
func (s *Service) Logout(userID uint64, deviceID *uint64, refreshToken string) error {
	query := s.db.Model(&database.DeviceSession{}).Where("user_id = ? AND status = ?", userID, "active")
	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}
	if refreshToken != "" {
		jti, err := refreshJTI(refreshToken)
		if err != nil {
			return httperr.New(httperr.AuthTokenExpired, "refresh token invalid")
		}
		query = query.Where("refresh_jti = ?", jti)
	}
	return query.Update("status", "revoked").Error
}

// CurrentUserInfo 查询当前用户、组织和设备信息。
func (s *Service) CurrentUserInfo(userID uint64, deviceID *uint64) (database.User, *database.Organization, *database.Device, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return user, nil, nil, err
	}
	var org *database.Organization
	if user.OrgCode != "" {
		var found database.Organization
		if err := s.db.Where("code = ?", user.OrgCode).First(&found).Error; err == nil {
			org = &found
		} else if err != gorm.ErrRecordNotFound {
			return user, nil, nil, err
		}
	}
	var device *database.Device
	if deviceID != nil {
		var found database.Device
		if err := s.db.First(&found, *deviceID).Error; err != nil {
			return user, org, nil, err
		}
		device = &found
	}
	return user, org, device, nil
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
func randomToken() (string, string, error) {
	jti := uuid.NewString()
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(raw)
	return "rt_" + jti + "." + secret, jti, nil
}

func refreshJTI(token string) (string, error) {
	if !strings.HasPrefix(token, "rt_") {
		return "", httperr.New(httperr.AuthTokenExpired, "refresh token invalid")
	}
	body := strings.TrimPrefix(token, "rt_")
	parts := strings.SplitN(body, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", httperr.New(httperr.AuthTokenExpired, "refresh token invalid")
	}
	return parts[0], nil
}

func ensureDeviceAllowed(device database.Device, userID uint64) error {
	switch device.Status {
	case "revoked":
		return httperr.New(httperr.DeviceRevoked, "device revoked")
	case "lost_disabled":
		return httperr.New(httperr.DeviceDisabledLost, "device disabled as lost")
	}
	if device.Status != "active" {
		return httperr.New(httperr.DeviceRevoked, "device is not active")
	}
	if device.BoundUserID != nil && *device.BoundUserID != userID {
		return httperr.New(httperr.AuthForbidden, "device is bound to another user")
	}
	return nil
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
