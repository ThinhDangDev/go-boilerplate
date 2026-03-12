package generator

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Check if proto files exist
	protoPath := filepath.Join(v.projectDir, "api", "proto", "v1", "health.proto")
	if _, err := os.Stat(protoPath); os.IsNotExist(err) {
		return fmt.Errorf("proto files not found")
	}

	// Note: Skipping go mod tidy and go vet because proto code must be generated first
	// Users should run: cd <project> && make proto && go mod download && make run

	return nil
}
