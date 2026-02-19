package envfile

import (
	"slices"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func defaultNoEncryptRules() []NoEncryptRule {
	return []NoEncryptRule{
		&NoEncryptURLRule{},
		&NoEncryptHostnameRule{},
		&NoEncryptBooleanRule{},
		&NoEncryptNumberRule{},
		&NoEncryptNodeEnvRule{},
	}
}

type NoEncryptURLRule struct{}

func (r *NoEncryptURLRule) ShouldSkip(key, value string) bool {
	if !r.looksLikeURL(value) {
		return false
	}
	u, err := url.Parse(value)
	if err != nil {
		return false
	}
	if u.User != nil {
		username := u.User.Username()
		_, hasPassword := u.User.Password()
		if username != "" || hasPassword {
			return false
		}
	}
	return true
}

func (r *NoEncryptURLRule) looksLikeURL(value string) bool {
	return strings.Contains(value, "://")
}

type NoEncryptHostnameRule struct{}

var localhostPattern = regexp.MustCompile(`(?i)^(localhost|127\.0\.0\.1|::1)$`)

func (r *NoEncryptHostnameRule) ShouldSkip(key, value string) bool {
	return localhostPattern.MatchString(value)
}

type NoEncryptBooleanRule struct{}

func (r *NoEncryptBooleanRule) ShouldSkip(key, value string) bool {
	lower := strings.ToLower(value)
	return lower == "true" || lower == "false" || lower == "yes" || lower == "no"
}

type NoEncryptNumberRule struct{}

func (r *NoEncryptNumberRule) ShouldSkip(key, value string) bool {
	if _, err := strconv.Atoi(value); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return true
	}
	return false
}

type NoEncryptNodeEnvRule struct{}

var nodeEnvAllowed = map[string][]string{
	"NODE_ENV":   {"development", "production", "test"},
	"LOG_LEVEL":  {"debug", "info", "warn", "error", "verbose"},
	"HOST":       {"localhost", "127.0.0.1", "::1", "0.0.0.0"},
}

func (r *NoEncryptNodeEnvRule) ShouldSkip(key, value string) bool {
	allowed, ok := nodeEnvAllowed[strings.ToUpper(key)]
	if !ok {
		if strings.ToUpper(key) == "PORT" {
			_, err := strconv.Atoi(value)
			return err == nil
		}
		return false
	}
	v := strings.ToLower(strings.TrimSpace(value))
	return slices.Contains(allowed, v)
}
