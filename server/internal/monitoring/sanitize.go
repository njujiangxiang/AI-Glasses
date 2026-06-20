package monitoring

import (
	"regexp"
	"strings"
)

var (
	ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	redactors   = []*regexp.Regexp{
		regexp.MustCompile(`(?i)authorization\s*[:=]\s*bearer\s+[^\s,;]+`),
		regexp.MustCompile(`(?i)bearer\s+[^\s,;]+`),
		regexp.MustCompile(`(?i)(access_token|token|password|secret|jwt)\s*[:=]\s*[^\s,;&]+`),
		regexp.MustCompile(`(?i)"(access_token|token|password|secret|jwt)"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)cookie\s*[:=]\s*[^\r\n]+`),
	}
)

func Sanitize(message string) string {
	message = StripANSI(message)
	for _, redactor := range redactors {
		message = redactor.ReplaceAllStringFunc(message, redactMatch)
	}
	return strings.TrimSpace(message)
}

func StripANSI(message string) string {
	return ansiPattern.ReplaceAllString(message, "")
}

func redactMatch(match string) string {
	lower := strings.ToLower(match)
	switch {
	case strings.HasPrefix(lower, "authorization"):
		return "Authorization: [REDACTED]"
	case strings.HasPrefix(lower, "bearer"):
		return "Bearer [REDACTED]"
	case strings.HasPrefix(lower, "cookie"):
		return "Cookie: [REDACTED]"
	case strings.HasPrefix(strings.TrimSpace(match), "\""):
		parts := strings.SplitN(match, ":", 2)
		return parts[0] + `:"[REDACTED]"`
	default:
		parts := strings.FieldsFunc(match, func(r rune) bool { return r == ':' || r == '=' })
		if len(parts) > 0 {
			separator := "="
			if strings.Contains(match, ":") && !strings.Contains(match, "=") {
				separator = ":"
			}
			return strings.TrimSpace(parts[0]) + separator + "[REDACTED]"
		}
		return "[REDACTED]"
	}
}
