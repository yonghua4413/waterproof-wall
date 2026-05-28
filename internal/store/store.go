package store

import (
	"context"
	"errors"
	"time"

	"local/captcha-service/internal/config"
	"local/captcha-service/internal/model"
)

var (
	ErrNotFound = errors.New("not found")
	ErrUsed     = errors.New("already used")
)

type Store interface {
	SaveChallenge(ctx context.Context, challenge model.Challenge, ttl time.Duration) error
	GetChallenge(ctx context.Context, id string) (model.Challenge, error)
	UpdateChallenge(ctx context.Context, challenge model.Challenge, ttl time.Duration) error
	DeleteChallenge(ctx context.Context, id string) error
	SaveTicket(ctx context.Context, ticket model.TicketRecord, ttl time.Duration) error
	ConsumeTicket(ctx context.Context, id string) (model.TicketRecord, error)
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

func New(ctx context.Context, cfg config.Config) (Store, func(), error) {
	if cfg.Store == "redis" {
		return NewRedis(ctx, cfg)
	}
	return NewMemory(), func() {}, nil
}
