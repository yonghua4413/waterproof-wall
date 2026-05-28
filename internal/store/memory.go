package store

import (
	"context"
	"sync"
	"time"

	"local/captcha-service/internal/model"
)

type memoryItem[T any] struct {
	value     T
	expiresAt time.Time
}

type Memory struct {
	mu         sync.Mutex
	challenges map[string]memoryItem[model.Challenge]
	tickets    map[string]memoryItem[model.TicketRecord]
	limits     map[string]memoryItem[int]
}

func NewMemory() *Memory {
	m := &Memory{
		challenges: map[string]memoryItem[model.Challenge]{},
		tickets:    map[string]memoryItem[model.TicketRecord]{},
		limits:     map[string]memoryItem[int]{},
	}
	go m.cleanupLoop()
	return m
}

func (m *Memory) SaveChallenge(_ context.Context, challenge model.Challenge, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.challenges[challenge.ID] = memoryItem[model.Challenge]{value: challenge, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (m *Memory) GetChallenge(_ context.Context, id string) (model.Challenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.challenges[id]
	if !ok || time.Now().After(item.expiresAt) {
		delete(m.challenges, id)
		return model.Challenge{}, ErrNotFound
	}
	return item.value, nil
}

func (m *Memory) UpdateChallenge(_ context.Context, challenge model.Challenge, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.challenges[challenge.ID] = memoryItem[model.Challenge]{value: challenge, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (m *Memory) DeleteChallenge(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.challenges, id)
	return nil
}

func (m *Memory) SaveTicket(_ context.Context, ticket model.TicketRecord, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tickets[ticket.ID] = memoryItem[model.TicketRecord]{value: ticket, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (m *Memory) ConsumeTicket(_ context.Context, id string) (model.TicketRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.tickets[id]
	if !ok || time.Now().After(item.expiresAt) {
		delete(m.tickets, id)
		return model.TicketRecord{}, ErrNotFound
	}
	if item.value.Used {
		return model.TicketRecord{}, ErrUsed
	}
	item.value.Used = true
	m.tickets[id] = item
	return item.value, nil
}

func (m *Memory) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	item, ok := m.limits[key]
	if !ok || now.After(item.expiresAt) {
		m.limits[key] = memoryItem[int]{value: 1, expiresAt: now.Add(window)}
		return true, nil
	}
	item.value++
	m.limits[key] = item
	return item.value <= limit, nil
}

func (m *Memory) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.cleanup()
	}
}

func (m *Memory) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, item := range m.challenges {
		if now.After(item.expiresAt) {
			delete(m.challenges, id)
		}
	}
	for id, item := range m.tickets {
		if now.After(item.expiresAt) {
			delete(m.tickets, id)
		}
	}
	for id, item := range m.limits {
		if now.After(item.expiresAt) {
			delete(m.limits, id)
		}
	}
}
