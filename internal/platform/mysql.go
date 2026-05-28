package platform

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"local/captcha-service/internal/config"
)

type MySQL struct {
	db     *sql.DB
	secret []byte
}

func NewMySQL(ctx context.Context, cfg config.Config) (*MySQL, error) {
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci", cfg.DBUser, cfg.DBPassword, cfg.DBAddr)
	root, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return nil, err
	}
	if _, err := root.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+cfg.DBName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		_ = root.Close()
		return nil, err
	}
	_ = root.Close()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci", cfg.DBUser, cfg.DBPassword, cfg.DBAddr, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	store := &MySQL{db: db, secret: cfg.Secret}
	if err := store.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (m *MySQL) RegisterUser(ctx context.Context, email, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	id, err := NewCredential("usr")
	if err != nil {
		return User{}, err
	}
	now := time.Now().UTC()
	_, err = m.db.ExecContext(ctx, `INSERT INTO users (id,email,password_hash,status,created_at) VALUES (?,?,?,?,?)`, id, email, hash, "inactive", now)
	if isDuplicate(err) {
		return User{}, ErrConflict
	}
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, PasswordHash: hash, Status: "inactive", CreatedAt: now}, nil
}

func (m *MySQL) AuthenticateUser(ctx context.Context, email, password string) (User, error) {
	var user User
	err := m.db.QueryRowContext(ctx, `SELECT id,email,password_hash,status,created_at FROM users WHERE email=?`, strings.ToLower(strings.TrimSpace(email))).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, err
	}
	if !CheckPassword(user.PasswordHash, password) {
		return User{}, ErrUnauthorized
	}
	return user, nil
}

