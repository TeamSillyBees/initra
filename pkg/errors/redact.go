package apperrors

import (
	"regexp"
	"strings"
)

const redactedValue = "[REDACTED]"

var (
	sensitiveKeyParts = []string{
		"password",
		"passwd",
		"pwd",
		"token",
		"secret",
		"credential",
		"authorization",
		"accesskey",
		"secretkey",
		"privatekey",
		"session",
		"phone",
		"mobile",
		"idcard",
		"identitycard",
		"dsn",
		"sql",
		"query",
		"objectkey",
		"osskey",
		"coskey",
		"s3key",
		"responsebody",
		"rawresponse",
	}
	sensitiveTextPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd|token|secret|authorization|access[_-]?key|secret[_-]?key|dsn)\s*[:=]\s*[^,\s;&]+`),
	}
)

// SanitizeMap 复制并脱敏结构化字段，可用于 HTTP 响应 details 和日志 attrs。
func SanitizeMap(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	sanitized := make(map[string]any, len(fields))
	for key, value := range fields {
		sanitized[key] = SanitizeValue(key, value)
	}
	return sanitized
}

// SanitizeValue 按字段名和值类型递归脱敏敏感信息。
func SanitizeValue(key string, value any) any {
	if isSensitiveKey(key) {
		return redactedValue
	}
	switch typed := value.(type) {
	case map[string]any:
		return SanitizeMap(typed)
	case map[string]string:
		sanitized := make(map[string]string, len(typed))
		for itemKey, itemValue := range typed {
			if isSensitiveKey(itemKey) {
				sanitized[itemKey] = redactedValue
				continue
			}
			sanitized[itemKey] = SanitizeText(itemValue)
		}
		return sanitized
	case []any:
		sanitized := make([]any, 0, len(typed))
		for _, item := range typed {
			sanitized = append(sanitized, SanitizeValue("", item))
		}
		return sanitized
	case []string:
		sanitized := make([]string, 0, len(typed))
		for _, item := range typed {
			sanitized = append(sanitized, SanitizeText(item))
		}
		return sanitized
	case string:
		return SanitizeText(typed)
	default:
		return value
	}
}

// SanitizeText 脱敏常见 key=value 或 key:value 形式的敏感片段。
func SanitizeText(text string) string {
	sanitized := text
	for _, pattern := range sensitiveTextPatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, redactAssignment)
	}
	return sanitized
}

func redactAssignment(text string) string {
	for _, separator := range []string{":", "="} {
		if index := strings.Index(text, separator); index >= 0 {
			return strings.TrimSpace(text[:index]) + separator + redactedValue
		}
	}
	return redactedValue
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.NewReplacer("_", "", "-", "", ".", "", " ", "").Replace(normalized)
	for _, part := range sensitiveKeyParts {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}
