package workspace

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type EnvTreeNode struct {
	Name     string
	Children []*EnvTreeNode
	File     string // Empty if directory, contains relative path if file
}

func BuildEnvTree(paths []string) *EnvTreeNode {
	root := &EnvTreeNode{Name: ".", Children: nil}

	for _, p := range paths {
		parts := strings.Split(filepath.ToSlash(p), "/")
		cur := root
		for i, part := range parts {
			if i == len(parts)-1 {
				cur.Children = append(cur.Children, &EnvTreeNode{Name: part, File: p})
				break
			}
			var next *EnvTreeNode
			for _, ch := range cur.Children {
				if ch.Name == part && ch.File == "" {
					next = ch
					break
				}
			}
			if next == nil {
				next = &EnvTreeNode{Name: part, Children: nil}
				cur.Children = append(cur.Children, next)
			}
			cur = next
		}
	}

	SortEnvTree(root)
	return root
}

func SortEnvTree(node *EnvTreeNode) {
	if len(node.Children) == 0 {
		return
	}

	sort.Slice(node.Children, func(i, j int) bool {
		ci, cj := node.Children[i], node.Children[j]
		fileI := ci.File != ""
		fileJ := cj.File != ""
		if fileI != fileJ {
			return fileI
		}
		return ci.Name < cj.Name
	})

	for _, ch := range node.Children {
		SortEnvTree(ch)
	}
}

func PrintEnvTree(node *EnvTreeNode, prefix string, last bool) {
	if node.Name != "." {
		conn := "├─ "
		if last && len(node.Children) == 0 {
			conn = "└─ "
		}
		fmt.Println(prefix + conn + node.Name)
	}

	childPrefix := prefix
	if node.Name != "." {
		if last && len(node.Children) == 0 {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}

	for i, ch := range node.Children {
		PrintEnvTree(ch, childPrefix, i == len(node.Children)-1)
	}
}

func IsEnvFilename(name string) bool {
	if name == ".env" {
		return true
	}
	if name == ".env.example" {
		return false
	}
	return strings.HasPrefix(name, ".env.") && len(name) > 5
}
