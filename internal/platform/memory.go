package platform

import (
	"context"
	"strings"
	"sync"
	"time"

	"local/captcha-service/internal/secure"
)

type Memory struct {
	mu       sync.Mutex
	users    map[string]User
	byEmail  map[string]string
	apps     map[string]App
	byAppKey map[string]string
	events   []CaptchaEvent
}

func NewMemory() *Memory {
	return &Memory{
		users:    map[string]User{},
		byEmail:  map[string]string{},
		apps:     map[string]App{},
		byAppKey: map[string]string{},
		events:   []CaptchaEvent{},
	}
}

func (m *Memory) RegisterUser(_ context.Context, email, password string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	email = strings.ToLower(strings.TrimSpace(email))
	if _, ok := m.byEmail[email]; ok {
		return User{}, ErrConflict
	}
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	id, err := NewCredential("usr")
	if err != nil {
		return User{}, err
	}
	user := User{ID: id, Email: email, PasswordHash: hash, Status: "inactive", CreatedAt: time.Now().UTC()}
	m.users[id] = user
	m.byEmail[email] = id
	return user, nil
}

func (m *Memory) AuthenticateUser(_ context.Context, email, password string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, ErrUnauthorized
	}
	user := m.users[id]
	if !CheckPassword(user.PasswordHash, password) {
		return User{}, ErrUnauthorized
	}
	return user, nil
}

func (m *Memory) GetUser(_ context.Context, userID string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.users[userID]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (m *Memory) ActivateUser(_ context.Context, userID string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.users[userID]
	if !ok {
		return User{}, ErrNotFound
	}
	user.Status = "active"
	m.users[userID] = user
	return user, nil
}

func (m *Memory) CreateApp(_ context.Context, userID, name string, domains []string) (App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	app, err := newApp(userID, name, domains)
	if err != nil {
		return App{}, err
	}
	m.apps[app.ID] = app
	m.byAppKey[app.AppKey] = app.ID
	return app, nil
}

func (m *Memory) ListApps(_ context.Context, userID string) ([]App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []App
	for _, app := range m.apps {
		if app.UserID == userID {
			out = append(out, app)
		}
	}
	return out, nil
}

func (m *Memory) GetApp(_ context.Context, appID string) (App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	app, ok := m.apps[appID]
	if !ok {
		return App{}, ErrNotFound
	}
	return app, nil
}

func (m *Memory) GetAppByKey(_ context.Context, appKey string) (App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byAppKey[appKey]
	if !ok {
		return App{}, ErrNotFound
	}
	return m.apps[id], nil
}

func (m *Memory) UpdateAppDomains(_ context.Context, userID, appID string, domains []string) (App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	app, ok := m.apps[appID]
	if !ok || app.UserID != userID {
		return App{}, ErrNotFound
	}
	app.Domains = cleanDomains(domains)
	m.apps[appID] = app
	return app, nil
}

func (m *Memory) RotateAppSecret(_ context.Context, userID, appID string) (App, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	app, ok := m.apps[appID]
	if !ok || app.UserID != userID {
		return App{}, ErrNotFound
	}
	secret, err := NewCredential("sec")
	if err != nil {
		return App{}, err
	}
	app.AppSecret = secret
	m.apps[appID] = app
	return app, nil
}

func (m *Memory) LogEvent(_ context.Context, event CaptchaEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if event.ID == "" {
		id, err := secure.ID(12)
		if err != nil {
			return err
		}
		event.ID = id
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	m.events = append(m.events, event)
	return nil
}

func (m *Memory) GetDailyUsage(_ context.Context, appID, date string) (DailyUsage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	usage := DailyUsage{AppID: appID, Date: date}
	for _, e := range m.events {
		if e.AppID != appID {
			continue
		}
		if e.CreatedAt.Format("2006-01-02") != date {
			continue
		}
		switch e.EventType {
		case EventChallenge:
			usage.Challenges++
		case EventVerify:
			usage.Verifies++
		case EventPass:
			usage.Passes++
		case EventFail:
			usage.Fails++
		case EventTicket:
			usage.Tickets++
		}
	}
	return usage, nil
}

func (m *Memory) Close() error {
	return nil
}

func newApp(userID, name string, domains []string) (App, error) {
	appID, err := NewCredential("app")
	if err != nil {
		return App{}, err
	}
	key, err := NewCredential("key")
	if err != nil {
		return App{}, err
	}
	secret, err := NewCredential("sec")
	if err != nil {
		return App{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Untitled App"
	}
	return App{
		ID:        appID,
		UserID:    userID,
		Name:      name,
		AppKey:    key,
		AppSecret: secret,
		Domains:   cleanDomains(domains),
		Status:    "active",
		CreatedAt: time.Now().UTC(),
	}, nil
}

func cleanDomains(domains []string) []string {
	out := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.TrimSpace(strings.ToLower(domain))
		domain = strings.TrimPrefix(domain, "https://")
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.TrimSuffix(domain, "/")
		if domain != "" {
			out = append(out, domain)
		}
	}
	return out
}
