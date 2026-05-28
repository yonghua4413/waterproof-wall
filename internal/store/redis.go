package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"local/captcha-service/internal/config"
	"local/captcha-service/internal/model"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, cfg config.Config) (*Redis, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, func() {}, err
	}
	return &Redis{client: client}, func() { _ = client.Close() }, nil
}

func (r *Redis) SaveChallenge(ctx context.Context, challenge model.Challenge, ttl time.Duration) error {
	return r.setJSON(ctx, "captcha:challenge:"+challenge.ID, challenge, ttl)
}

func (r *Redis) GetChallenge(ctx context.Context, id string) (model.Challenge, error) {
	var challenge model.Challenge
	err := r.getJSON(ctx, "captcha:challenge:"+id, &challenge)
	return challenge, err
}

func (r *Redis) UpdateChallenge(ctx context.Context, challenge model.Challenge, ttl time.Duration) error {
	return r.SaveChallenge(ctx, challenge, ttl)
}

func (r *Redis) DeleteChallenge(ctx context.Context, id string) error {
	return r.client.Del(ctx, "captcha:challenge:"+id).Err()
}

func (r *Redis) SaveTicket(ctx context.Context, ticket model.TicketRecord, ttl time.Duration) error {
	return r.setJSON(ctx, "captcha:ticket:"+ticket.ID, ticket, ttl)
}

func (r *Redis) ConsumeTicket(ctx context.Context, id string) (model.TicketRecord, error) {
	key := "captcha:ticket:" + id
	for {
		var out model.TicketRecord
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			if err := getJSONTx(ctx, tx, key, &out); err != nil {
				return err
			}
			if out.Used {
				return ErrUsed
			}
			out.Used = true
			ttl := time.Until(out.ExpiresAt)
			if ttl <= 0 {
				return ErrNotFound
			}
			b, err := json.Marshal(out)
			if err != nil {
				return err
			}
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, key, b, ttl)
				return nil
			})
			return err
		}, key)
		if err == redis.TxFailedErr {
			continue
		}
		if err != nil {
			return model.TicketRecord{}, err
		}
		return out, nil
	}
}

func (r *Redis) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	redisKey := "captcha:limit:" + key
	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := r.client.Expire(ctx, redisKey, window).Err(); err != nil {
			return false, err
		}
	}
	return count <= int64(limit), nil
}

func (r *Redis) setJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

func (r *Redis) getJSON(ctx context.Context, key string, out any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func getJSONTx(ctx context.Context, tx *redis.Tx, key string, out any) error {
	data, err := tx.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
