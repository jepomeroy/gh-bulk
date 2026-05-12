package execute

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecute_success(t *testing.T) {
	err := Command{CommandValue: "true"}.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecute_failure(t *testing.T) {
	err := Command{CommandValue: "false"}.Execute()
	if err == nil {
		t.Error("expected error from failing command")
	}
}

func TestExecute_sideEffect(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")

	err := Command{CommandValue: "touch " + marker}.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("marker file not created: %v", err)
	}
}
