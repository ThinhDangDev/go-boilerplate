package config

// Config represents the generator configuration gathered from user prompts
type Config struct {
	ProjectName     string
	ModuleName      string
	OutputDir       string
	Features        Features
	InitGit         bool
	GenerateExample bool
}

// Features represents optional features to include in the generated project
type Features struct {
	Auth          bool // JWT + OAuth2 + Redis
	Observability bool // Prometheus + OpenTelemetry + slog
	Docker        bool // Dockerfile + docker-compose
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ProjectName == "" {
		return ErrProjectNameRequired
	}
	if c.ModuleName == "" {
		return ErrModuleNameRequired
	}
	return nil
}
