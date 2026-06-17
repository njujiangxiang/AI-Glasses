package businesscodes

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"

	DateFormatDaily = "yyyyMMdd"
	DateFormatShort = "yyMMdd"
	redisTTLSeconds = 48 * 60 * 60
)

// goDateFormats 将业务日期格式标识映射到 Go time.Format 模板。
var goDateFormats = map[string]string{
	DateFormatDaily: "20060102",
	DateFormatShort: "060102",
}

var (
	codePattern = regexp.MustCompile(`^[A-Z0-9_-]{1,64}$`)
	incrScript  = `local current = tonumber(redis.call('GET', KEYS[1]) or '0')
if current >= tonumber(ARGV[2]) then
  return current + 1
end
current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current`
)

// Input 是创建和更新业务编码配置共用的输入结构。
type Input struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	DateFormat   string `json:"date_format"`
	SeqPadding   int    `json:"seq_padding"`
	Separator    string `json:"separator"`
	UseSeparator bool   `json:"use_separator"`
	Status       string `json:"status"`
}

type Service struct {
	db       *gorm.DB
	redis    *redis.Client
	location *time.Location
	now      func() time.Time
}

// NewService 创建业务编码服务，注入数据库、Redis 和固定业务时区。
func NewService(db *gorm.DB, redisClient *redis.Client) (*Service, error) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	return &Service{db: db, redis: redisClient, location: loc, now: time.Now}, nil
}

// SetNowForTest 覆盖当前时间来源，供按日流水号测试使用。
func (s *Service) SetNowForTest(now func() time.Time) {
	if now == nil {
		s.now = time.Now
		return
	}
	s.now = now
}

// Create 创建业务编码规则。
func (s *Service) Create(input Input) (database.BusinessCode, error) {
	model := database.BusinessCode{
		Name:         strings.TrimSpace(input.Name),
		Code:         normalizeCode(input.Code),
		DateFormat:   normalizeDateFormat(input.DateFormat),
		SeqPadding:   input.SeqPadding,
		Separator:    strings.TrimSpace(input.Separator),
		UseSeparator: input.UseSeparator,
		Status:       normalizeStatus(input.Status),
	}
	if err := s.validate(model, 0); err != nil {
		return database.BusinessCode{}, err
	}
	if err := s.db.Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return database.BusinessCode{}, httperr.New(httperr.ValidationFailed, "业务编码代码已存在")
		}
		return database.BusinessCode{}, err
	}
	return model, nil
}

// Update 更新业务编码规则。Code 字段不可变，以数据库中的原始值为准。
func (s *Service) Update(id uint64, input Input) (database.BusinessCode, error) {
	var model database.BusinessCode
	if err := s.db.First(&model, id).Error; err != nil {
		return database.BusinessCode{}, notFound(err, "business code not found")
	}
	model.Name = strings.TrimSpace(input.Name)
	// Code 不可变：保持数据库中的原始值，忽略 input.Code。
	model.DateFormat = normalizeDateFormat(input.DateFormat)
	model.SeqPadding = input.SeqPadding
	model.Separator = strings.TrimSpace(input.Separator)
	model.UseSeparator = input.UseSeparator
	model.Status = normalizeStatus(input.Status)
	if err := s.validate(model, id); err != nil {
		return database.BusinessCode{}, err
	}
	if err := s.db.Save(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return database.BusinessCode{}, httperr.New(httperr.ValidationFailed, "业务编码代码已存在")
		}
		return database.BusinessCode{}, err
	}
	return model, nil
}

// Get 查询单个业务编码规则。
func (s *Service) Get(id uint64) (database.BusinessCode, error) {
	var model database.BusinessCode
	if err := s.db.First(&model, id).Error; err != nil {
		return database.BusinessCode{}, notFound(err, "business code not found")
	}
	return model, nil
}

// List 查询业务编码规则列表，支持按名称或代码模糊过滤。
func (s *Service) List(keyword string) ([]database.BusinessCode, error) {
	var items []database.BusinessCode
	query := s.db.Order("code asc")
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		query = query.Where(`code LIKE ? ESCAPE '\' OR name LIKE ? ESCAPE '\'`, like, like)
	}
	return items, query.Find(&items).Error
}

// Enable 启用业务编码规则。
func (s *Service) Enable(id uint64) error { return s.setStatus(id, StatusActive) }

// Disable 停用业务编码规则。
func (s *Service) Disable(id uint64) error { return s.setStatus(id, StatusDisabled) }

// Delete 删除业务编码规则。后续业务模块引用 code 后，应升级为软删除或引用保护。
func (s *Service) Delete(id uint64) error {
	result := s.db.Delete(&database.BusinessCode{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "business code not found")
	}
	return nil
}

