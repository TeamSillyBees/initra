package entx

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"entgo.io/ent"

	"github.com/teamsillybees/initra/pkg/idgen"
)

type fakeMutation struct {
	ent.Mutation
	op         ent.Op
	fields     map[string]ent.Value
	unknown    map[string]bool
	typeErrors map[string]bool
}

func newFakeMutation(op ent.Op) *fakeMutation {
	return &fakeMutation{
		op:     op,
		fields: map[string]ent.Value{},
	}
}

func (m *fakeMutation) Op() ent.Op {
	return m.op
}

func (m *fakeMutation) Field(name string) (ent.Value, bool) {
	value, ok := m.fields[name]
	return value, ok
}

func (m *fakeMutation) SetField(name string, value ent.Value) error {
	if m.unknown[name] {
		return fmt.Errorf("unknown fieldx %q", name)
	}
	if m.typeErrors[name] {
		return fmt.Errorf("unexpected type %T for fieldx %s", value, name)
	}
	m.fields[name] = value
	return nil
}

func TestAuditHookSetsOperatorFieldsOnCreate(t *testing.T) {
	mutation := newFakeMutation(ent.OpCreate)
	hook := AuditHook(AuditHookOptions{
		Operator: func(ctx context.Context) (idgen.ID, bool) {
			return idgen.New(9001), true
		},
	})

	_, err := hook(ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		return "ok", nil
	})).Mutate(context.Background(), mutation)

	if err != nil {
		t.Fatalf("Mutate() error = %v", err)
	}
	if got := mutation.fields[FieldCreatedBy]; got != idgen.New(9001) {
		t.Fatalf("created_by = %v, want 9001", got)
	}
	if got := mutation.fields[FieldUpdatedBy]; got != idgen.New(9001) {
		t.Fatalf("updated_by = %v, want 9001", got)
	}
}

func TestAuditHookSetsOnlyUpdatedByOnUpdate(t *testing.T) {
	mutation := newFakeMutation(ent.OpUpdateOne)
	hook := AuditHook(AuditHookOptions{
		Operator: func(ctx context.Context) (idgen.ID, bool) {
			return idgen.New(9002), true
		},
	})

	_, err := hook(ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		return "ok", nil
	})).Mutate(context.Background(), mutation)

	if err != nil {
		t.Fatalf("Mutate() error = %v", err)
	}
	if _, ok := mutation.fields[FieldCreatedBy]; ok {
		t.Fatalf("created_by should not be set on update")
	}
	if got := mutation.fields[FieldUpdatedBy]; got != idgen.New(9002) {
		t.Fatalf("updated_by = %v, want 9002", got)
	}
}

func TestAuditHookIgnoresUnknownFields(t *testing.T) {
	mutation := newFakeMutation(ent.OpCreate)
	mutation.unknown = map[string]bool{
		FieldCreatedBy: true,
		FieldUpdatedBy: true,
	}
	hook := AuditHook(AuditHookOptions{
		Operator: func(ctx context.Context) (idgen.ID, bool) {
			return idgen.New(9001), true
		},
	})

	_, err := hook(ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		return "ok", nil
	})).Mutate(context.Background(), mutation)

	if err != nil {
		t.Fatalf("Mutate() error = %v", err)
	}
	if _, ok := mutation.fields[FieldCreatedBy]; ok {
		t.Fatalf("created_by should be ignored when the mutation does not know that fieldx")
	}
}

func TestAuditHookReturnsTypeErrors(t *testing.T) {
	mutation := newFakeMutation(ent.OpUpdate)
	mutation.typeErrors = map[string]bool{
		FieldUpdatedBy: true,
	}
	hook := AuditHook(AuditHookOptions{
		Operator: func(ctx context.Context) (idgen.ID, bool) {
			return idgen.New(9002), true
		},
	})

	_, err := hook(ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		return "ok", nil
	})).Mutate(context.Background(), mutation)

	if err == nil {
		t.Fatalf("Mutate() error = nil, want type error")
	}
}

func TestRejectDeleteHookRejectsPhysicalDeletes(t *testing.T) {
	hook := RejectDeleteHook()

	_, err := hook(ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		return "deleted", nil
	})).Mutate(context.Background(), newFakeMutation(ent.OpDeleteOne))

	if err == nil {
		t.Fatalf("Mutate() error = nil, want delete rejection")
	}
	if !errors.Is(err, ErrPhysicalDeleteRejected) {
		t.Fatalf("Mutate() error = %v, want ErrPhysicalDeleteRejected", err)
	}
}
