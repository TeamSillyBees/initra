package awss3

import (
	"context"

	"github.com/teamsillybees/initra/pkg/storage"
	"github.com/teamsillybees/initra/pkg/storage/s3compat"
)

// NewSTS 创建 AWS STS 临时授权服务。
func NewSTS(ctx context.Context, cfg storage.Config) (*s3compat.STS, error) {
	if cfg.Provider == "" {
		cfg.Provider = storage.ProviderAWSS3
	}
	cfg = storage.ConfigWithDefaults(cfg)
	return s3compat.NewSTS(ctx, cfg)
}
