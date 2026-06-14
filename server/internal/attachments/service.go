// Package attachments 管理巡检结果证据上传元数据。它负责签发 MinIO 预签名上传地址、记录
// 拍摄上下文、校验已上传对象、保存内容哈希，并将证据绑定到任务节点结果。
package attachments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

const (
	BindUploaded = "uploaded"
	BindBound    = "bound"
	BindDeleted  = "deleted"
)

type Service struct {
	db     *gorm.DB
	client *minio.Client
	bucket string
	cfg    config.Config
}

type PresignInput struct {
	FileName    string     `json:"file_name"`
	ContentType string     `json:"content_type"`
	SizeBytes   int64      `json:"size_bytes"`
	CaptureTime *time.Time `json:"capture_time"`
	DeviceID    *uint64    `json:"device_id"`
	GPSLat      *float64   `json:"gps_lat"`
	GPSLng      *float64   `json:"gps_lng"`
}

type PresignResult struct {
	Attachment database.Attachment `json:"attachment"`
	UploadURL  string              `json:"upload_url"`
	ObjectKey  string              `json:"object_key"`
}

// NewService 创建附件服务，并初始化 MinIO 客户端与目标 bucket 配置。
func NewService(db *gorm.DB, cfg config.Config) (*Service, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""), Secure: cfg.S3UseSSL})
	if err != nil {
		return nil, err
	}
	return &Service{db: db, client: client, bucket: cfg.S3Bucket, cfg: cfg}, nil
}

// Presign 创建待上传附件记录，并返回客户端直传对象存储所需的预签名地址。
func (s *Service) Presign(ctx context.Context, userID uint64, input PresignInput) (PresignResult, error) {
	if !allowedContentType(input.ContentType) {
		return PresignResult{}, httperr.New(httperr.AttachmentNotUploaded, "unsupported attachment content type")
	}
	if input.SizeBytes <= 0 || input.SizeBytes > s.maxBytes(input.ContentType) {
		return PresignResult{}, httperr.New(httperr.AttachmentNotUploaded, "attachment size is not allowed")
	}
	objectKey := fmt.Sprintf("evidence/%s/%s", time.Now().UTC().Format("2006/01/02"), uuid.NewString())
	url, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, 15*time.Minute)
	if err != nil {
		return PresignResult{}, err
	}
	attachment := database.Attachment{
		ObjectKey:   objectKey,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		SizeBytes:   input.SizeBytes,
		BindStatus:  BindUploaded,
		UserID:      userID,
		DeviceID:    input.DeviceID,
		CaptureTime: input.CaptureTime,
		GPSLat:      input.GPSLat,
		GPSLng:      input.GPSLng,
	}
	return PresignResult{Attachment: attachment, UploadURL: url.String(), ObjectKey: objectKey}, s.db.Create(&attachment).Error
}

// MarkUploaded 校验对象存储中的文件并回写上传状态、大小和哈希。
func (s *Service) MarkUploaded(ctx context.Context, attachmentID uint64) error {
	var attachment database.Attachment
	if err := s.db.First(&attachment, attachmentID).Error; err != nil {
		return err
	}
	object, err := s.client.GetObject(ctx, s.bucket, attachment.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, object); err != nil {
		return err
	}
	now := time.Now().UTC()
	return s.db.Model(&attachment).Updates(map[string]any{"sha256": hex.EncodeToString(hash.Sum(nil)), "upload_time": now}).Error
}

// CleanupOrphans 清理超过指定时间仍未绑定业务结果的孤立附件记录。
func (s *Service) CleanupOrphans(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().UTC().Add(-olderThan)
	var attachments []database.Attachment
	if err := s.db.Where("bind_status = ? AND created_at < ?", BindUploaded, cutoff).Find(&attachments).Error; err != nil {
		return err
	}
	for _, attachment := range attachments {
		_ = s.client.RemoveObject(ctx, s.bucket, attachment.ObjectKey, minio.RemoveObjectOptions{})
		s.db.Model(&attachment).Update("bind_status", BindDeleted)
	}
	return nil
}

// allowedContentType 判断上传 MIME 类型是否属于证据允许范围。
func allowedContentType(contentType string) bool {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/png", "audio/aac", "audio/m4a", "audio/wav", "audio/x-wav":
		return true
	default:
		return false
	}
}

// maxBytes 根据文件类型返回对应的最大允许上传字节数。
func (s *Service) maxBytes(contentType string) int64 {
	if strings.HasPrefix(strings.ToLower(contentType), "audio/") {
		return s.cfg.AudioMaxBytes
	}
	return s.cfg.RequiredPhotoMaxBytes
}
