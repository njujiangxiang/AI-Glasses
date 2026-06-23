package users

import (
	"bytes"
	"regexp"
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"

	GenderMale    = "male"
	GenderFemale  = "female"
	GenderUnknown = "unknown"

	MaxAvatarBytes = 2 << 20
)

var idCardPattern = regexp.MustCompile(`^\d{17}[\dXx]$`)

type ListQuery struct {
	Keyword  string
	OrgCode  string
	Status   string
	Page     int
	PageSize int
}

type CreateInput struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	BirthYear  int    `json:"birth_year"`
	BirthMonth int    `json:"birth_month"`
	IDCardNo   string `json:"id_card_no"`
	OrgCode    string `json:"org_code"`
	Status     string `json:"status"`
}

type UpdateInput struct {
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	BirthYear  int    `json:"birth_year"`
	BirthMonth int    `json:"birth_month"`
	IDCardNo   string `json:"id_card_no"`
	OrgCode    string `json:"org_code"`
	Status     string `json:"status"`
}

type UserDTO struct {
	ID                uint64    `json:"id"`
	Username          string    `json:"username"`
	DisplayName       string    `json:"display_name"`
	Name              string    `json:"name"`
	Gender            string    `json:"gender"`
	BirthYear         int       `json:"birth_year"`
	BirthMonth        int       `json:"birth_month"`
	IDCardNo          string    `json:"id_card_no"`
	OrgCode           string    `json:"org_code"`
	Status            string    `json:"status"`
	AvatarContentType string    `json:"avatar_content_type"`
	AvatarSize        int64     `json:"avatar_size"`
	HasAvatar         bool      `json:"has_avatar"`
	CreatedAt         time.Time `json:"CreatedAt"`
	UpdatedAt         time.Time `json:"UpdatedAt"`
}

