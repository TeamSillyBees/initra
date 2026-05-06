package data

import (
	"context"
	"database/sql"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/teamsillybees/initra/examples/api/internal/ent"
	_ "github.com/teamsillybees/initra/examples/api/internal/ent/runtime"
	"github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/entx"
	"github.com/teamsillybees/initra/pkg/idgen"
)

// NewEntClient 基于既有 sql.DB 创建 Ent Client，并注册统一运行时 Hook。
func NewEntClient(db *sql.DB, generator *idgen.Generator) *ent.Client {
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	client.Use(entx.AuditHook(entx.AuditHookOptions{
		IDGen: generator,
		Now:   time.Now,
		Operator: func(ctx context.Context) (int64, bool) {
			if id, ok := entx.OperatorIDFromContext(ctx); ok {
				return id, true
			}
			principal, ok := auth.PrincipalFromContext(ctx)
			if !ok {
				return 0, false
			}
			return principal.UserID, true
		},
	}))
	client.Use(entx.RejectDeleteHook())

	return client
}
