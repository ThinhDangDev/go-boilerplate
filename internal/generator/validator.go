package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Validator validates the generated project structure
type Validator struct {
	projectDir string
}

// NewValidator creates a new Validator instance
func NewValidator(projectDir string) *Validator {
	return &Validator{projectDir: projectDir}
}

// Validate performs validation checks on the generated project
func (v *Validator) Validate() error {
	// Check if go.mod exists
	goModPath := filepath.Join(v.projectDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found")
	}

	// Check if main.go exists
	mainPath := filepath.Join(v.projectDir, "cmd", "server", "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		return fmt.Errorf("cmd/server/main.go not found")
	}

	// Run go mod tidy to fetch dependencies and create go.sum (with timeout)
	tidyCtx, tidyCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer tidyCancel()

	tidyCmd := exec.CommandContext(tidyCtx, "go", "mod", "tidy")
	tidyCmd.Dir = v.projectDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, string(output))
	}

	// Check if the project compiles (go vet) with timeout
	vetCtx, vetCancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer vetCancel()

	vetCmd := exec.CommandContext(vetCtx, "go", "vet", "./...")
	vetCmd.Dir = v.projectDir
	if output, err := vetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go vet failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
