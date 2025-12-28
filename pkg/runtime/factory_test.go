package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ptone/scion-agent/pkg/runtime/kubernetes"
)

func TestGetRuntime(t *testing.T) {
	// Clear PATH to avoid auto-detection of local runtimes (container, docker)
	// which might override the settings-based resolution on different machines.
	t.Setenv("PATH", "")

	// Test default behavior (LoadSettings defaults to "docker")
	t.Run("Default", func(t *testing.T) {
		// Ensure we are not picking up some random settings file
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		t.Setenv("SCION_GROVE", "") // Ensure no grove path influence

		r := GetRuntime("")
		if _, ok := r.(*DockerRuntime); !ok {
			t.Errorf("expected *DockerRuntime by default (from LoadSettings), got %T", r)
		}
	})

	t.Run("Settings_Global_Container", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		globalDir := filepath.Join(tmpHome, ".scion")
		if err := os.MkdirAll(globalDir, 0755); err != nil {
			t.Fatal(err)
		}

		err := os.WriteFile(filepath.Join(globalDir, "settings.json"), 
			[]byte(`{"default_runtime": "container"}`), 0644)
		if err != nil {
			t.Fatal(err)
		}

		r := GetRuntime("")
		if _, ok := r.(*AppleContainerRuntime); !ok {
			t.Errorf("expected *AppleContainerRuntime from settings, got %T", r)
		}
	})

	t.Run("Settings_Global_Remote", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		globalDir := filepath.Join(tmpHome, ".scion")
		if err := os.MkdirAll(globalDir, 0755); err != nil {
			t.Fatal(err)
		}

		err := os.WriteFile(filepath.Join(globalDir, "settings.json"), 
			[]byte(`{"default_runtime": "remote"}`), 0644)
		if err != nil {
			t.Fatal(err)
		}

		r := GetRuntime("")
		// Remote resolves to kubernetes
		if _, ok := r.(*kubernetes.KubernetesRuntime); !ok {
			t.Errorf("expected *kubernetes.KubernetesRuntime, got %T", r)
		}
	})

	t.Run("Settings_Grove_Override", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		
		// Create a fake grove project
		grovePath := filepath.Join(tmpHome, "myproject")
		groveScionDir := filepath.Join(grovePath, ".scion")
		if err := os.MkdirAll(groveScionDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Global says container
		globalDir := filepath.Join(tmpHome, ".scion")
		if err := os.MkdirAll(globalDir, 0755); err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(globalDir, "settings.json"), []byte(`{"default_runtime": "container"}`), 0644)

		// Grove says docker
		os.WriteFile(filepath.Join(groveScionDir, "settings.json"), []byte(`{"default_runtime": "docker"}`), 0644)

		r := GetRuntime(groveScionDir)
		if _, ok := r.(*DockerRuntime); !ok {
			t.Errorf("expected *DockerRuntime from grove override, got %T", r)
		}
	})
}
