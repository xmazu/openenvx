package scanner

import (
	"regexp"
	"testing"
)

func TestPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		matches  []string
		noMatch  []string
		severity string
	}{
		{
			name:     "AWS Access Key ID",
			pattern:  "AWS Access Key ID",
			matches:  []string{"AKIAIOSFODNN7EXAMPLE", "AKIA1234567890ABCDEF", "AKIA1234567890ABCXYZ"},
			noMatch:  []string{"AKIA123", "akialowercase1234567890ABCDEF", "AKI1234567890ABCDEF"},
			severity: "high",
		},
		{
			name:     "GitHub Personal Access Token",
			pattern:  "GitHub Personal Access Token",
			matches:  []string{"ghp_1234567890abcdefghijklmnopqrstuvwxyz12", "ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			noMatch:  []string{"ghp_12345", "github_token_1234567890abcdefghijklmnopqrstuvwxyz12"},
			severity: "high",
		},
		{
			name:     "GitHub OAuth Token",
			pattern:  "GitHub OAuth Token",
			matches:  []string{"gho_1234567890abcdefghijklmnopqrstuvwxyz12", "gho_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			noMatch:  []string{"gho_12345", "oauth_token_1234567890abcdefghijklmnopqrstuvwxyz12"},
			severity: "high",
		},
		{
			name:     "GitLab Personal Access Token",
			pattern:  "GitLab Personal Access Token",
			matches:  []string{"glpat-xxxxxxxxxxxxxxxxxxxx", "glpat-aBcDeFgHiJkLmNoPqRsT"},
			noMatch:  []string{"glpat-short", "gitlab_token"},
			severity: "high",
		},
		{
			name:    "Slack Token",
			pattern: "Slack Token",
			matches: []string{
				"xoxb-1234567890123-1234567890123-AbCdEfGhIjKlMnOpQrStUvWx",
				"xoxp-1234567890123-1234567890123",
				"xoxa-1234567890123-1234567890123-123456789012345678901234",
			},
			noMatch:  []string{"slack_token", "xoxb-123", "xoxb-invalid"},
			severity: "high",
		},
		{
			name:    "Private Key",
			pattern: "Private Key",
			matches: []string{
				"-----BEGIN RSA PRIVATE KEY-----",
				"-----BEGIN PRIVATE KEY-----",
				"-----BEGIN EC PRIVATE KEY-----",
				"-----BEGIN OPENSSH PRIVATE KEY-----",
			},
			noMatch:  []string{"BEGIN PRIVATE KEY", "-----BEGIN PUBLIC KEY-----"},
			severity: "high",
		},
		{
			name:     "Google Maps API Key",
			pattern:  "Google Maps API Key",
			matches:  []string{"AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", "AIzaSyA1234567890_-abcdefghijklmnopqrstuvwxyz"},
			noMatch:  []string{"AIzaShort", "AIza123"},
			severity: "high",
		},
		{
			name:     "Mapbox Token",
			pattern:  "Mapbox Token",
			matches:  []string{"pk.eyJ1IjoiZXhhbXBsZSIsImEiOiJjanJhZ3BhYnkwdjRkM3hwZ2RqZmZ3YWQifQ.example"},
			noMatch:  []string{"pk.short", "mapbox_token"},
			severity: "high",
		},
		{
			name:     "Stripe Secret Key",
			pattern:  "Stripe Secret Key",
			matches:  []string{"sk_live_4eC39HqLyjWDarjtT1zdp7dc", "sk_test_4eC39HqLyjWDarjtT1zdp7dc"},
			noMatch:  []string{"pk_live_123", "sk_short"},
			severity: "high",
		},
		{
			name:     "Twilio API Key",
			pattern:  "Twilio API Key",
			matches:  []string{"SKa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"},
			noMatch:  []string{"SK-short", "twilio_key"},
			severity: "high",
		},
		{
			name:     "OpenAI API Key",
			pattern:  "OpenAI API Key",
			matches:  []string{"sk-proj-abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJ"},
			noMatch:  []string{"sk-short", "openai_key"},
			severity: "high",
		},
		{
			name:     "NPM Access Token",
			pattern:  "NPM Access Token",
			matches:  []string{"//registry.npmjs.org/:_authToken=abcd1234-ef56-7890-abcd-ef1234567890"},
			noMatch:  []string{"npm_token", "//npmjs.org/token"},
			severity: "high",
		},
		{
			name:     "PyPI API Token",
			pattern:  "PyPI API Token",
			matches:  []string{"pypi-AgEIcHlwaS5vcmcaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			noMatch:  []string{"pypi_token", "pypi-"},
			severity: "high",
		},
		{
			name:     "Discord Bot Token",
			pattern:  "Discord Bot Token",
			matches:  []string{"MTIzNDU2Nzg5MDEyMzQ1Njc4.GaBcDe.FgHiJkLmNoPqRsTuVwXyZaBcDeFgHiJkLmNoPqR"},
			noMatch:  []string{"discord_token", "MTIz.GaBc.FgHi"},
			severity: "high",
		},
		{
			name:     "Telegram Bot Token",
			pattern:  "Telegram Bot Token",
			matches:  []string{"123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsawa"},
			noMatch:  []string{"telegram_token", "1234567:AAHdq"},
			severity: "high",
		},
		{
			name:     "SendGrid API Key",
			pattern:  "SendGrid API Key",
			matches:  []string{"SG.abcdefghijklmnopqrstuvwxABCDEFGHIJKLMNOPqrstuvwxYZ0123456789.abcde"},
			noMatch:  []string{"sendgrid_key", "SG.short"},
			severity: "high",
		},
		{
			name:     "JWT Token",
			pattern:  "JWT Token",
			matches:  []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
			noMatch:  []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "jwt_token"},
			severity: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pattern *Pattern
			for _, p := range Patterns {
				if p.Name == tt.pattern {
					pattern = &p
					break
				}
			}

			if pattern == nil {
				t.Fatalf("pattern %q not found", tt.pattern)
			}

			if pattern.Severity != tt.severity {
				t.Errorf("severity = %q, want %q", pattern.Severity, tt.severity)
			}

			for _, match := range tt.matches {
				if !pattern.Regex.MatchString(match) {
					t.Errorf("pattern %q should match %q", tt.pattern, match)
				}
			}

			for _, noMatch := range tt.noMatch {
				if pattern.Regex.MatchString(noMatch) {
					t.Errorf("pattern %q should NOT match %q", tt.pattern, noMatch)
				}
			}
		})
	}
}

