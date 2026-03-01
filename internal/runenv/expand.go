package runenv

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const commandSubstPrefix = "$("

// varRef matches ${VAR} where VAR is a single identifier (letters, digits, underscore).
var varRef = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// runShellCommand runs command via sh -c and returns trimmed stdout.
// Uses the current process environment (PATH, HOME, etc.).
func runShellCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = nil
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("command %q: %w (stderr: %s)", command, err, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("command %q: %w", command, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// findNextCommandSubst finds the first $(...) in s and returns the text before it,
// the command inside the parens, and the text after. Handles nested parentheses.
// If no $(...) is found, before is s and found is false.
func findNextCommandSubst(s string) (before, command, after string, found bool, err error) {
	start := strings.Index(s, commandSubstPrefix)
	if start < 0 {
		return s, "", "", false, nil
	}
	before = s[:start]
	// Find the matching ')' by counting nested parens
	open := start + len(commandSubstPrefix)
	pos := open
	depth := 1
	for pos < len(s) && depth > 0 {
		switch s[pos] {
		case '(':
			depth++
		case ')':
			depth--
		}
		pos++
	}
	if depth != 0 {
		return "", "", "", true, fmt.Errorf("unclosed $(...) in value")
	}
	command = strings.TrimSpace(s[open : pos-1])
	after = s[pos:]
	return before, command, after, true, nil
}

// expandCommandsInValue replaces every $(...) in s with the output of running that command.
func expandCommandsInValue(s string) (string, error) {
	var result strings.Builder
	for {
		before, command, after, found, err := findNextCommandSubst(s)
		if err != nil {
			return "", err
		}
		result.WriteString(before)
		if !found {
			break
		}
		if command == "" {
			return "", fmt.Errorf("empty $(...) command")
		}
		out, err := runShellCommand(command)
		if err != nil {
			return "", err
		}
		result.WriteString(out)
		s = after
	}
	return result.String(), nil
}

// expandCommandSubstitutionInMap runs $(...) in every value and returns a new map.
func expandCommandSubstitutionInMap(env map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(env))
	for k, v := range env {
		expanded, err := expandCommandsInValue(v)
		if err != nil {
			return nil, err
		}
		out[k] = expanded
	}
	return out, nil
}

// expandVarRefsInString replaces all ${VAR} in s by looking up VAR via getVar.
// getVar is typically a resolver that may recurse for other keys.
func expandVarRefsInString(s string, env map[string]string, getVar func(string) (string, error)) (string, error) {
	result := s
	for {
		loc := varRef.FindStringSubmatchIndex(result)
		if loc == nil {
			return result, nil
		}
		refKey := result[loc[2]:loc[3]]
		if _, ok := env[refKey]; !ok {
			return "", fmt.Errorf("undefined variable: %s", refKey)
		}
		refVal, err := getVar(refKey)
		if err != nil {
			return "", err
		}
		result = result[:loc[0]] + refVal + result[loc[1]:]
	}
}

// expandVariableReferencesInMap resolves all ${VAR} references; detects circular and undefined refs.
func expandVariableReferencesInMap(env map[string]string) (map[string]string, error) {
	resolved := make(map[string]string, len(env))
	onStack := make(map[string]bool)

	var resolve func(key string) (string, error)
	resolve = func(key string) (string, error) {
		if onStack[key] {
			return "", fmt.Errorf("circular reference involving %q", key)
		}
		if v, ok := resolved[key]; ok {
			return v, nil
		}
		raw, ok := env[key]
		if !ok {
			return "", fmt.Errorf("undefined variable: %s", key)
		}
		onStack[key] = true
		defer func() { delete(onStack, key) }()

		expanded, err := expandVarRefsInString(raw, env, resolve)
		if err != nil {
			return "", err
		}
		resolved[key] = expanded
		return expanded, nil
	}

	for key := range env {
		if _, err := resolve(key); err != nil {
			return nil, err
		}
	}
	return resolved, nil
}

// ExpandMap expands env in two phases:
// 1) Command substitution: every $(command) is run and replaced by its stdout.
// 2) Variable expansion: every ${VAR} is replaced by the value of VAR in the same map.
// Returns an error on undefined variables, circular references, or command failures.
func ExpandMap(env map[string]string) (map[string]string, error) {
	if len(env) == 0 {
		return map[string]string{}, nil
	}
	env, err := expandCommandSubstitutionInMap(env)
	if err != nil {
		return nil, err
	}
	return expandVariableReferencesInMap(env)
}
