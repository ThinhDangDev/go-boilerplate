package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns project directory
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid project structure",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Create go.mod
				goModContent := `module example.com/test

go 1.22
`
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create main.go
				mainDir := filepath.Join(tmpDir, "cmd", "server")
				if err := os.MkdirAll(mainDir, 0755); err != nil {
					t.Fatal(err)
				}

				mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}
`
				if err := os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(mainContent), 0644); err != nil {
					t.Fatal(err)
				}

				return tmpDir
			},
			wantErr: false,
		},
		{
			name: "missing go.mod",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Create main.go but no go.mod
				mainDir := filepath.Join(tmpDir, "cmd", "server")
				if err := os.MkdirAll(mainDir, 0755); err != nil {
					t.Fatal(err)
				}

				mainContent := `package main

func main() {}
`
				if err := os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(mainContent), 0644); err != nil {
					t.Fatal(err)
				}

				return tmpDir
			},
			wantErr: true,
			errMsg:  "go.mod not found",
		},
		{
			name: "missing main.go",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Create go.mod but no main.go
				goModContent := `module example.com/test

go 1.22
`
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
					t.Fatal(err)
				}

				return tmpDir
			},
			wantErr: true,
			errMsg:  "cmd/server/main.go not found",
		},
		{
			name: "invalid go code syntax",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Create go.mod
				goModContent := `module example.com/test

go 1.22
`
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create main.go with syntax error
				mainDir := filepath.Join(tmpDir, "cmd", "server")
				if err := os.MkdirAll(mainDir, 0755); err != nil {
					t.Fatal(err)
				}

				mainContent := `package main

func main() {
	// Missing closing brace
`
				if err := os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(mainContent), 0644); err != nil {
					t.Fatal(err)
				}

				return tmpDir
			},
			wantErr: true,
			errMsg:  "go vet failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := tt.setup(t)
			validator := NewValidator(projectDir)

			err := validator.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validator.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validator.Validate() error = %v, want error containing %s", err, tt.errMsg)
				}
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	projectDir := "/tmp/test"
	validator := NewValidator(projectDir)

	if validator == nil {
		t.Fatal("NewValidator() returned nil")
	}

	if validator.projectDir != projectDir {
		t.Errorf("NewValidator() projectDir = %s, want %s", validator.projectDir, projectDir)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
