package httpclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidateAndSafeForLog(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Proxy:   "http://proxy-user:proxy-pass@127.0.0.1:7890",
		Services: map[string]ServiceConfig{
			"user_center": {
				BaseURL: "https://api.example.com",
				Proxy:   "http://service-user:service-pass@127.0.0.1:7891",
				Headers: map[string]string{
					"X-App-Id":  "initra",
					"X-API-Key": "secret-key",
				},
				Properties: map[string]string{
					"app_id":        "initra",
					"client_secret": "secret-value",
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
	properties := service["properties"].(map[string]string)
	auth := service["auth"].(map[string]any)
	require.Equal(t, maskedValue, headers["X-API-Key"])
	require.Equal(t, maskedValue, properties["client_secret"])
	require.Equal(t, maskedValue, auth["token"])
	require.Equal(t, "initra", headers["X-App-Id"])
	require.Equal(t, "initra", properties["app_id"])
	require.Equal(t, "http://proxy-user:"+maskedValue+"@127.0.0.1:7890", safe["proxy"])
	require.Equal(t, "http://service-user:"+maskedValue+"@127.0.0.1:7891", service["proxy"])
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

func TestConfigValidateProxy(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "global proxy",
			cfg: Config{
				Enabled: true,
				Proxy:   "127.0.0.1:7890",
				Services: map[string]ServiceConfig{
					"user_center": {BaseURL: "https://api.example.com"},
				},
			},
		},
		{
			name: "service proxy",
			cfg: Config{
				Enabled: true,
				Services: map[string]ServiceConfig{
					"user_center": {
						BaseURL: "https://api.example.com",
						Proxy:   "ftp://127.0.0.1:7890",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			require.Error(t, err)
			require.True(t, errors.Is(err, ErrInvalidConfig))
			require.ErrorContains(t, err, "proxy")
		})
	}
}
