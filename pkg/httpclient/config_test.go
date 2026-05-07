package httpclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidateAndSafeForLog(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Services: map[string]ServiceConfig{
			"user_center": {
				BaseURL: "https://api.example.com",
				Headers: map[string]string{
					"X-App-Id":  "initra",
					"X-API-Key": "secret-key",
				},
				Auth: AuthConfig{
					Type:  AuthTypeBearer,
					Token: "token-value",
				},
				Retry: RetryConfig{
					Enabled: true,
				},
			},
		},
	}

	require.NoError(t, cfg.Validate())

	safe := cfg.SafeForLog()
	services := safe["services"].(map[string]any)
	service := services["user_center"].(map[string]any)
	headers := service["headers"].(map[string]string)
	auth := service["auth"].(map[string]any)
	require.Equal(t, maskedValue, headers["X-API-Key"])
	require.Equal(t, maskedValue, auth["token"])
	require.Equal(t, "initra", headers["X-App-Id"])
}

func TestConfigValidateRejectsUnsupportedStandardAPIInV1(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Services: map[string]ServiceConfig{
			"user_center": {
				BaseURL: "https://api.example.com",
				Response: ResponseConfig{
					Type: ResponseTypeStandardAPI,
				},
			},
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnsupported))
	require.ErrorContains(t, err, "V2")
}

func TestConfigValidateAuthRequirements(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Services: map[string]ServiceConfig{
			"user_center": {
				BaseURL: "https://api.example.com",
				Auth: AuthConfig{
					Type: AuthTypeAPIKey,
				},
			},
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	require.ErrorContains(t, err, "auth.header")
}
