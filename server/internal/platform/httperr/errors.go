package httperr

import "net/http"

type ErrorCode string

const (
	AuthTokenExpired         ErrorCode = "AUTH_TOKEN_EXPIRED"
	AuthForbidden            ErrorCode = "AUTH_FORBIDDEN"
	DeviceRevoked            ErrorCode = "DEVICE_REVOKED"
	DeviceDisabledLost       ErrorCode = "DEVICE_DISABLED_LOST"
	TaskNotAssigned          ErrorCode = "TASK_NOT_ASSIGNED"
	TaskAlreadyClaimed       ErrorCode = "TASK_ALREADY_CLAIMED"
	TaskStateConflict        ErrorCode = "TASK_STATE_CONFLICT"
	NodeRequiredPhotoMissing ErrorCode = "NODE_REQUIRED_PHOTO_MISSING"
	NodeRequiredTextMissing  ErrorCode = "NODE_REQUIRED_TEXT_MISSING"
	AttachmentNotUploaded    ErrorCode = "ATTACHMENT_NOT_UPLOADED"
	AttachmentOrphaned       ErrorCode = "ATTACHMENT_ORPHANED"
	IdempotencyConflict      ErrorCode = "IDEMPOTENCY_CONFLICT"
	RateLimited              ErrorCode = "RATE_LIMITED"
	ValidationFailed         ErrorCode = "VALIDATION_FAILED"
	ResourceNotFound         ErrorCode = "RESOURCE_NOT_FOUND"
	InternalError            ErrorCode = "INTERNAL_ERROR"
)

type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Retryable  bool      `json:"retryable"`
	UserAction string    `json:"user_action"`
	HTTPStatus int       `json:"-"`
}

// Error 让 APIError 满足 error 接口，返回稳定错误码。
func (e APIError) Error() string { return string(e.Code) }

// New 根据错误码创建标准 API 错误，并允许调用方覆盖展示消息。
func New(code ErrorCode, message string) APIError {
	e := catalog[code]
	e.Message = message
	if e.Message == "" {
		e.Message = string(code)
	}
	return e
}

var catalog = map[ErrorCode]APIError{
	AuthTokenExpired:         {Code: AuthTokenExpired, HTTPStatus: http.StatusUnauthorized, Retryable: true, UserAction: "refresh_or_relogin"},
	AuthForbidden:            {Code: AuthForbidden, HTTPStatus: http.StatusForbidden, Retryable: false, UserAction: "contact_admin"},
	DeviceRevoked:            {Code: DeviceRevoked, HTTPStatus: http.StatusUnauthorized, Retryable: false, UserAction: "rebind_device"},
	DeviceDisabledLost:       {Code: DeviceDisabledLost, HTTPStatus: http.StatusForbidden, Retryable: false, UserAction: "contact_admin"},
	TaskNotAssigned:          {Code: TaskNotAssigned, HTTPStatus: http.StatusForbidden, Retryable: false, UserAction: "refresh_task_list"},
	TaskAlreadyClaimed:       {Code: TaskAlreadyClaimed, HTTPStatus: http.StatusConflict, Retryable: false, UserAction: "refresh_task_list"},
	TaskStateConflict:        {Code: TaskStateConflict, HTTPStatus: http.StatusConflict, Retryable: true, UserAction: "refresh_detail_and_retry"},
	NodeRequiredPhotoMissing: {Code: NodeRequiredPhotoMissing, HTTPStatus: http.StatusUnprocessableEntity, Retryable: false, UserAction: "take_required_photo"},
	NodeRequiredTextMissing:  {Code: NodeRequiredTextMissing, HTTPStatus: http.StatusUnprocessableEntity, Retryable: false, UserAction: "add_required_note"},
	AttachmentNotUploaded:    {Code: AttachmentNotUploaded, HTTPStatus: http.StatusUnprocessableEntity, Retryable: true, UserAction: "retry_upload"},
	AttachmentOrphaned:       {Code: AttachmentOrphaned, HTTPStatus: http.StatusGone, Retryable: true, UserAction: "upload_again"},
	IdempotencyConflict:      {Code: IdempotencyConflict, HTTPStatus: http.StatusConflict, Retryable: true, UserAction: "retry_same_key_or_refresh"},
	RateLimited:              {Code: RateLimited, HTTPStatus: http.StatusTooManyRequests, Retryable: true, UserAction: "retry_later"},
	ValidationFailed:         {Code: ValidationFailed, HTTPStatus: http.StatusUnprocessableEntity, Retryable: false, UserAction: "fix_input"},
	ResourceNotFound:         {Code: ResourceNotFound, HTTPStatus: http.StatusNotFound, Retryable: false, UserAction: "refresh_list"},
	InternalError:            {Code: InternalError, HTTPStatus: http.StatusInternalServerError, Retryable: true, UserAction: "retry_or_contact_support"},
}
