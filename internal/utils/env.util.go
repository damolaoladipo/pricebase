package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var current *ENV

type ENV struct {
	IsDev       bool
	IsStaging   bool
	IsProd      bool
	Environment string

	APP_NAME    string
	APP_VERSION string
	GO_ENV      string
	APP_ENV     string
	APP_PORT    int
	API_ROUTE   string

	APP_CHANNELS string
	AUTH_TYPE    string
	CORS_ORIGINS []string

	JWT_SECRET        string
	JWT_EXPIRES       string
	JWT_COOKIE_EXPIRE string

	POSTGRES_URL      string
	POSTGRES_TEST_URL string

	EMAIL_ENABLED  bool
	RESEND_API_KEY string
	EMAIL_FROM     string
	EMAIL_REPLY_TO string

	LOGS_PATH             string
	LOGTAIL_MONITOR_TOKEN string
	LOGTAIL_AUDIT_TOKEN   string
	LOGTAIL_INGESTION_URL string

	SUPERADMIN_EMAIL string

	RL_GLOBAL_WINDOW_MS    string
	RL_GLOBAL_MAX          string
	RL_AUTH_WINDOW_MS      string
	RL_AUTH_MAX            string
	RL_LOGIN_WINDOW_MS     string
	RL_LOGIN_MAX           string
	RL_PROTECTED_WINDOW_MS string
	RL_PROTECTED_MAX       string

	REDIS_PORT     string
	REDIS_HOST     string
	REDIS_PASSWORD string
	REDIS_USER     string
	REDIS_PLATFORM string
}

