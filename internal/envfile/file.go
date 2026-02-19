package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/xmazu/openenvx/internal/crypto"
)

type File struct {
	path     string
	lines    []*Line
	keyIndex map[string]int
}

func New(path string) *File {
	return &File{
		path:     path,
		lines:    []*Line{},
		keyIndex: make(map[string]int),
	}
}

func Load(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(path), nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	f := &File{
		path:     path,
		lines:    []*Line{},
		keyIndex: make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	const maxCapacity = 512 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		f.parseLine(line, lineNum)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return f, nil
}

func (f *File) parseLine(line string, num int) {
	if strings.TrimSpace(line) == "" {
		f.lines = append(f.lines, &Line{Type: LineTypeEmpty, Num: num, Raw: line})
		return
	}

	trimmed := strings.TrimSpace(line)
	if after, ok :=strings.CutPrefix(trimmed, "#"); ok  {
		afterComment := strings.TrimSpace(after)
		if afterComment != "" && strings.Contains(afterComment, "=") {
			parts := strings.SplitN(afterComment, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key != "" {
					value = unquote(value)
					f.lines = append(f.lines, &Line{
						Type:  LineTypeCommentedAssignment,
						Num:   num,
						Raw:   line,
						Key:   key,
						Value: value,
					})
					return
				}
			}
		}

		f.lines = append(f.lines, &Line{Type: LineTypeComment, Num: num, Raw: line})
		return
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		f.lines = append(f.lines, &Line{Type: LineTypeInvalid, Num: num, Raw: line})
		return
	}

	key := strings.TrimSpace(parts[0])
	valuePart := parts[1]

	value, inlineComment := parseInlineComment(valuePart)
	value = unquote(value)

	idx := len(f.lines)
	f.lines = append(f.lines, &Line{
		Type:          LineTypeVariable,
		Num:           num,
		Raw:           line,
		Key:           key,
		Value:         value,
		InlineComment: inlineComment,
	})
	f.keyIndex[key] = idx
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func parseInlineComment(valuePart string) (value, inlineComment string) {
	commentIdx := findInlineCommentStart(valuePart)
	if commentIdx == -1 {
		return strings.TrimSpace(valuePart), ""
	}
	value = strings.TrimSpace(valuePart[:commentIdx])
	inlineComment = valuePart[commentIdx:]
	return value, inlineComment
}

func findInlineCommentStart(s string) int {
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !inQuote && (c == '"' || c == '\'') {
			inQuote = true
			quoteChar = c
			continue
		}
		if inQuote && c == quoteChar {
			inQuote = false
			quoteChar = 0
			continue
		}
		if !inQuote && c == '#' {
			start := i
			for start > 0 && (s[start-1] == ' ' || s[start-1] == '\t') {
				start--
			}
			return start
		}
	}
	return -1
}

func (f *File) Save() error {
	file, err := os.Create(f.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, line := range f.lines {
		switch line.Type {
		case LineTypeVariable:
			if line.InlineComment != "" {
				fmt.Fprintf(writer, "%s=%s%s\n", line.Key, line.Value, line.InlineComment)
			} else {
				fmt.Fprintf(writer, "%s=%s\n", line.Key, line.Value)
			}
		default:
			fmt.Fprintln(writer, line.Raw)
		}
	}

	return writer.Flush()
}

func (f *File) Get(key string) (string, bool) {
	idx, ok := f.keyIndex[key]
	if !ok {
		return "", false
	}
	return f.lines[idx].Value, true
}

func (f *File) Set(key, value string) {
	if idx, exists := f.keyIndex[key]; exists {
		f.lines[idx].Value = value
		return
	}

	idx := len(f.lines)
	f.lines = append(f.lines, &Line{
		Type:  LineTypeVariable,
		Num:   idx + 1,
		Key:   key,
		Value: value,
	})
	f.keyIndex[key] = idx
}

func (f *File) Keys() []string {
	keys := make([]string, 0, len(f.keyIndex))
	for k := range f.keyIndex {
		keys = append(keys, k)
	}
	return keys
}

func (f *File) Delete(key string) bool {
	idx, ok := f.keyIndex[key]
	if !ok {
		return false
	}
	f.lines = append(f.lines[:idx], f.lines[idx+1:]...)
	delete(f.keyIndex, key)
	for k, v := range f.keyIndex {
		if v > idx {
			f.keyIndex[k] = v - 1
		}
	}
	return true
}

func (f *File) DecryptAll(env *crypto.Envelope) (map[string]string, error) {
	decrypted := make(map[string]string)

	for _, line := range f.lines {
		if line.Type != LineTypeVariable {
			continue
		}

		ev, err := crypto.ParseEncryptedValue(line.Value)
		if err != nil {
			decrypted[line.Key] = line.Value
			continue
		}

		plaintext, err := env.Decrypt(ev, line.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %w", line.Key, err)
		}

		decrypted[line.Key] = string(plaintext)
	}

	return decrypted, nil
}