func (m *MySQL) GetUser(ctx context.Context, userID string) (User, error) {
	var user User
	err := m.db.QueryRowContext(ctx, `SELECT id,email,password_hash,status,created_at FROM users WHERE id=?`, userID).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (m *MySQL) ActivateUser(ctx context.Context, userID string) (User, error) {
	_, err := m.db.ExecContext(ctx, `UPDATE users SET status='active', updated_at=? WHERE id=?`, time.Now().UTC(), userID)
	if err != nil {
		return User{}, err
	}
	return m.GetUser(ctx, userID)
}

func (m *MySQL) CreateApp(ctx context.Context, userID, name string, domains []string) (App, error) {
	app, err := newApp(userID, name, domains)
	if err != nil {
		return App{}, err
	}
	ciphertext, err := EncryptSecret(m.secret, app.AppSecret)
	if err != nil {
		return App{}, err
	}
	app.SecretCipher = ciphertext
	domainsJSON, _ := json.Marshal(app.Domains)
	_, err = m.db.ExecContext(ctx, `INSERT INTO apps (id,user_id,name,app_key,secret_cipher,domains,status,created_at) VALUES (?,?,?,?,?,?,?,?)`,
		app.ID, app.UserID, app.Name, app.AppKey, app.SecretCipher, domainsJSON, app.Status, app.CreatedAt)
	if err != nil {
		return App{}, err
	}
	return app, nil
}

func (m *MySQL) ListApps(ctx context.Context, userID string) ([]App, error) {
	rows, err := m.db.QueryContext(ctx, `SELECT id,user_id,name,app_key,secret_cipher,domains,status,created_at FROM apps WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []App
	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (m *MySQL) GetApp(ctx context.Context, appID string) (App, error) {
	app, err := scanApp(m.db.QueryRowContext(ctx, `SELECT id,user_id,name,app_key,secret_cipher,domains,status,created_at FROM apps WHERE id=?`, appID))
	if errors.Is(err, sql.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, err
	}
	return app, nil
}

func (m *MySQL) GetAppByKey(ctx context.Context, appKey string) (App, error) {
	app, err := scanApp(m.db.QueryRowContext(ctx, `SELECT id,user_id,name,app_key,secret_cipher,domains,status,created_at FROM apps WHERE app_key=?`, appKey))
	if errors.Is(err, sql.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, err
	}
	app.AppSecret, err = DecryptSecret(m.secret, app.SecretCipher)
	if err != nil {
		log.Printf("GetAppByKey decrypt failed for appKey=%q: %v", appKey, err)
		return App{}, err
	}
	return app, nil
}

func (m *MySQL) UpdateAppDomains(ctx context.Context, userID, appID string, domains []string) (App, error) {
	app, err := m.GetApp(ctx, appID)
	if err != nil {
		return App{}, err
	}
	if app.UserID != userID {
		return App{}, ErrNotFound
	}
	domains = cleanDomains(domains)
	domainsJSON, _ := json.Marshal(domains)
	_, err = m.db.ExecContext(ctx, `UPDATE apps SET domains=?, updated_at=? WHERE id=? AND user_id=?`, domainsJSON, time.Now().UTC(), appID, userID)
	if err != nil {
		return App{}, err
	}
	app.Domains = domains
	return app, nil
}

func (m *MySQL) RotateAppSecret(ctx context.Context, userID, appID string) (App, error) {
	app, err := m.GetApp(ctx, appID)
	if err != nil {
		return App{}, err
	}
	if app.UserID != userID {
		return App{}, ErrNotFound
	}
	secret, err := NewCredential("sec")
	if err != nil {
		return App{}, err
	}
	ciphertext, err := EncryptSecret(m.secret, secret)
	if err != nil {
		return App{}, err
	}
	_, err = m.db.ExecContext(ctx, `UPDATE apps SET secret_cipher=?, updated_at=? WHERE id=? AND user_id=?`, ciphertext, time.Now().UTC(), appID, userID)
	if err != nil {
		return App{}, err
	}
	app.SecretCipher = ciphertext
	app.AppSecret = secret
	return app, nil
}

func (m *MySQL) LogEvent(ctx context.Context, event CaptchaEvent) error {
	if event.ID == "" {
		id, err := NewCredential("evt")
		if err != nil {
			return err
		}
		event.ID = id
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO captcha_events (id, app_id, scene, biz_id, event_type, reason, ip, event_date, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		event.ID, event.AppID, event.Scene, event.BizID, string(event.EventType), event.Reason, event.IP, event.CreatedAt.Format("2006-01-02"), event.CreatedAt)
	return err
}

func (m *MySQL) GetDailyUsage(ctx context.Context, appID, date string) (DailyUsage, error) {
	var usage DailyUsage
	err := m.db.QueryRowContext(ctx,
		`SELECT app_id, COALESCE(SUM(CASE WHEN event_type='challenge' THEN 1 ELSE 0 END),0),
		        COALESCE(SUM(CASE WHEN event_type='verify' THEN 1 ELSE 0 END),0),
		        COALESCE(SUM(CASE WHEN event_type='pass' THEN 1 ELSE 0 END),0),
		        COALESCE(SUM(CASE WHEN event_type='fail' THEN 1 ELSE 0 END),0),
		        COALESCE(SUM(CASE WHEN event_type='ticket_check' THEN 1 ELSE 0 END),0)
		 FROM captcha_events WHERE app_id=? AND event_date=? GROUP BY app_id`,
		appID, date).Scan(&usage.AppID, &usage.Challenges, &usage.Verifies, &usage.Passes, &usage.Fails, &usage.Tickets)
	if errors.Is(err, sql.ErrNoRows) {
		return DailyUsage{AppID: appID, Date: date}, nil
	}
	usage.Date = date
	return usage, err
}

func (m *MySQL) Close() error {
	return m.db.Close()
}

func (m *MySQL) init(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(80) PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			status VARCHAR(24) NOT NULL DEFAULT 'inactive',
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS apps (
			id VARCHAR(80) PRIMARY KEY,
			user_id VARCHAR(80) NOT NULL,
			name VARCHAR(120) NOT NULL,
			app_key VARCHAR(80) NOT NULL UNIQUE,
			secret_cipher TEXT NOT NULL,
			domains JSON NOT NULL,
			status VARCHAR(24) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NULL,
			INDEX idx_apps_user_id (user_id),
			CONSTRAINT fk_apps_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS captcha_events (
			id VARCHAR(80) PRIMARY KEY,
			app_id VARCHAR(80) NOT NULL,
			scene VARCHAR(64) NOT NULL DEFAULT '',
			biz_id VARCHAR(128) NOT NULL DEFAULT '',
			event_type VARCHAR(32) NOT NULL,
			reason VARCHAR(128) NOT NULL DEFAULT '',
			ip VARCHAR(45) NOT NULL DEFAULT '',
			event_date DATE NOT NULL,
			created_at DATETIME(6) NOT NULL,
			INDEX idx_events_app_date (app_id, event_date),
			INDEX idx_events_app_type (app_id, event_type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, stmt := range stmts {
		if _, err := m.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	// migrate: add status column to users table
	var colCount int
	m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='users' AND COLUMN_NAME='status'`).Scan(&colCount)
	if colCount == 0 {
		if _, err := m.db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN status VARCHAR(24) NOT NULL DEFAULT 'inactive' AFTER password_hash`); err != nil {
			return err
		}
	}
	return m.db.PingContext(ctx)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanApp(row scanner) (App, error) {
	var app App
	var domainsJSON []byte
	err := row.Scan(&app.ID, &app.UserID, &app.Name, &app.AppKey, &app.SecretCipher, &domainsJSON, &app.Status, &app.CreatedAt)
	if err != nil {
		return App{}, err
	}
	_ = json.Unmarshal(domainsJSON, &app.Domains)
	return app, nil
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Duplicate entry")
}