// NewENV loads process environment configuration, validates required keys, and installs
// the package singleton used by Current.
//
// A .env file is loaded via godotenv.Load when present; existing OS environment
// variables are not overridden by file values. Required keys are GO_ENV, APP_ENV,
// APP_PORT, API_ROUTE, APP_CHANNELS, AUTH_TYPE, JWT_SECRET, JWT_EXPIRES,
// JWT_COOKIE_EXPIRE, POSTGRES_URL, POSTGRES_TEST_URL, CORS_ORIGINS, and
// SUPERADMIN_EMAIL. Optional keys include APP_NAME, APP_VERSION, EMAIL_*,
// LOGS_PATH, Logtail tokens/URLs, RL_* rate-limit keys, and REDIS_* keys;
// empty optional values are stored as empty strings (EMAIL_ENABLED defaults
// to false when unset).
//
// APP_ENV drives boolean flags IsDev (development), IsStaging (staging), and
// IsProd (production). The constructed ENV struct is assigned to the
// package-level current variable before return.
//
// Side effects: may read .env from the working directory; mutates the package
// singleton current on success. Does not connect to Postgres/Redis or validate
// URI formats.
//
// Returns a pointer to the populated ENV and nil when all required variables
// are set. Returns nil and fmt.Errorf("Env Error: %s is not defined", key)
// for the first missing required variable encountered, or an error when
// APP_PORT is not a valid integer.
func NewENV() (*ENV, error) {
	// Load main environment file.
	// godotenv.Load() does not override existing environment variables.
	_ = godotenv.Load()

	required := []string{
		"GO_ENV",
		"APP_ENV",
		"APP_PORT",
		"API_ROUTE",
		"APP_CHANNELS",
		"AUTH_TYPE",
		"JWT_SECRET",
		"JWT_EXPIRES",
		"JWT_COOKIE_EXPIRE",
		"POSTGRES_URL",
		"POSTGRES_TEST_URL",
		"CORS_ORIGINS",
		"SUPERADMIN_EMAIL",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			return nil, fmt.Errorf("Env Error: %s is not defined", key)
		}
	}

	appPortRaw := os.Getenv("APP_PORT")
	appPort, err := strconv.Atoi(strings.TrimSpace(appPortRaw))
	if err != nil {
		return nil, fmt.Errorf("Env Error: APP_PORT must be a valid integer, got %q", appPortRaw)
	}

	emailEnabled := false
	if raw := strings.TrimSpace(os.Getenv("EMAIL_ENABLED")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("Env Error: EMAIL_ENABLED must be a valid boolean, got %q", raw)
		}
		emailEnabled = parsed
	}

	appEnv := os.Getenv("APP_ENV")

	env := &ENV{
		IsDev:       appEnv == "development",
		IsStaging:   appEnv == "staging",
		IsProd:      appEnv == "production",
		Environment: appEnv,

		APP_NAME:    os.Getenv("APP_NAME"),
		APP_VERSION: os.Getenv("APP_VERSION"),
		GO_ENV:      os.Getenv("GO_ENV"),
		APP_ENV:     appEnv,
		APP_PORT:    appPort,
		API_ROUTE:   os.Getenv("API_ROUTE"),

		APP_CHANNELS: os.Getenv("APP_CHANNELS"),
		AUTH_TYPE:    os.Getenv("AUTH_TYPE"),
		CORS_ORIGINS: splitCSV(os.Getenv("CORS_ORIGINS")),

		JWT_SECRET:        os.Getenv("JWT_SECRET"),
		JWT_EXPIRES:       os.Getenv("JWT_EXPIRES"),
		JWT_COOKIE_EXPIRE: os.Getenv("JWT_COOKIE_EXPIRE"),

		POSTGRES_URL:      os.Getenv("POSTGRES_URL"),
		POSTGRES_TEST_URL: os.Getenv("POSTGRES_TEST_URL"),

		EMAIL_ENABLED:  emailEnabled,
		RESEND_API_KEY: os.Getenv("RESEND_API_KEY"),
		EMAIL_FROM:     os.Getenv("EMAIL_FROM"),
		EMAIL_REPLY_TO: os.Getenv("EMAIL_REPLY_TO"),

		LOGS_PATH:             os.Getenv("LOGS_PATH"),
		LOGTAIL_MONITOR_TOKEN: os.Getenv("LOGTAIL_MONITOR_TOKEN"),
		LOGTAIL_AUDIT_TOKEN:   os.Getenv("LOGTAIL_AUDIT_TOKEN"),
		LOGTAIL_INGESTION_URL: os.Getenv("LOGTAIL_INGESTION_URL"),

		SUPERADMIN_EMAIL: os.Getenv("SUPERADMIN_EMAIL"),

		RL_GLOBAL_WINDOW_MS:    os.Getenv("RL_GLOBAL_WINDOW_MS"),
		RL_GLOBAL_MAX:          os.Getenv("RL_GLOBAL_MAX"),
		RL_AUTH_WINDOW_MS:      os.Getenv("RL_AUTH_WINDOW_MS"),
		RL_AUTH_MAX:            os.Getenv("RL_AUTH_MAX"),
		RL_LOGIN_WINDOW_MS:     os.Getenv("RL_LOGIN_WINDOW_MS"),
		RL_LOGIN_MAX:           os.Getenv("RL_LOGIN_MAX"),
		RL_PROTECTED_WINDOW_MS: os.Getenv("RL_PROTECTED_WINDOW_MS"),
		RL_PROTECTED_MAX:       os.Getenv("RL_PROTECTED_MAX"),

		REDIS_PORT:     os.Getenv("REDIS_PORT"),
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_USER:     os.Getenv("REDIS_USER"),
		REDIS_PLATFORM: os.Getenv("REDIS_PLATFORM"),
	}

	current = env

	return env, nil
}

// Current returns the environment singleton last successfully loaded by NewENV.
//
// If NewENV has not been called or the last call failed before assigning current,
// this function returns nil. Callers such as GetSignedJWTToken and Logger.env rely on
// a prior successful bootstrap during application startup.
//
// Side effects: none; read-only access to the package-level current pointer.
//
// Returns the cached *ENV or nil when no environment has been loaded yet.
func Current() *ENV {
	return current
}

// CacheEnv returns the active environment name used for cache key namespacing.
//
// It reads the package singleton loaded by NewENV (APP_ENV / Environment) and
// falls back to "development" when no environment has been loaded yet — e.g.
// in tests or before bootstrap — so cache callers never have to hardcode the
// env string. Use with ComputeKey:
//
//	cacheKey := utils.ComputeKey(utils.CacheEnv(), opts.CacheKey)
func CacheEnv() string {
	if current == nil {
		return "development"
	}
	if current.APP_ENV != "" {
		return current.APP_ENV
	}
	if current.Environment != "" {
		return current.Environment
	}
	return "development"
}

// CacheKey namespaces key per environment via ComputeKey and the active env.
//
// It is shorthand for ComputeKey(CacheEnv(), key), keeping middleware and
// services free of hardcoded env strings like "development".
func CacheKey(key string) string {
	return ComputeKey(CacheEnv(), key)
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}
