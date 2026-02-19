package envfile

type NoEncryptRule interface {
	ShouldSkip(key, value string) bool
}

type NoEncryptDetector struct {
	rules []NoEncryptRule
}

func NewNoEncryptDetector() *NoEncryptDetector {
	return &NoEncryptDetector{
		rules: defaultNoEncryptRules(),
	}
}

func (d *NoEncryptDetector) ShouldSkip(key, value string) bool {
	for _, rule := range d.rules {
		if rule.ShouldSkip(key, value) {
			return true
		}
	}
	return false
}
