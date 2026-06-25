package glass_api

import (
	"time"

	"aiglasses/server/internal/platform/database"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     uint64 `json:"device_id"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type deviceRegisterRequest struct {
	SerialNo        string `json:"serial_no"`
	DeviceSN        string `json:"device_sn"`
	DeviceModel     string `json:"device_model"`
	DeviceName      string `json:"device_name"`
	FirmwareVersion string `json:"firmware_version"`
}

type deviceStatusRequest struct {
	DeviceID       uint64   `json:"device_id"`
	BatteryLevel   int      `json:"battery_level"`
	SignalStrength int      `json:"signal_strength"`
	StorageUsed    int64    `json:"storage_used"`
	StorageTotal   int64    `json:"storage_total"`
	NetworkType    string   `json:"network_type"`
	GPSLat         *float64 `json:"gps_lat"`
	GPSLng         *float64 `json:"gps_lng"`
	IsOnline       bool     `json:"is_online"`
}

type createDefectRequest struct {
	InsID         string   `json:"ins_id"`
	TaskID        string   `json:"task_id"`
	NodeID        string   `json:"node_id"`
	Description   string   `json:"description"`
	ReporterID    string   `json:"reporter_id"`
	Status        string   `json:"status"`
	AttachmentIDs []uint64 `json:"attachment_ids"`
}

type realtimeQueryRequest struct {
	TaskID      string         `json:"task_id"`
	NodeID      string         `json:"node_id"`
	QueryID     string         `json:"query_id"`
	ExtraParams map[string]any `json:"extra_params"`
}

type algorithmInvokeRequest struct {
	AlgorithmID   string         `json:"algorithm_id"`
	TaskID        string         `json:"task_id"`
	NodeID        string         `json:"node_id"`
	AttachmentIDs []uint64       `json:"attachment_ids"`
	InputParams   map[string]any `json:"input_params"`
}

type taskTypeResponse struct {
	TaskTypeID       string `json:"task_type_id"`
	TypeCode         string `json:"type_code"`
	TypeName         string `json:"type_name"`
	TypeDesc         string `json:"type_desc"`
	SupportAlgorithm string `json:"support_algorithm"`
	SupportQuery     string `json:"support_query"`
	SupportMandatory string `json:"support_mandatory"`
}

type attachmentUploadResult struct {
	AttachmentID string              `json:"attachment_id"`
	ObjectKey    string              `json:"object_key"`
	FileName     string              `json:"file_name"`
	SizeBytes    int64               `json:"size_bytes"`
	ContentType  string              `json:"content_type"`
	Attachment   database.Attachment `json:"attachment"`
}

type batchUploadItem struct {
	Index      int                     `json:"index"`
	Success    bool                    `json:"success"`
	Attachment *attachmentUploadResult `json:"attachment,omitempty"`
	Error      string                  `json:"error,omitempty"`
}

func parseTimePtr(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return &parsed, nil
	}
	parsed, err = time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