type ListResult struct {
	Items []UserDTO `json:"items"`
	Total int64     `json:"total"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建用户管理服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// List 查询用户列表，显式排除头像二进制大字段。
func (s *Service) List(query ListQuery) (ListResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := s.db.Model(&database.User{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR name LIKE ? OR display_name LIKE ? OR id_card_no LIKE ?", like, like, like, like)
	}
	if orgCode := strings.TrimSpace(query.OrgCode); orgCode != "" {
		db = db.Where("org_code = ?", orgCode)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}
	var models []database.User
	if err := db.Select(userColumns()).Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return ListResult{}, err
	}
	items := make([]UserDTO, 0, len(models))
	for _, user := range models {
		items = append(items, toDTO(user))
	}
	return ListResult{Items: items, Total: total}, nil
}

// ListAll 查询所有启用用户（不分页，用于下拉选择）。
func (s *Service) ListAll() ([]UserDTO, error) {
	var models []database.User
	if err := s.db.Select(userColumns()).Where("status = ?", StatusActive).Order("id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]UserDTO, 0, len(models))
	for _, user := range models {
		items = append(items, toDTO(user))
	}
	return items, nil
}

// Get 查询用户详情，不返回头像二进制内容。
func (s *Service) Get(id uint64) (UserDTO, error) {
	var user database.User
	if err := s.db.Select(userColumns()).First(&user, id).Error; err != nil {
		return UserDTO{}, notFound(err, "user not found")
	}
	return toDTO(user), nil
}

// Create 创建后台用户基础资料，密码使用 bcrypt 哈希存储。
func (s *Service) Create(input CreateInput) (UserDTO, error) {
	password := strings.TrimSpace(input.Password)
	if len(password) < 4 {
		return UserDTO{}, httperr.New(httperr.ValidationFailed, "密码长度不能少于4位")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return UserDTO{}, err
	}
	user := database.User{
		Username:     strings.TrimSpace(input.Username),
		PasswordHash: string(hash),
		DisplayName:  strings.TrimSpace(input.Name),
		Name:         strings.TrimSpace(input.Name),
		Gender:       normalizeGender(input.Gender),
		BirthYear:    input.BirthYear,
		BirthMonth:   input.BirthMonth,
		IDCardNo:     strings.TrimSpace(input.IDCardNo),
		OrgCode:      strings.TrimSpace(input.OrgCode),
		Status:       normalizeStatus(input.Status),
	}
	if err := s.validateUser(user, 0, true); err != nil {
		return UserDTO{}, err
	}
	if err := s.db.Create(&user).Error; err != nil {
		return UserDTO{}, err
	}
	return toDTO(user), nil
}

// Update 更新用户基础资料。
func (s *Service) Update(id uint64, input UpdateInput) (UserDTO, error) {
	var user database.User
	if err := s.db.First(&user, id).Error; err != nil {
		return UserDTO{}, notFound(err, "user not found")
	}
	user.Name = strings.TrimSpace(input.Name)
	user.DisplayName = user.Name
	user.Gender = normalizeGender(input.Gender)
	user.BirthYear = input.BirthYear
	user.BirthMonth = input.BirthMonth
	user.IDCardNo = strings.TrimSpace(input.IDCardNo)
	user.OrgCode = strings.TrimSpace(input.OrgCode)
	user.Status = normalizeStatus(input.Status)
	if err := s.validateUser(user, id, false); err != nil {
		return UserDTO{}, err
	}
	if err := s.db.Save(&user).Error; err != nil {
		return UserDTO{}, err
	}
	return toDTO(user), nil
}

// Enable 启用用户。
func (s *Service) Enable(id uint64) error { return s.setStatus(id, StatusActive) }

// Disable 停用用户。
func (s *Service) Disable(id uint64) error { return s.setStatus(id, StatusDisabled) }

// SetAvatar 将用户头像保存到数据库。
func (s *Service) SetAvatar(id uint64, data []byte, contentType string) error {
	if len(data) == 0 {
		return httperr.New(httperr.ValidationFailed, "avatar is required")
	}
	if len(data) > MaxAvatarBytes {
		return httperr.New(httperr.ValidationFailed, "avatar is too large")
	}
	if !validAvatarContentType(contentType) {
		return httperr.New(httperr.ValidationFailed, "invalid avatar content type")
	}
	result := s.db.Model(&database.User{}).Where("id = ?", id).Updates(map[string]any{
		"avatar_data":         data,
		"avatar_content_type": contentType,
		"avatar_size":         len(data),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "user not found")
	}
	return nil
}

// GetAvatar 读取数据库中保存的用户头像。
func (s *Service) GetAvatar(id uint64) ([]byte, string, error) {
	var user database.User
	if err := s.db.Select("avatar_data", "avatar_content_type", "avatar_size").First(&user, id).Error; err != nil {
		return nil, "", notFound(err, "user not found")
	}
	if user.AvatarSize == 0 || len(user.AvatarData) == 0 {
		return nil, "", httperr.New(httperr.ResourceNotFound, "avatar not found")
	}
	return user.AvatarData, user.AvatarContentType, nil
}

func (s *Service) setStatus(id uint64, status string) error {
	result := s.db.Model(&database.User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "user not found")
	}
	return nil
}

func (s *Service) validateUser(user database.User, currentID uint64, requireUsername bool) error {
	if requireUsername && user.Username == "" {
		return httperr.New(httperr.ValidationFailed, "username is required")
	}
	if user.Name == "" {
		return httperr.New(httperr.ValidationFailed, "name is required")
	}
	if user.Status != StatusActive && user.Status != StatusDisabled {
		return httperr.New(httperr.ValidationFailed, "invalid user status")
	}
	if user.Gender != "" && user.Gender != GenderMale && user.Gender != GenderFemale && user.Gender != GenderUnknown {
		return httperr.New(httperr.ValidationFailed, "invalid gender")
	}
	if user.BirthYear != 0 {
		year := time.Now().UTC().Year()
		if user.BirthYear < 1900 || user.BirthYear > year || user.BirthMonth < 1 || user.BirthMonth > 12 {
			return httperr.New(httperr.ValidationFailed, "invalid birth date")
		}
	} else if user.BirthMonth != 0 {
		return httperr.New(httperr.ValidationFailed, "birth year is required")
	}
	if user.IDCardNo != "" {
		if !idCardPattern.MatchString(user.IDCardNo) {
			return httperr.New(httperr.ValidationFailed, "invalid id card number")
		}
		var same database.User
		if err := s.db.Select("id").Where("id_card_no = ?", user.IDCardNo).First(&same).Error; err == nil && same.ID != currentID {
			return httperr.New(httperr.ValidationFailed, "id card number already exists")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
	}
	if user.OrgCode != "" {
		var count int64
		if err := s.db.Model(&database.Organization{}).Where("code = ? AND status = ?", user.OrgCode, StatusActive).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return httperr.New(httperr.ValidationFailed, "organization not found or disabled")
		}
	}
	return nil
}

func userColumns() []string {
	return []string{"id", "username", "display_name", "name", "gender", "birth_year", "birth_month", "id_card_no", "org_code", "status", "avatar_content_type", "avatar_size", "created_at", "updated_at"}
}

func toDTO(user database.User) UserDTO {
	return UserDTO{
		ID:                user.ID,
		Username:          user.Username,
		DisplayName:       user.DisplayName,
		Name:              user.Name,
		Gender:            user.Gender,
		BirthYear:         user.BirthYear,
		BirthMonth:        user.BirthMonth,
		IDCardNo:          user.IDCardNo,
		OrgCode:           user.OrgCode,
		Status:            user.Status,
		AvatarContentType: user.AvatarContentType,
		AvatarSize:        user.AvatarSize,
		HasAvatar:         user.AvatarSize > 0,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func normalizeGender(gender string) string {
	gender = strings.TrimSpace(gender)
	if gender == "" {
		return GenderUnknown
	}
	return gender
}

func normalizeStatus(status string) string {
	if status == "" {
		return StatusActive
	}
	return status
}

func validAvatarContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/webp"
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}

// DetectContentType 规范化前端上传头像时传入的 MIME 类型。
func DetectContentType(data []byte, declared string) string {
	declared = strings.ToLower(strings.TrimSpace(declared))
	if validAvatarContentType(declared) {
		return declared
	}
	if bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}) {
		return "image/jpeg"
	}
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4e, 0x47}) {
		return "image/png"
	}
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	return declared
}
