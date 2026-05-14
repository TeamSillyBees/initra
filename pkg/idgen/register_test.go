package idgen

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

func TestRegisterProvidesGenerator(t *testing.T) {
	injector := do.New()
	Register(injector, 1)

	generator := do.MustInvoke[*Generator](injector)

	require.Positive(t, generator.NextID())
}
