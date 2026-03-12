package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileWriter_WriteFile(t *testing.T) {
	fw := NewFileWriter()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (path string, data []byte, cleanup func())
		wantErr bool
		verify  func(t *testing.T, path string)
	}{
		{
			name: "write to new file",
			setup: func(t *testing.T) (string, []byte, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.txt")
				data := []byte("hello world")
				return path, data, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "hello world" {
					t.Errorf("file content = %s, want %s", string(content), "hello world")
				}
			},
		},
		{
			name: "write to nested directory",
			setup: func(t *testing.T) (string, []byte, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "a", "b", "c", "test.txt")
				data := []byte("nested content")
				return path, data, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "nested content" {
					t.Errorf("file content = %s, want %s", string(content), "nested content")
				}
			},
		},
		{
			name: "overwrite existing file",
			setup: func(t *testing.T) (string, []byte, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.txt")
				// Create initial file
				if err := os.WriteFile(path, []byte("old content"), 0644); err != nil {
					t.Fatal(err)
				}
				data := []byte("new content")
				return path, data, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "new content" {
					t.Errorf("file content = %s, want %s", string(content), "new content")
				}
			},
		},
		{
			name: "write empty file",
			setup: func(t *testing.T) (string, []byte, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "empty.txt")
				data := []byte("")
				return path, data, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, path string) {
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("failed to stat file: %v", err)
				}
				if info.Size() != 0 {
					t.Errorf("file size = %d, want 0", info.Size())
				}
			},
		},
		{
			name: "write large file",
			setup: func(t *testing.T) (string, []byte, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "large.txt")
				// Create 1MB of data
				data := make([]byte, 1024*1024)
				for i := range data {
					data[i] = byte(i % 256)
				}
				return path, data, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, path string) {
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("failed to stat file: %v", err)
				}
				if info.Size() != 1024*1024 {
					t.Errorf("file size = %d, want %d", info.Size(), 1024*1024)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, data, cleanup := tt.setup(t)
			defer cleanup()

			err := fw.WriteFile(path, data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileWriter.WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, path)
			}
		})
	}
}

func TestFileWriter_Permissions(t *testing.T) {
	fw := NewFileWriter()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	data := []byte("test")

	err := fw.WriteFile(path, data)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	// Check file permissions (should be 0644)
	mode := info.Mode()
	expectedMode := os.FileMode(0644)
	if mode.Perm() != expectedMode {
		t.Errorf("file permissions = %v, want %v", mode.Perm(), expectedMode)
	}
}

func TestFileWriter_AtomicWrite(t *testing.T) {
	fw := NewFileWriter()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "atomic.txt")

	// Write initial content
	err := fw.WriteFile(path, []byte("initial"))
	if err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	// Verify initial content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "initial" {
		t.Errorf("initial content = %s, want %s", string(content), "initial")
	}

	// Write new content
	err = fw.WriteFile(path, []byte("updated"))
	if err != nil {
		t.Fatalf("failed to write updated file: %v", err)
	}

	// Verify updated content
	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "updated" {
		t.Errorf("updated content = %s, want %s", string(content), "updated")
	}

	// Verify no temporary files left behind
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".tmp" || f.Name() == ".tmp-*" {
			t.Errorf("temporary file found: %s", f.Name())
		}
	}
}
