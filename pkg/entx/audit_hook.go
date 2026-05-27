package entx

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/teamsillybees/initra/pkg/idgen"
)

// IDGenerator 定义雪花 ID 生成器的最小能力。
type IDGenerator interface {
	NextID() idgen.ID
}

// OperatorFunc 从 context 中解析当前操作人 ID。
type OperatorFunc func(ctx context.Context) (idgen.ID, bool)

// AuditHookOptions 描述审计自动填充 hook 的运行时依赖。
type AuditHookOptions struct {
	IDGen    IDGenerator
	Now      func() time.Time
	Operator OperatorFunc
}

type int64IDMutation interface {
	ID() (idgen.ID, bool)
	SetID(idgen.ID)
}

// AuditHook 为 Ent create/update mutation 自动填充 ID 与审计字段。
func AuditHook(options AuditHookOptions) ent.Hook {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	operator := options.Operator
	if operator == nil {
		operator = OperatorIDFromContext
	}

	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			switch {
			case mutation.Op().Is(ent.OpCreate):
				if options.IDGen != nil {
					if err := setIDIfMissing(mutation, options.IDGen); err != nil {
						return nil, err
					}
				}
				current := now()
				if err := setFieldIfExists(mutation, FieldCreatedAt, current); err != nil {
					return nil, err
				}
				if err := setFieldIfExists(mutation, FieldUpdatedAt, current); err != nil {
					return nil, err
				}
				if operatorID, ok := operator(ctx); ok {
					if err := setFieldIfExists(mutation, FieldCreatedBy, operatorID); err != nil {
						return nil, err
					}
					if err := setFieldIfExists(mutation, FieldUpdatedBy, operatorID); err != nil {
						return nil, err
					}
				}
			case mutation.Op().Is(ent.OpUpdate | ent.OpUpdateOne):
				current := now()
				if err := setFieldIfExists(mutation, FieldUpdatedAt, current); err != nil {
					return nil, err
				}
				if operatorID, ok := operator(ctx); ok {
					if err := setFieldIfExists(mutation, FieldUpdatedBy, operatorID); err != nil {
						return nil, err
					}
				}
			}

			return next.Mutate(ctx, mutation)
		})
	}
}

func setIDIfMissing(mutation ent.Mutation, generator IDGenerator) error {
	if typed, ok := mutation.(int64IDMutation); ok {
		if _, exists := typed.ID(); exists {
			return nil
		}
		typed.SetID(generator.NextID())
		return nil
	}
	if _, exists := mutation.Field(FieldID); exists {
		return nil
	}
	return setFieldIfExists(mutation, FieldID, generator.NextID())
}
