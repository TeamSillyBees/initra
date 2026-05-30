package httpclient

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/logx"
)

func TestRegisterProvidesFactory(t *testing.T) {
	injector := do.New()
	do.ProvideValue(injector, logx.NewNop())
	Register(injector, Config{Enabled: false})

	factory := do.MustInvoke[Factory](injector)

	require.NotNil(t, factory)
}

func TestRegisterProvidesNamedServiceClient(t *testing.T) {
	injector := do.New()
	do.ProvideValue(injector, logx.NewNop())
	Register(injector, Config{
		Enabled: true,
		Services: map[string]ServiceConfig{
			"httpbingo": {BaseURL: "https://httpbingo.org"},
		},
	})

	client := do.MustInvokeNamed[*Client](injector, ClientName("httpbingo"))
	factory := do.MustInvoke[Factory](injector)
	sameClient, err := factory.Get("httpbingo")

	require.NoError(t, err)
	require.Same(t, sameClient, client)
}
