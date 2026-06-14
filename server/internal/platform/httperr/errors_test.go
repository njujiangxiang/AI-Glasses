package httperr

import "testing"

// TestErrorCatalog 验证错误码目录中的 HTTP 状态码和可重试标记。
func TestErrorCatalog(t *testing.T) {
	cases := []struct {
		code      ErrorCode
		status    int
		retryable bool
	}{
		{AuthTokenExpired, 401, true},
		{AuthForbidden, 403, false},
		{TaskAlreadyClaimed, 409, false},
		{NodeRequiredPhotoMissing, 422, false},
		{AttachmentOrphaned, 410, true},
	}
	for _, tc := range cases {
		err := New(tc.code, "")
		if err.HTTPStatus != tc.status || err.Retryable != tc.retryable || err.Code != tc.code {
			t.Fatalf("unexpected catalog entry for %s", tc.code)
		}
	}
}
