package entx

import (
	"strings"

	"entgo.io/ent"
)

// 审计与软删除字段名是 Ent hook 和业务 schema 之间的稳定契约。
const (
	FieldID        = "id"
	FieldCreatedAt = "created_at"
	FieldUpdatedAt = "updated_at"
	FieldCreatedBy = "created_by"
	FieldUpdatedBy = "updated_by"
	FieldDeletedAt = "deleted_at"
)

func setFieldIfExists(mutation ent.Mutation, name string, value ent.Value) error {
	if err := mutation.SetField(name, value); err != nil {
		if isUnknownFieldError(err) {
			return nil
		}
		return err
	}
	return nil
}

func isUnknownFieldError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "unknown") && strings.Contains(message, "field") {
		return true
	}
	if strings.Contains(message, "not found") && strings.Contains(message, "field") {
		return true
	}
	return false
}