// GenerateDaily 按配置生成按日递增的业务编号。
//
// 流水号由 Redis INCR 分配，不参与调用方的数据库事务。调用方如果在生成编号后回滚
// 自己的事务，会留下流水号空洞；这是该公共方法的预期语义，业务接入方必须接受空洞，
// 不要为了连续性把编号生成放回本地内存或数据库锁里。
func (s *Service) GenerateDaily(ctx context.Context, code string) (string, error) {
	code = normalizeCode(code)
	if code == "" {
		return "", httperr.New(httperr.ValidationFailed, "请输入业务编码代码")
	}
	var cfg database.BusinessCode
	if err := s.db.Where("code = ?", code).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", httperr.New(httperr.ValidationFailed, fmt.Sprintf("请先在业务编码配置中配置代码 %s", code))
		}
		return "", err
	}
	if cfg.Status == StatusDisabled {
		return "", httperr.New(httperr.ValidationFailed, fmt.Sprintf("业务编码 %s 已停用，请先启用后再生成", code))
	}
	if cfg.Status != StatusActive {
		return "", httperr.New(httperr.ValidationFailed, fmt.Sprintf("业务编码 %s 的状态不正确", code))
	}

	now := s.now().In(s.location)
	goFmt := goDateFormats[cfg.DateFormat]
	date := now.Format(goFmt)
	// Redis key 统一使用完整日期（yyyyMMdd），避免 yyMMdd 跨世纪冲突。
	keyDate := now.Format("20060102")
	key := fmt.Sprintf("BNO:%s:%s", cfg.Code, keyDate)
	maxSeq := int64(math.Pow10(cfg.SeqPadding)) - 1
	if s.redis == nil {
		return "", httperr.New(httperr.InternalError, "业务编码流水号服务未配置，生成编号暂不可用")
	}
	result, err := s.redis.Eval(ctx, incrScript, []string{key}, redisTTLSeconds, maxSeq).Int64()
	if err != nil {
		return "", httperr.New(httperr.InternalError, "业务编码流水号服务不可用，请检查 Redis")
	}
	if result > maxSeq {
		return "", httperr.New(httperr.ValidationFailed, "当日流水号已超出位数上限，请调整流水号位数")
	}
	seq := fmt.Sprintf("%0*d", cfg.SeqPadding, result)
	if cfg.UseSeparator {
		return strings.Join([]string{cfg.Code, date, seq}, cfg.Separator), nil
	}
	return cfg.Code + date + seq, nil
}

func (s *Service) setStatus(id uint64, status string) error {
	result := s.db.Model(&database.BusinessCode{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "business code not found")
	}
	return nil
}

func (s *Service) validate(model database.BusinessCode, currentID uint64) error {
	if model.Name == "" {
		return httperr.New(httperr.ValidationFailed, "编码名称不能为空")
	}
	if !codePattern.MatchString(model.Code) {
		return httperr.New(httperr.ValidationFailed, "业务编码代码只能包含大写字母、数字、下划线和中划线")
	}
	if _, ok := goDateFormats[model.DateFormat]; !ok {
		return httperr.New(httperr.ValidationFailed, fmt.Sprintf("业务编码 %s 的日期格式不正确，支持 yyyyMMdd 和 yyMMdd", model.Code))
	}
	if model.SeqPadding < 1 || model.SeqPadding > 12 {
		return httperr.New(httperr.ValidationFailed, fmt.Sprintf("业务编码 %s 的流水号位数不正确", model.Code))
	}
	if model.UseSeparator && model.Separator == "" {
		return httperr.New(httperr.ValidationFailed, "启用分隔符时请填写分隔符")
	}
	if len([]rune(model.Separator)) > 8 {
		return httperr.New(httperr.ValidationFailed, "分隔符长度不能超过8个字符")
	}
	if model.Status != StatusActive && model.Status != StatusDisabled {
		return httperr.New(httperr.ValidationFailed, "业务编码状态不正确")
	}
	var same database.BusinessCode
	if err := s.db.Where("code = ?", model.Code).First(&same).Error; err == nil && same.ID != currentID {
		return httperr.New(httperr.ValidationFailed, "业务编码代码已存在")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func normalizeCode(code string) string { return strings.ToUpper(strings.TrimSpace(code)) }

// escapeLike 转义 SQL LIKE 通配符，防止用户输入中的 % _ \ 被解释为模式匹配。
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func normalizeDateFormat(dateFormat string) string {
	dateFormat = strings.TrimSpace(dateFormat)
	if dateFormat == "" {
		return DateFormatDaily
	}
	return dateFormat
}

func normalizeStatus(status string) string {
	if status == "" {
		return StatusActive
	}
	return status
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}
