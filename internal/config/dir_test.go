package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	t.Run("respects OPENENVX_CONFIG_DIR", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv(ConfigDirEnv, tmpDir)

		got := ConfigDir()
		if got != tmpDir {
			t.Errorf("ConfigDir() = %q, want %q", got, tmpDir)
		}
	})

	t.Run("uses ~/.config/openenvx when OPENENVX_CONFIG_DIR unset", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("cannot get user home dir: %v", err)
		}

		t.Setenv(ConfigDirEnv, "")

		got := ConfigDir()
		want := filepath.Join(home, ".config", ConfigSubdir)
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})

	t.Run("falls back to current directory when home not available", func(t *testing.T) {
		// Temporarily unset HOME to simulate no home directory
		oldHome := os.Getenv("HOME")
		oldUserProfile := os.Getenv("USERPROFILE")
		oldHomedrive := os.Getenv("HOMEDRIVE")
		oldHomepath := os.Getenv("HOMEPATH")

		t.Setenv(ConfigDirEnv, "")
		os.Unsetenv("HOME")
		os.Unsetenv("USERPROFILE")
		os.Unsetenv("HOMEDRIVE")
		os.Unsetenv("HOMEPATH")

		defer func() {
			os.Setenv("HOME", oldHome)
			os.Setenv("USERPROFILE", oldUserProfile)
			os.Setenv("HOMEDRIVE", oldHomedrive)
			os.Setenv("HOMEPATH", oldHomepath)
		}()

		got := ConfigDir()
		want := filepath.Join(".", ConfigSubdir)
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})
}

func TestKeysPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv(ConfigDirEnv, tmpDir)

	got := KeysPath()
	want := filepath.Join(tmpDir, KeysFileName)
	if got != want {
		t.Errorf("KeysPath() = %q, want %q", got, want)
	}
}
