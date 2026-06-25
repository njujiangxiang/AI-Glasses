package glass_api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aiglasses/server/internal/attachments"
	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

func (h *Handler) refreshToken(c *gin.Context) {
	var body refreshRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	pair, err := h.auth.RefreshAccessToken(body.RefreshToken, body.DeviceID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken, "token_type": "Bearer", "expires_in": pair.ExpiresIn, "refresh_expires_in": pair.RefreshExpiresIn, "device_id": pair.Session.DeviceID})
}

func (h *Handler) logout(c *gin.Context) {
	var body logoutRequest
	_ = c.ShouldBindJSON(&body)
	if err := h.auth.Logout(authUserID(c), authDeviceID(c), body.RefreshToken); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"logged_out": true})
}

func (h *Handler) userInfo(c *gin.Context) {
	user, org, device, err := h.auth.CurrentUserInfo(authUserID(c), authDeviceID(c))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"user": user, "organization": org, "device": device, "scope": "glasses"})
}

func (h *Handler) registerDevice(c *gin.Context) {
	var body deviceRegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	serial := strings.TrimSpace(body.SerialNo)
	if serial == "" {
		serial = strings.TrimSpace(body.DeviceSN)
	}
	if serial == "" {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "serial_no is required"))
		return
	}
	name := strings.TrimSpace(body.DeviceName)
	if name == "" {
		name = serial
	}
	device := database.Device{SerialNo: serial}
	if err := h.db.Where("serial_no = ?", serial).FirstOrCreate(&device, database.Device{SerialNo: serial, Name: name, Status: "pending"}).Error; err != nil {
		httperr.Respond(c, err)
		return
	}
	if device.Name == "" {
		_ = h.db.Model(&device).Update("name", name).Error
		device.Name = name
	}
	httperr.OK(c, gin.H{"device_id": idString(device.ID), "serial_no": device.SerialNo, "status": device.Status, "is_bound": device.BoundUserID != nil, "bind_user_id": device.BoundUserID})
}

func (h *Handler) reportDeviceStatus(c *gin.Context) {
	var body deviceStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	deviceID := authDeviceID(c)
	if deviceID == nil || *deviceID != body.DeviceID {
		httperr.Respond(c, httperr.New(httperr.AuthForbidden, "device_id mismatch"))
		return
	}
	now := time.Now().UTC()
	if err := h.db.Model(&database.Device{}).Where("id = ?", body.DeviceID).Update("updated_at", now).Error; err != nil {
		httperr.Respond(c, err)
		return
	}
	detail, _ := json.Marshal(body)
	_ = h.db.Create(&database.AuditLog{ActorID: authUserID(c), Action: "device.status_report", Target: "device", TargetID: body.DeviceID, Detail: string(detail), CreatedAt: now}).Error
	httperr.OK(c, gin.H{"reported": true, "server_time": formatTime(now), "device_id": body.DeviceID})
}

func (h *Handler) createDefect(c *gin.Context) {
	var body createDefectRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	taskID, err := parseID(body.TaskID)
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	nodeID, err := parseID(body.NodeID)
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid node_id"))
		return
	}
	if err := h.tasks.EnsureAccessible(taskID, authUserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	var node database.InspectionTaskNode
	if err := h.db.Where("task_id = ? AND id = ?", taskID, nodeID).First(&node).Error; err != nil {
		httperr.Respond(c, httperr.New(httperr.ResourceNotFound, "task node not found"))
		return
	}
	desc := strings.TrimSpace(body.Description)
	if desc == "" {
		desc = "AR眼镜端上报缺陷"
	}
	defect, err := h.defects.Create(taskID, nodeID, authUserID(c), desc)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	if len(body.AttachmentIDs) > 0 {
		_ = h.db.Model(&database.Attachment{}).Where("id IN ?", body.AttachmentIDs).Updates(map[string]any{"task_id": taskID, "node_id": nodeID, "bind_status": attachments.BindBound}).Error
	}
	httperr.OK(c, gin.H{"defect_id": idString(defect.ID), "status": defect.Status})
}

func (h *Handler) uploadAttachment(c *gin.Context) {
	result, err := h.saveUploadedFile(c, "file", 0)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

func (h *Handler) uploadAttachmentBatch(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	items := make([]batchUploadItem, 0, len(files))
	for i, file := range files {
		result, err := h.saveMultipartFile(c, file, i)
		if err != nil {
			items = append(items, batchUploadItem{Index: i, Success: false, Error: err.Error()})
			continue
		}
		items = append(items, batchUploadItem{Index: i, Success: true, Attachment: &result})
	}
	httperr.OK(c, gin.H{"items": items})
}

func (h *Handler) attachmentDetail(c *gin.Context) {
	id, err := parseID(c.Param("attachment_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid attachment_id"))
		return
	}
	var attachment database.Attachment
	if err := h.db.First(&attachment, id).Error; err != nil {
		httperr.Respond(c, err)
		return
	}
	if attachment.UserID != authUserID(c) {
		httperr.Respond(c, httperr.New(httperr.ResourceNotFound, "attachment not found"))
		return
	}
	downloadURL := "local-dev://" + attachment.ObjectKey
	httperr.OK(c, gin.H{"attachment": attachment, "download_url": downloadURL, "expires_in": 900})
}

