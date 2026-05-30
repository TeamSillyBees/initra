package logx

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// RedactedValue 是日志脱敏后的统一占位值。
const RedactedValue = "[REDACTED]"

var (
	// defaultSensitiveKeys 是所有项目默认脱敏的字段名集合。
	defaultSensitiveKeys = []string{
		"password",
		"passwd",
		"pwd",
		"token",
		"accesstoken",
		"refreshtoken",
		"authorization",
		"cookie",
		"setcookie",
		"secret",
		"credential",
		"accesskey",
		"secretkey",
		"privatekey",
		"session",
		"phone",
		"mobile",
		"email",
		"idcard",
		"identitycard",
		"dsn",
		"sql",
		"query",
		"body",
		"requestbody",
		"responsebody",
		"rawresponse",
	}
	// assignmentPatterns 匹配字符串中常见的敏感 key=value/key:value 片段。
	assignmentPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd|token|secret|authorization|access[_-]?key|secret[_-]?key|dsn)\s*[:=]\s*[^,\s;&]+`),
	}
)

// IsSensitiveKey 判断字段名是否属于敏感信息。
func IsSensitiveKey(key string, extraFields []string) bool {
	normalized := normalizeKey(key)
	if normalized == "" {
		return false
	}
	for _, candidate := range append(defaultSensitiveKeys, extraFields...) {
		candidate = normalizeKey(candidate)
		if candidate == "" {
			continue
		}
		if normalized == candidate || strings.Contains(normalized, candidate) {
			return true
		}
	}
	return false
}

// RedactValue 按字段名和值类型递归脱敏。
func RedactValue(key string, value any) any {
	if IsSensitiveKey(key, nil) {
		return RedactedValue
	}
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for itemKey, itemValue := range typed {
			result[itemKey] = RedactValue(itemKey, itemValue)
		}
		return result
	case map[string]string:
		result := make(map[string]string, len(typed))
		for itemKey, itemValue := range typed {
			if IsSensitiveKey(itemKey, nil) {
				result[itemKey] = RedactedValue
				continue
			}
			result[itemKey] = RedactText(itemValue)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, RedactValue("", item))
		}
		return result
	case []string:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, RedactText(item))
		}
		return result
	case string:
		return RedactText(typed)
	default:
		return value
	}
}

// RedactFields 对 zap 字段执行集中脱敏。
func RedactFields(fields []Field, cfg RedactConfig) []Field {
	if !cfg.Enabled || len(fields) == 0 {
		return fields
	}
	result := make([]Field, len(fields))
	copy(result, fields)
	for index, field := range result {
		if IsSensitiveKey(field.Key, cfg.Fields) {
			result[index] = zap.String(field.Key, RedactedValue)
			continue
		}
		switch field.Type {
		case zapcore.StringType:
			result[index] = zap.String(field.Key, RedactText(field.String))
		case zapcore.ErrorType:
			if err, ok := field.Interface.(error); ok && err != nil {
				result[index] = zap.String(field.Key, RedactText(err.Error()))
			}
		case zapcore.ReflectType:
			result[index] = zap.Any(field.Key, RedactValue(field.Key, field.Interface))
		}
	}
	return result
}

// RedactText 脱敏字符串中的常见敏感赋值片段。
func RedactText(text string) string {
	result := text
	for _, pattern := range assignmentPatterns {
		result = pattern.ReplaceAllStringFunc(result, redactAssignment)
	}
	return result
}

// redactCore 是包裹 zapcore.Core 的脱敏适配器。
type redactCore struct {
	zapcore.Core
	cfg RedactConfig
}

// With 返回携带已脱敏固定字段的新 core。
func (c *redactCore) With(fields []zapcore.Field) zapcore.Core {
	return &redactCore{Core: c.Core.With(RedactFields(fields, c.cfg)), cfg: c.cfg}
}

// Check 将启用级别的日志条目绑定到当前 core。
func (c *redactCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

// Write 在写入底层 core 前对字段做最终脱敏。
func (c *redactCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return c.Core.Write(entry, RedactFields(fields, c.cfg))
}

// redactMap 对 map 字段做递归脱敏。
func redactMap(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	return RedactValue("", fields).(map[string]any)
}

// redactAssignment 替换单个敏感赋值片段的值部分。
func redactAssignment(text string) string {
	colon := strings.Index(text, ":")
	equal := strings.Index(text, "=")
	switch {
	case equal >= 0 && (colon < 0 || equal < colon):
		return strings.TrimSpace(text[:equal]) + "=" + RedactedValue
	case colon >= 0:
		return strings.TrimSpace(text[:colon]) + ":" + RedactedValue
	default:
		return RedactedValue
	}
}

// normalizeKey 将字段名规整为只含小写字母和数字的形式。
func normalizeKey(key string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(fmt.Sprint(key))) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