func TestPatternsNotEmpty(t *testing.T) {
	if len(Patterns) == 0 {
		t.Error("Patterns slice is empty")
	}

	for i, p := range Patterns {
		if p.Name == "" {
			t.Errorf("pattern %d has no name", i)
		}
		if p.Regex == nil {
			t.Errorf("pattern %q has nil regex", p.Name)
		}
		if p.Severity == "" {
			t.Errorf("pattern %q has no severity", p.Name)
		}
	}
}

func TestMatchStruct(t *testing.T) {
	pattern := Pattern{
		Name:     "Test Pattern",
		Regex:    regexp.MustCompile(`test`),
		Severity: "high",
	}

	match := Match{
		Pattern: pattern,
		File:    "test.go",
		Line:    42,
		Column:  10,
		Match:   "test",
		Context: "this is a test line",
	}

	if match.Pattern.Name != "Test Pattern" {
		t.Error("Match.Pattern.Name not accessible")
	}

	if match.File != "test.go" {
		t.Error("Match.File not accessible")
	}

	if match.Line != 42 {
		t.Error("Match.Line not accessible")
	}

	if match.Column != 10 {
		t.Error("Match.Column not accessible")
	}

	if match.Match != "test" {
		t.Error("Match.Match not accessible")
	}

	if match.Context != "this is a test line" {
		t.Error("Match.Context not accessible")
	}
}

func TestPatternRegexCompilation(t *testing.T) {
	for _, p := range Patterns {
		if p.Regex == nil {
			t.Errorf("pattern %q has nil regex", p.Name)
			continue
		}

		_ = p.Regex.String()
	}
}
