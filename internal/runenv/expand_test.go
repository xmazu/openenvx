package runenv

import (
	"strings"
	"testing"
)

func TestExpandMap(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    map[string]string
		wantErr string
	}{
		{
			name: "no refs unchanged",
			env:  map[string]string{"A": "a", "B": "b"},
			want: map[string]string{"A": "a", "B": "b"},
		},
		{
			name: "simple ref",
			env:  map[string]string{"USERNAME": "MAC", "DB_URL": "${USERNAME}/test"},
			want: map[string]string{"USERNAME": "MAC", "DB_URL": "MAC/test"},
		},
		{
			name: "chained",
			env:  map[string]string{"C": "hi", "B": "${C}", "A": "${B}"},
			want: map[string]string{"A": "hi", "B": "hi", "C": "hi"},
		},
		{
			name:    "undefined",
			env:     map[string]string{"X": "${MISSING}"},
			wantErr: "undefined variable: MISSING",
		},
		{
			name:    "circular",
			env:     map[string]string{"A": "${B}", "B": "${A}"},
			wantErr: "circular reference",
		},
		{
			name: "empty value",
			env:  map[string]string{"EMPTY": "", "X": "${EMPTY}"},
			want: map[string]string{"EMPTY": "", "X": ""},
		},
		{
			name: "multiple refs in one value",
			env:  map[string]string{"A": "x", "B": "y", "C": "${A}-${B}"},
			want: map[string]string{"A": "x", "B": "y", "C": "x-y"},
		},
		{
			name: "empty map",
			env:  map[string]string{},
			want: map[string]string{},
		},
		{
			name: "self reference is circular",
			env:  map[string]string{"A": "${A}"},
			wantErr: "circular reference",
		},
		// Command substitution $(...)
		{
			name: "command substitution single",
			env:  map[string]string{"USER": "$(printf %s openenvx)"},
			want: map[string]string{"USER": "openenvx"},
		},
		{
			name: "command substitution with variable ref",
			env:  map[string]string{"WHO": "$(printf %s alice)", "GREET": "hello-${WHO}"},
			want: map[string]string{"WHO": "alice", "GREET": "hello-alice"},
		},
		{
			name: "command substitution multiple in one value",
			env:  map[string]string{"X": "$(printf %s a)-$(printf %s b)"},
			want: map[string]string{"X": "a-b"},
		},
		{
			name: "command substitution nested parens",
			env:  map[string]string{"OUT": `$(printf %s "$(printf %s inner)")`},
			want: map[string]string{"OUT": "inner"},
		},
		{
			name:    "command substitution empty",
			env:     map[string]string{"X": "$()"},
			wantErr: "empty $(...)",
		},
		{
			name:    "command substitution fails",
			env:     map[string]string{"X": "$(exit 1)"},
			wantErr: "command",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandMap(tt.env)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ExpandMap() err = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ExpandMap() err = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandMap() err = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Errorf("len(got) = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("got[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
