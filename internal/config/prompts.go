package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

var (
	projectNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
	moduleNameRegex  = regexp.MustCompile(`^[a-z0-9.-]+/[a-z0-9.-/]+$`)
)

// PromptForConfig interactively collects configuration from the user
func PromptForConfig() (*Config, error) {
	cfg := &Config{}

	// Project Name
	projectNamePrompt := &survey.Input{
		Message: "Project name:",
		Help:    "Name for your project directory (lowercase, numbers, hyphens only)",
		Default: "my-backend",
	}
	if err := survey.AskOne(projectNamePrompt, &cfg.ProjectName, survey.WithValidator(validateProjectName)); err != nil {
		return nil, fmt.Errorf("project name prompt failed: %w", err)
	}

	// Module Name
	moduleNamePrompt := &survey.Input{
		Message: "Go module name:",
		Help:    "Go module path (e.g., github.com/yourusername/project)",
		Default: fmt.Sprintf("github.com/yourusername/%s", cfg.ProjectName),
	}
	if err := survey.AskOne(moduleNamePrompt, &cfg.ModuleName, survey.WithValidator(validateModuleName)); err != nil {
		return nil, fmt.Errorf("module name prompt failed: %w", err)
	}

	// Output Directory
	outputDirPrompt := &survey.Input{
		Message: "Output directory:",
		Help:    "Where to generate the project",
		Default: fmt.Sprintf("./%s", cfg.ProjectName),
	}
	if err := survey.AskOne(outputDirPrompt, &cfg.OutputDir); err != nil {
		return nil, fmt.Errorf("output directory prompt failed: %w", err)
	}

	// Feature: Authentication
	authPrompt := &survey.Confirm{
		Message: "Enable Authentication? (JWT + OAuth2 + Redis)",
		Default: false,
	}
	if err := survey.AskOne(authPrompt, &cfg.Features.Auth); err != nil {
		return nil, fmt.Errorf("auth feature prompt failed: %w", err)
	}

	// Feature: Observability
	observabilityPrompt := &survey.Confirm{
		Message: "Enable Full Observability? (Prometheus + OpenTelemetry + slog)",
		Default: false,
	}
	if err := survey.AskOne(observabilityPrompt, &cfg.Features.Observability); err != nil {
		return nil, fmt.Errorf("observability feature prompt failed: %w", err)
	}

	// Feature: Docker
	dockerPrompt := &survey.Confirm{
		Message: "Include Docker Environment? (Dockerfile + docker-compose)",
		Default: true,
	}
	if err := survey.AskOne(dockerPrompt, &cfg.Features.Docker); err != nil {
		return nil, fmt.Errorf("docker feature prompt failed: %w", err)
	}

	// Example CRUD
	examplePrompt := &survey.Confirm{
		Message: "Generate Example CRUD? (User resource)",
		Default: false,
	}
	if err := survey.AskOne(examplePrompt, &cfg.GenerateExample); err != nil {
		return nil, fmt.Errorf("example CRUD prompt failed: %w", err)
	}

	// Initialize Git
	gitPrompt := &survey.Confirm{
		Message: "Initialize Git repository?",
		Default: true,
	}
	if err := survey.AskOne(gitPrompt, &cfg.InitGit); err != nil {
		return nil, fmt.Errorf("git init prompt failed: %w", err)
	}

	return cfg, nil
}

func validateProjectName(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return ErrProjectNameRequired
	}
	if !projectNameRegex.MatchString(str) {
		return ErrInvalidProjectName
	}
	return nil
}

func validateModuleName(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return ErrModuleNameRequired
	}
	if !moduleNameRegex.MatchString(str) {
		return ErrInvalidModuleName
	}
	return nil
}