func (h *Handler) realtimeQuery(c *gin.Context) {
	var body realtimeQueryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"query_result": gin.H{"status": "normal", "temperature": "36.5℃", "source": "mock"}, "formatted_result": "设备状态正常，温度36.5℃", "mocked": true, "task_id": body.TaskID, "node_id": body.NodeID})
}

func (h *Handler) invokeAlgorithm(c *gin.Context) {
	var body algorithmInvokeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"algorithm_result": "AI识别结果：未发现明显异常", "is_abnormal": "0", "abnormal_type": "", "confidence": 0.98, "mocked": true, "algorithm_id": body.AlgorithmID})
}

func (h *Handler) taskTypes(c *gin.Context) {
	items := []taskTypeResponse{
		{TaskTypeID: "1", TypeCode: "check", TypeName: "检查确认", TypeDesc: "确认设备状态或选项", SupportAlgorithm: "0", SupportQuery: "0", SupportMandatory: "1"},
		{TaskTypeID: "2", TypeCode: "read", TypeName: "读数记录", TypeDesc: "读取仪表或设备数值", SupportAlgorithm: "1", SupportQuery: "1", SupportMandatory: "1"},
		{TaskTypeID: "3", TypeCode: "photo", TypeName: "拍照留证", TypeDesc: "拍摄现场照片", SupportAlgorithm: "1", SupportQuery: "0", SupportMandatory: "1"},
		{TaskTypeID: "4", TypeCode: "text", TypeName: "文本记录", TypeDesc: "填写文本备注", SupportAlgorithm: "0", SupportQuery: "0", SupportMandatory: "0"},
	}
	httperr.OK(c, items)
}

func (h *Handler) saveUploadedFile(c *gin.Context, field string, index int) (attachmentUploadResult, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return attachmentUploadResult{}, httperr.New(httperr.AttachmentNotUploaded, "file is required")
	}
	return h.saveMultipartFile(c, file, index)
}

func (h *Handler) saveMultipartFile(c *gin.Context, file *multipart.FileHeader, index int) (attachmentUploadResult, error) {
	opened, err := file.Open()
	if err != nil {
		return attachmentUploadResult{}, err
	}
	defer opened.Close()
	data, err := io.ReadAll(opened)
	if err != nil {
		return attachmentUploadResult{}, err
	}
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = c.PostForm("content_type")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !allowedUploadType(contentType) {
		return attachmentUploadResult{}, httperr.New(httperr.AttachmentNotUploaded, "unsupported attachment content type")
	}
	maxBytes := h.cfg.RequiredPhotoMaxBytes
	if strings.HasPrefix(strings.ToLower(contentType), "audio/") {
		maxBytes = h.cfg.AudioMaxBytes
	}
	if int64(len(data)) <= 0 || int64(len(data)) > maxBytes {
		return attachmentUploadResult{}, httperr.New(httperr.AttachmentNotUploaded, "attachment size is not allowed")
	}
	taskID := parseOptionalID(c.PostForm("task_id"))
	nodeID := parseOptionalID(c.PostForm("node_id"))
	resultID := parseOptionalID(c.PostForm("result_id"))
	captureTime, _ := parseTimePtr(c.PostForm("capture_time"))
	gpsLat := parseOptionalFloat(c.PostForm("gps_lat"))
	gpsLng := parseOptionalFloat(c.PostForm("gps_lng"))
	sum := sha256.Sum256(data)
	objectKey := fmt.Sprintf("local-dev/evidence/%s/%s%s", time.Now().UTC().Format("2006/01/02"), uuid.NewString(), filepath.Ext(file.Filename))
	attachment := database.Attachment{ObjectKey: objectKey, FileName: file.Filename, ContentType: contentType, SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), BindStatus: attachments.BindUploaded, TaskID: taskID, NodeID: nodeID, ResultID: resultID, UserID: authUserID(c), DeviceID: authDeviceID(c), CaptureTime: captureTime, UploadTime: ptrTime(time.Now().UTC()), GPSLat: gpsLat, GPSLng: gpsLng}
	if err := h.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&attachment).Error; err != nil {
		return attachmentUploadResult{}, err
	}
	return attachmentUploadResult{AttachmentID: idString(attachment.ID), ObjectKey: attachment.ObjectKey, FileName: attachment.FileName, SizeBytes: attachment.SizeBytes, ContentType: attachment.ContentType, Attachment: attachment}, nil
}

func allowedUploadType(contentType string) bool {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/png", "audio/aac", "audio/m4a", "audio/wav", "audio/x-wav", "video/mp4", "application/octet-stream":
		return true
	default:
		return false
	}
}

func authUserID(c *gin.Context) uint64    { return auth.UserID(c) }
func authDeviceID(c *gin.Context) *uint64 { return auth.DeviceID(c) }

func parseOptionalID(value string) *uint64 {
	if value == "" {
		return nil
	}
	id, err := parseID(value)
	if err != nil {
		return nil
	}
	return &id
}

func parseOptionalFloat(value string) *float64 {
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func ptrTime(value time.Time) *time.Time { return &value }
