package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSuggestDevRunCommand(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantCmd  string
		writeFile bool
	}{
		{"no package.json", "", "", false},
		{"empty scripts", `{"scripts":{}}`, "", true},
		{"dev script", `{"scripts":{"dev":"next dev"}}`, "npm run dev", true},
		{"start script", `{"scripts":{"start":"node server.js"}}`, "npm run start", true},
		{"dev preferred over start", `{"scripts":{"start":"node .","dev":"next dev"}}`, "npm run dev", true},
		{"invalid json", `{scripts}`, "", true},
		{"no scripts key", `{}`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if tt.writeFile {
				path := filepath.Join(root, "package.json")
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := SuggestDevRunCommand(root)
			if got != tt.wantCmd {
				t.Errorf("SuggestDevRunCommand() = %q, want %q", got, tt.wantCmd)
			}
		})
	}
}
