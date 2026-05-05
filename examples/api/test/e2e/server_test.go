package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/server"
	"go.uber.org/zap"
)

// TestServerObservabilityRoutes 覆盖 API 骨架默认提供的公开观测接口。
func TestServerObservabilityRoutes(t *testing.T) {
	jwtManager, err := platformauth.NewJWTManager(platformauth.JWTConfig{
		Issuer:          "initra",
		Secret:          "e2e-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})
	require.NoError(t, err)

	modelPath, policyPath := writeCasbinFiles(t)
	enforcer, err := platformauth.NewEnforcer(modelPath, policyPath)
	require.NoError(t, err)

	app, err := server.NewApp(server.Options{
		Title:   "initra",
		Version: "test",
		Env:     "test",
	}, zap.NewNop(), jwtManager, enforcer)
	require.NoError(t, err)

	observability.NewModule(observability.BuildInfo{
		Version:   "test",
		Commit:    "abc123",
		BuildTime: "2026-04-21T00:00:00Z",
	}).Register(app.API, app.Registry)

	for _, path := range []string{"/health", "/ready", "/version"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		app.Engine.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, path)
	}
}

func writeCasbinFiles(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "rbac_model.conf")
	policyPath := filepath.Join(dir, "rbac_policy.csv")

	model := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	require.NoError(t, os.WriteFile(modelPath, []byte(strings.TrimSpace(model)), 0o600))
	require.NoError(t, os.WriteFile(policyPath, []byte(""), 0o600))
	return modelPath, policyPath
}
