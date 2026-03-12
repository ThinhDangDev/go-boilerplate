package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config with all fields",
			config: Config{
				ProjectName: "my-project",
				ModuleName:  "github.com/user/my-project",
				OutputDir:   "/tmp/test",
				Features: Features{
					Auth:          true,
					Observability: true,
					Docker:        true,
				},
				InitGit:         true,
				GenerateExample: true,
			},
			wantErr: nil,
		},
		{
			name: "valid config with minimal fields",
			config: Config{
				ProjectName: "simple-project",
				ModuleName:  "example.com/simple",
			},
			wantErr: nil,
		},
		{
			name: "invalid - missing project name",
			config: Config{
				ModuleName: "github.com/user/project",
			},
			wantErr: ErrProjectNameRequired,
		},
		{
			name: "invalid - missing module name",
			config: Config{
				ProjectName: "my-project",
			},
			wantErr: ErrModuleNameRequired,
		},
		{
			name: "invalid - empty project name",
			config: Config{
				ProjectName: "",
				ModuleName:  "github.com/user/project",
			},
			wantErr: ErrProjectNameRequired,
		},
		{
			name: "invalid - empty module name",
			config: Config{
				ProjectName: "my-project",
				ModuleName:  "",
			},
			wantErr: ErrModuleNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatures(t *testing.T) {
	tests := []struct {
		name     string
		features Features
		validate func(t *testing.T, f Features)
	}{
		{
			name: "all features enabled",
			features: Features{
				Auth:          true,
				Observability: true,
				Docker:        true,
			},
			validate: func(t *testing.T, f Features) {
				if !f.Auth {
					t.Error("Auth should be enabled")
				}
				if !f.Observability {
					t.Error("Observability should be enabled")
				}
				if !f.Docker {
					t.Error("Docker should be enabled")
				}
			},
		},
		{
			name: "no features enabled",
			features: Features{
				Auth:          false,
				Observability: false,
				Docker:        false,
			},
			validate: func(t *testing.T, f Features) {
				if f.Auth {
					t.Error("Auth should be disabled")
				}
				if f.Observability {
					t.Error("Observability should be disabled")
				}
				if f.Docker {
					t.Error("Docker should be disabled")
				}
			},
		},
		{
			name: "only auth enabled",
			features: Features{
				Auth:          true,
				Observability: false,
				Docker:        false,
			},
			validate: func(t *testing.T, f Features) {
				if !f.Auth {
					t.Error("Auth should be enabled")
				}
				if f.Observability {
					t.Error("Observability should be disabled")
				}
				if f.Docker {
					t.Error("Docker should be disabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.features)
		})
	}
}
