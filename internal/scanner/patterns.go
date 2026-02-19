package scanner

import (
	"regexp"
)

type Pattern struct {
	Name     string
	Regex    *regexp.Regexp
	Severity string
}

var Patterns = []Pattern{
	{
		Name:     "AWS Access Key ID",
		Regex:    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		Severity: "high",
	},
	{
		Name:     "AWS Secret Access Key",
		Regex:    regexp.MustCompile(`(?i)aws(.{0,20})?['\"][0-9a-zA-Z/+=]{40}['\"]`),
		Severity: "high",
	},
	{
		Name:     "GitHub Personal Access Token",
		Regex:    regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`),
		Severity: "high",
	},
	{
		Name:     "GitHub OAuth Token",
		Regex:    regexp.MustCompile(`gho_[0-9a-zA-Z]{36}`),
		Severity: "high",
	},
	{
		Name:     "GitHub App Token",
		Regex:    regexp.MustCompile(`(ghu|ghs)_[0-9a-zA-Z]{36}`),
		Severity: "high",
	},
	{
		Name:     "GitHub Fine-Grained PAT",
		Regex:    regexp.MustCompile(`github_pat_[0-9a-zA-Z_]{22,}`),
		Severity: "high",
	},
	{
		Name:     "GitLab Personal Access Token",
		Regex:    regexp.MustCompile(`glpat-[0-9a-zA-Z\-]{20}`),
		Severity: "high",
	},
	{
		Name:     "GitLab Deploy Token",
		Regex:    regexp.MustCompile(`gldt-[0-9a-zA-Z\-]{20}`),
		Severity: "high",
	},
	{
		Name:     "GitLab Feed Token",
		Regex:    regexp.MustCompile(`glft-[0-9a-zA-Z\-]{20}`),
		Severity: "high",
	},
	{
		Name:     "Slack Token",
		Regex:    regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}(-[0-9a-zA-Z]{24})?`),
		Severity: "high",
	},
	{
		Name:     "Slack Webhook URL",
		Regex:    regexp.MustCompile(`https://hooks\.slack\.com/services/T[0-9A-Z]{8,12}/B[0-9A-Z]{8,12}/[0-9a-zA-Z]{24}`),
		Severity: "high",
	},
	{
		Name:     "Private Key",
		Regex:    regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
		Severity: "high",
	},
	{
		Name:     "Google Maps API Key",
		Regex:    regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),
		Severity: "high",
	},
	{
		Name:     "Google Cloud API Key",
		Regex:    regexp.MustCompile(`(?i)(google|gcp|gc)[\s._-]?(api[\s._-]?key|key)[\s]*[=:][\s]*['"]?[0-9a-zA-Z_-]{39}['"]?`),
		Severity: "high",
	},
	{
		Name:     "Mapbox Token",
		Regex:    regexp.MustCompile(`pk\.[a-zA-Z0-9]{60,}`),
		Severity: "high",
	},
	{
		Name:     "Stripe Secret Key",
		Regex:    regexp.MustCompile(`sk_(live|test)_[0-9a-zA-Z]{24,}`),
		Severity: "high",
	},
	{
		Name:     "Stripe Restricted Key",
		Regex:    regexp.MustCompile(`rk_(live|test)_[0-9a-zA-Z]{24,}`),
		Severity: "high",
	},
	{
		Name:     "Square Access Token",
		Regex:    regexp.MustCompile(`sq0atp-[0-9A-Za-z\-_]{22}`),
		Severity: "high",
	},
	{
		Name:     "Square OAuth Secret",
		Regex:    regexp.MustCompile(`sq0csp-[0-9A-Za-z\-_]{43}`),
		Severity: "high",
	},
	{
		Name:     "Twilio API Key",
		Regex:    regexp.MustCompile(`SK[0-9a-f]{32}`),
		Severity: "high",
	},
	{
		Name:     "Twilio Auth Token",
		Regex:    regexp.MustCompile(`(?i)twilio.{0,20}auth[_-]?token[\s]*[=:][\s]*['"]?[0-9a-f]{32}['"]?`),
		Severity: "high",
	},
	{
		Name:     "Twilio Account SID",
		Regex:    regexp.MustCompile(`AC[0-9a-f]{32}`),
		Severity: "medium",
	},
	{
		Name:     "OpenAI API Key",
		Regex:    regexp.MustCompile(`sk-[a-zA-Z0-9_-]{20,}`),
		Severity: "high",
	},
	{
		Name:     "Anthropic API Key",
		Regex:    regexp.MustCompile(`sk-ant-[a-zA-Z0-9\-_]{80,}`),
		Severity: "high",
	},
	{
		Name:     "Azure Access Key",
		Regex:    regexp.MustCompile(`(?i)(azure|storage)[\s._-]?(key|access[\s._-]?key)[\s]*[=:][\s]*['"]?[0-9a-zA-Z/+]{88}['"]?`),
		Severity: "high",
	},
	{
		Name:     "Azure SAS Token",
		Regex:    regexp.MustCompile(`\?.*(sig|se|st|sv)=`),
		Severity: "medium",
	},
	{
		Name:     "Cloudflare API Token",
		Regex:    regexp.MustCompile(`(?i)cloudflare.{0,20}['"]?[0-9a-zA-Z_-]{40}['"]?`),
		Severity: "high",
	},
	{
		Name:     "DigitalOcean Token",
		Regex:    regexp.MustCompile(`(?i)(digitalocean|do)[\s._-]?(api[\s._-]?key|token|pat)[\s]*[=:][\s]*['"]?[a-f0-9]{64}['"]?`),
		Severity: "high",
	},
	{
		Name:     "Heroku API Key",
		Regex:    regexp.MustCompile(`(?i)heroku.{0,20}['"]?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}['"]?`),
		Severity: "high",
	},
	{
		Name:     "NPM Access Token",
		Regex:    regexp.MustCompile(`//registry\.npmjs\.org/:_authToken=[0-9a-zA-Z-]{36}`),
		Severity: "high",
	},
	{
		Name:     "PyPI API Token",
		Regex:    regexp.MustCompile(`pypi-AgEIcHlwaS5vcmc[\w-]{50,}`),
		Severity: "high",
	},
	{
		Name:     "Docker Hub PAT",
		Regex:    regexp.MustCompile(`dckr_pat_[0-9a-zA-Z_-]{27}`),
		Severity: "high",
	},
	{
		Name:     "Discord Bot Token",
		Regex:    regexp.MustCompile(`[MN][A-Za-z\d]{23}\.[\w-]{6}\.[\w-]{27}`),
		Severity: "high",
	},
	{
		Name:     "Telegram Bot Token",
		Regex:    regexp.MustCompile(`[0-9]{5,16}:[a-zA-Z0-9_-]{35}`),
		Severity: "high",
	},
	{
		Name:     "SendGrid API Key",
		Regex:    regexp.MustCompile(`SG\.[a-zA-Z0-9=_\-\.]{66}`),
		Severity: "high",
	},
	{
		Name:     "Mailchimp API Key",
		Regex:    regexp.MustCompile(`[a-f0-9]{32}-us[0-9]{1,2}`),
		Severity: "high",
	},
	{
		Name:     "Firebase API Key",
		Regex:    regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),
		Severity: "medium",
	},
	{
		Name:     "Firebase Cloud Messaging Server Key",
		Regex:    regexp.MustCompile(`AAAA[a-zA-Z0-9_-]{7}:[a-zA-Z0-9_-]{140}`),
		Severity: "high",
	},
	{
		Name:     "Shopify Access Token",
		Regex:    regexp.MustCompile(`shpat_[a-fA-F0-9]{32}`),
		Severity: "high",
	},
	{
		Name:     "Shopify Custom App Access Token",
		Regex:    regexp.MustCompile(`shpca_[a-fA-F0-9]{32}`),
		Severity: "high",
	},
	{
		Name:     "Linear API Token",
		Regex:    regexp.MustCompile(`lin_api_[a-zA-Z0-9]{40}`),
		Severity: "high",
	},
	{
		Name:     "Notion Integration Token",
		Regex:    regexp.MustCompile(`secret_[a-zA-Z0-9]{43}`),
		Severity: "high",
	},
	{
		Name:     "Asana Personal Access Token",
		Regex:    regexp.MustCompile(`[0-9]/[0-9]{16}:[a-zA-Z0-9]{32}`),
		Severity: "high",
	},
	{
		Name:     "Reddit Refresh Token",
		Regex:    regexp.MustCompile(`[0-9a-zA-Z]{12}-[0-9a-zA-Z]{24}`),
		Severity: "high",
	},
	{
		Name:     "JWT Token",
		Regex:    regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
		Severity: "medium",
	},
}

type Match struct {
	Pattern Pattern
	File    string
	Line    int
	Column  int
	Match   string
	Context string
}
