package workspace

import (
	"testing"
)

func TestIsEnvFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"exact .env", ".env", true},
		{".env.local", ".env.local", true},
		{".env.production", ".env.production", true},
		{".env.example", ".env.example", false}, // ignored (template only)
		{"not .env", "env", false},
		{"random file", "config.yaml", false},
		{"too short", ".env", true},
		{"just prefix", ".env.", false},
		{"single char suffix", ".env.a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEnvFilename(tt.filename)
			if got != tt.want {
				t.Errorf("IsEnvFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestBuildEnvTree(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		check func(t *testing.T, root *EnvTreeNode)
	}{
		{
			name:  "single file",
			paths: []string{".env"},
			check: func(t *testing.T, root *EnvTreeNode) {
				if root.Name != "." {
					t.Errorf("root.Name = %q, want '.'", root.Name)
				}
				if len(root.Children) != 1 {
					t.Fatalf("len(root.Children) = %d, want 1", len(root.Children))
				}
				if root.Children[0].Name != ".env" {
					t.Errorf("child.Name = %q, want '.env'", root.Children[0].Name)
				}
				if root.Children[0].File != ".env" {
					t.Errorf("child.File = %q, want '.env'", root.Children[0].File)
				}
			},
		},
		{
			name:  "nested files",
			paths: []string{"packages/app/.env", "packages/api/.env"},
			check: func(t *testing.T, root *EnvTreeNode) {
				if len(root.Children) != 1 {
					t.Fatalf("len(root.Children) = %d, want 1", len(root.Children))
				}
				if root.Children[0].Name != "packages" {
					t.Errorf("child.Name = %q, want 'packages'", root.Children[0].Name)
				}
				if len(root.Children[0].Children) != 2 {
					t.Fatalf("len(packages.Children) = %d, want 2", len(root.Children[0].Children))
				}
			},
		},
		{
			name:  "files and directories mixed",
			paths: []string{".env", "packages/app/.env"},
			check: func(t *testing.T, root *EnvTreeNode) {
				if len(root.Children) != 2 {
					t.Fatalf("len(root.Children) = %d, want 2", len(root.Children))
				}
				// Files should come before directories
				if root.Children[0].Name != ".env" {
					t.Errorf("first child.Name = %q, want '.env'", root.Children[0].Name)
				}
				if root.Children[1].Name != "packages" {
					t.Errorf("second child.Name = %q, want 'packages'", root.Children[1].Name)
				}
			},
		},
		{
			name:  "empty paths",
			paths: []string{},
			check: func(t *testing.T, root *EnvTreeNode) {
				if len(root.Children) != 0 {
					t.Errorf("len(root.Children) = %d, want 0", len(root.Children))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := BuildEnvTree(tt.paths)
			tt.check(t, root)
		})
	}
}

func TestSortEnvTree(t *testing.T) {
	root := &EnvTreeNode{Name: ".", Children: []*EnvTreeNode{
		{Name: "z", File: "z"},
		{Name: "a", File: ""},
		{Name: "b", File: "b"},
	}}

	SortEnvTree(root)

	// After sorting: files first (alphabetically), then directories
	if root.Children[0].Name != "b" {
		t.Errorf("first child = %q, want 'b'", root.Children[0].Name)
	}
	if root.Children[1].Name != "z" {
		t.Errorf("second child = %q, want 'z'", root.Children[1].Name)
	}
	if root.Children[2].Name != "a" {
		t.Errorf("third child = %q, want 'a'", root.Children[2].Name)
	}
}
