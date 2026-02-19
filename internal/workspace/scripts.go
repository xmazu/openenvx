package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func SuggestDevRunCommand(root string) string {
	path := filepath.Join(root, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil || pkg.Scripts == nil {
		return ""
	}
	if cmd, ok := pkg.Scripts["dev"]; ok && cmd != "" {
		return "npm run dev"
	}
	if cmd, ok := pkg.Scripts["start"]; ok && cmd != "" {
		return "npm run start"
	}
	return ""
}
