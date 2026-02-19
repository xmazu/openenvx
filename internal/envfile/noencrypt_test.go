package envfile

import (
	"testing"
)

func TestNoEncryptDetector_URLRule(t *testing.T) {
	d := NewNoEncryptDetector()
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
	}{
		{"URL without credentials", "DATABASE_URL", "postgres://localhost:5432/mydb", true},
		{"URL with username only", "DATABASE_URL", "postgres://user@localhost:5432/mydb", false},
		{"URL with username and password", "DATABASE_URL", "postgres://user:pass@localhost:5432/mydb", false},
		{"Redis URL without credentials", "REDIS_URL", "redis://localhost:6379", true},
		{"Redis URL with password", "REDIS_URL", "redis://:password@localhost:6379", false},
		{"MySQL URL with credentials", "DATABASE_URL", "mysql://admin:secret@db.example.com:3306/mydb", false},
		{"HTTP URL without credentials", "API_URL", "http://api.example.com", true},
		{"HTTP URL with credentials", "API_URL", "http://user:pass@api.example.com", false},
		{"HTTPS URL without credentials", "API_URL", "https://api.example.com/v1", true},
		{"HTTPS URL with credentials", "API_URL", "https://user:pass@api.example.com/v1", false},
		{"MongoDB URL without credentials", "MONGO_URL", "mongodb://localhost:27017/mydb", true},
		{"MongoDB URL with credentials", "MONGO_URL", "mongodb://user:pass@localhost:27017/mydb", false},
		{"Not a URL", "SOME_VALUE", "not-a-url-value", false},
		{"Empty value", "EMPTY", "", false},
		{"URL with empty user info", "DATABASE_URL", "postgres://@localhost:5432/mydb", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.ShouldSkip(tt.key, tt.value); got != tt.expected {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}

func TestNoEncryptDetector_HostnameRule(t *testing.T) {
	d := NewNoEncryptDetector()
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
	}{
		{"localhost", "HOST", "localhost", true},
		{"LOCALHOST uppercase", "HOST", "LOCALHOST", true},
		{"Localhost mixed case", "HOST", "Localhost", true},
		{"127.0.0.1", "HOST", "127.0.0.1", true},
		{"::1", "HOST", "::1", true},
		{"Simple hostname (not skipped)", "HOST", "webserver", false},
		{"Hostname with hyphen (not skipped)", "HOST", "my-web-server", false},
		{"IP address", "HOST", "192.168.1.1", false},
		{"Domain name", "HOST", "example.com", false},
		{"Hostname with port", "HOST", "localhost:3000", false},
		{"Empty value", "HOST", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.ShouldSkip(tt.key, tt.value); got != tt.expected {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}

func TestNoEncryptDetector_BooleanRule(t *testing.T) {
	d := NewNoEncryptDetector()
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
	}{
		{"true", "DEBUG", "true", true},
		{"false", "DEBUG", "false", true},
		{"TRUE", "DEBUG", "TRUE", true},
		{"FALSE", "DEBUG", "FALSE", true},
		{"True", "DEBUG", "True", true},
		{"False", "DEBUG", "False", true},
		{"yes", "ENABLED", "yes", true},
		{"no", "ENABLED", "no", true},
		{"YES", "ENABLED", "YES", true},
		{"NO", "ENABLED", "NO", true},
		{"not boolean", "VALUE", "something", false},
		{"empty", "VALUE", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.ShouldSkip(tt.key, tt.value); got != tt.expected {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}

func TestNoEncryptDetector_NumberRule(t *testing.T) {
	d := NewNoEncryptDetector()
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
	}{
		{"Integer", "PORT", "3000", true},
		{"Zero", "PORT", "0", true},
		{"Negative integer", "OFFSET", "-100", true},
		{"Float", "RATE", "0.5", true},
		{"Negative float", "RATE", "-0.5", true},
		{"Large number", "BIG", "1234567890", true},
		{"Not a number", "VALUE", "abc", false},
		{"Number with unit", "SIZE", "100MB", false},
		{"Empty", "EMPTY", "", false},
		{"Hex string", "HEX", "0x1234", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.ShouldSkip(tt.key, tt.value); got != tt.expected {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}

func TestNoEncryptDetector_NodeEnvRule(t *testing.T) {
	d := NewNoEncryptDetector()
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
	}{
		{"NODE_ENV development", "NODE_ENV", "development", true},
		{"NODE_ENV production", "NODE_ENV", "production", true},
		{"NODE_ENV test", "NODE_ENV", "test", true},
		{"NODE_ENV case insensitive", "NODE_ENV", "PRODUCTION", true},
		{"NODE_ENV unknown (encrypt)", "NODE_ENV", "staging", false},
		{"PORT numeric", "PORT", "3000", true},
		{"PORT zero", "PORT", "0", true},
		{"PORT non-numeric (encrypt)", "PORT", "three-thousand", false},
		{"HOST localhost", "HOST", "localhost", true},
		{"HOST 0.0.0.0", "HOST", "0.0.0.0", true},
		{"HOST remote (encrypt)", "HOST", "api.example.com", false},
		{"LOG_LEVEL info", "LOG_LEVEL", "info", true},
		{"LOG_LEVEL debug", "LOG_LEVEL", "debug", true},
		{"LOG_LEVEL unknown (encrypt)", "LOG_LEVEL", "trace", false},
		{"Other key (encrypt)", "API_KEY", "dev", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.ShouldSkip(tt.key, tt.value); got != tt.expected {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}

func TestNoEncryptDetector_CombinedRules(t *testing.T) {
	d := NewNoEncryptDetector()
	tests := []struct {
		name     string
		key      string
		value    string
		expected bool
		reason   string
	}{
		{"Secret value", "API_SECRET", "super-secret-key-12345", false, "no rule matches"},
		{"Port number", "PORT", "3000", true, "number rule"},
		{"Debug boolean", "DEBUG", "true", true, "boolean rule"},
		{"Database URL with credentials", "DATABASE_URL", "postgres://admin:secret@localhost/db", false, "URL with creds"},
		{"Database URL without credentials", "DATABASE_URL", "postgres://localhost:5432/db", true, "URL without creds"},
		{"Host value", "HOST", "localhost", true, "hostname rule"},
		{"Connection string with password", "CONN", "Server=localhost;Password=secret", false, "no rule matches"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := d.ShouldSkip(tt.key, tt.value); got != tt.expected {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v (reason: %s)", tt.key, tt.value, got, tt.expected, tt.reason)
			}
		})
	}
}
