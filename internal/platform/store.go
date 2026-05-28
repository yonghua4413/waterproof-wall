package platform

import (
	"context"

	"local/captcha-service/internal/config"
)

func New(ctx context.Context, cfg config.Config) (Store, error) {
	if cfg.PlatformStore == "mysql" {
		return NewMySQL(ctx, cfg)
	}
	return NewMemory(), nil
}
