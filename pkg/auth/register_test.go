package auth

import (
	"testing"
	"time"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

func TestRegisterProvidesAuthComponents(t *testing.T) {
	injector := do.New()
	Register(injector, RegisterOptions{
		AppName:      "initra",
		Env:          "test",
		PasswordCost: 0,
		JWT: JWTConfig{
			Issuer:          "initra",
			Secret:          "register-test-secret",
			AccessTokenTTL:  time.Minute,
			RefreshTokenTTL: time.Hour,
		},
	})

	passwordManager := do.MustInvoke[*BcryptPasswordManager](injector)
	jwtManager := do.MustInvoke[*JWTManager](injector)

	require.NotNil(t, passwordManager)
	require.NotNil(t, jwtManager)
}
