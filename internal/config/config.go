package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr           string
	Secret         []byte
	Store          string
	PlatformStore  string
	DBAddr         string
	DBUser         string
	DBPassword     string
	DBName         string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	AllowedOrigins []string
	ChallengeLimit int
	VerifyLimit    int

	ChallengeTTL time.Duration
	TicketTTL    time.Duration
	ImageWidth   int
	ImageHeight  int
	PieceSize    int
}

type yamlConfig struct {
	Addr          string `yaml:"addr"`
	Secret        string `yaml:"secret"`
	Store         string `yaml:"store"`
	PlatformStore string `yaml:"platform_store"`
	DB struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"db"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	Limits struct {
		Challenge int `yaml:"challenge"`
		Verify    int `yaml:"verify"`
	} `yaml:"limits"`
	TTL struct {
		Challenge string `yaml:"challenge"`
		Ticket    string `yaml:"ticket"`
	} `yaml:"ttl"`
	Image struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
		Piece  int `yaml:"piece"`
	} `yaml:"image"`
}

func Load(path string) Config {
	var yc yamlConfig
	yc.Secret = "dev-secret-change-me-at-least-32-bytes"
	yc.Addr = ":8088"
	yc.Store = "memory"
	yc.PlatformStore = "memory"
	yc.DB.Addr = "127.0.0.1:3306"
	yc.DB.User = "root"
	yc.DB.Name = "waterproof_wall"
	yc.Redis.Addr = "127.0.0.1:6379"
	yc.Redis.DB = 0
	yc.AllowedOrigins = []string{"*"}
	yc.Limits.Challenge = 60
	yc.Limits.Verify = 120
	yc.TTL.Challenge = "2m"
	yc.TTL.Ticket = "3m"
	yc.Image.Width = 320
	yc.Image.Height = 160
	yc.Image.Piece = 58

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read config %s: %v\n", path, err)
			os.Exit(1)
		}
		if err := yaml.Unmarshal(data, &yc); err != nil {
			fmt.Fprintf(os.Stderr, "parse config %s: %v\n", path, err)
			os.Exit(1)
		}
	}

	challengeTTL, err := time.ParseDuration(yc.TTL.Challenge)
	if err != nil {
		challengeTTL = 2 * time.Minute
	}
	ticketTTL, err := time.ParseDuration(yc.TTL.Ticket)
	if err != nil {
		ticketTTL = 3 * time.Minute
	}

	cfg := Config{
		Addr:           envOr("CAPTCHA_ADDR", yc.Addr),
		Secret:         []byte(envOr("CAPTCHA_SECRET", yc.Secret)),
		Store:          strings.ToLower(envOr("CAPTCHA_STORE", yc.Store)),
		PlatformStore:  strings.ToLower(envOr("CAPTCHA_PLATFORM_STORE", yc.PlatformStore)),
		DBAddr:         envOr("CAPTCHA_DB_ADDR", yc.DB.Addr),
		DBUser:         envOr("CAPTCHA_DB_USER", yc.DB.User),
		DBPassword:     envOr("CAPTCHA_DB_PASSWORD", yc.DB.Password),
		DBName:         envOr("CAPTCHA_DB_NAME", yc.DB.Name),
		RedisAddr:      envOr("CAPTCHA_REDIS_ADDR", yc.Redis.Addr),
		RedisPassword:  envOr("CAPTCHA_REDIS_PASSWORD", yc.Redis.Password),
		RedisDB:        envIntOr("CAPTCHA_REDIS_DB", yc.Redis.DB),
		AllowedOrigins: envSliceOr("CAPTCHA_ALLOWED_ORIGINS", yc.AllowedOrigins),
		ChallengeLimit: envIntOr("CAPTCHA_CHALLENGE_LIMIT", yc.Limits.Challenge),
		VerifyLimit:    envIntOr("CAPTCHA_VERIFY_LIMIT", yc.Limits.Verify),
		ChallengeTTL:   challengeTTL,
		TicketTTL:      ticketTTL,
		ImageWidth:     envIntOr("CAPTCHA_IMAGE_WIDTH", yc.Image.Width),
		ImageHeight:    envIntOr("CAPTCHA_IMAGE_HEIGHT", yc.Image.Height),
		PieceSize:      envIntOr("CAPTCHA_IMAGE_PIECE", yc.Image.Piece),
	}

	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"*"}
	}
	return cfg
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envSliceOr(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	items := strings.Split(v, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func envBoolOr(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}