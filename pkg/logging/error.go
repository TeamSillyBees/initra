package logging

import (
	"errors"
	"fmt"

	"github.com/samber/oops"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"go.uber.org/zap"
)

// ErrorFields 将平台错误和 oops 错误链展开为适合 zap 的结构化字段。
func ErrorFields(err error) []zap.Field {
	if err == nil {
		return nil
	}

	fields := []zap.Field{
		zap.Error(err),
		zap.String("error_type", fmt.Sprintf("%T", err)),
	}

	if appErr := apperrors.From(err); appErr != nil {
		fields = append(fields,
			zap.String("error_code", string(appErr.Code)),
			zap.String("error_message", appErr.Message),
			zap.Int("error_status", appErr.Status),
		)
		if len(appErr.Details) > 0 {
			fields = append(fields, zap.Any("error_details", apperrors.SanitizeMap(appErr.Details)))
		}
	}

	if cause := rootCause(err); cause != nil && cause != err {
		fields = append(fields, zap.String("error_cause", apperrors.SanitizeText(cause.Error())))
	}

	if oopsErr, ok := oops.AsOops(err); ok {
		if domain := oopsErr.Domain(); domain != "" {
			fields = append(fields, zap.String("error_domain", domain))
		}
		if hint := oopsErr.Hint(); hint != "" {
			fields = append(fields, zap.String("error_hint", hint))
		}
		if trace := oopsErr.Trace(); trace != "" {
			fields = append(fields, zap.String("error_trace", trace))
		}
		if attrs := oopsErr.Context(); len(attrs) > 0 {
			fields = append(fields, zap.Any("error_attrs", apperrors.SanitizeMap(attrs)))
		}
		if stacktrace := oopsErr.Stacktrace(); stacktrace != "" {
			fields = append(fields, zap.String("error_stacktrace", stacktrace))
		}
	}

	return fields
}

func rootCause(err error) error {
	for err != nil {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
	return nil
}
