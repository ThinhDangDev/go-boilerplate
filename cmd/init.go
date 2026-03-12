package cmd

import (
	"fmt"
	"strings"

	"github.com/ThinhDangDev/go-boilerplate/internal/config"
	"github.com/ThinhDangDev/go-boilerplate/internal/generator"
	"github.com/spf13/cobra"
)

var (
	name         string
	module       string
	features     []string
	configFormat string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Go backend project",
	Long: `Initialize a new production-ready Go backend project with clean architecture,
REST + gRPC support, and optional features (auth, observability, docker).`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&name, "name", "", "Project name")
	initCmd.Flags().StringVar(&module, "module", "", "Go module name")
	initCmd.Flags().StringSliceVar(&features, "features", []string{}, "Features to include (auth,observability,docker)")
	initCmd.Flags().StringVar(&configFormat, "config", "env", "Configuration format (env or yaml)")
}

func runInit(cmd *cobra.Command, args []string) error {
	var cfg *config.Config
	var err error

	// Non-interactive mode if flags are provided
	if name != "" && module != "" {
		cfg = &config.Config{
			ProjectName:  name,
			ModuleName:   module,
			OutputDir:    fmt.Sprintf("./%s", name),
			ConfigFormat: config.ConfigFormatEnv, // Default
			InitGit:      true,
		}

		// Parse config format
		switch strings.TrimSpace(strings.ToLower(configFormat)) {
		case "env", ".env":
			cfg.ConfigFormat = config.ConfigFormatEnv
		case "yaml", "yml":
			cfg.ConfigFormat = config.ConfigFormatYAML
		}

		// Parse features
		for _, f := range features {
			switch strings.TrimSpace(strings.ToLower(f)) {
			case "auth":
				cfg.Features.Auth = true
			case "observability":
				cfg.Features.Observability = true
			case "docker":
				cfg.Features.Docker = true
			}
		}
	} else {
		// Interactive mode
		cfg, err = config.PromptForConfig()
		if err != nil {
			return fmt.Errorf("failed to collect configuration: %w", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Generate project
	gen := generator.New()
	if err := gen.Generate(cfg); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Success message
	fmt.Printf("\n✓ Project '%s' generated successfully!\n\n", cfg.ProjectName)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", cfg.OutputDir)
	fmt.Println("  make proto              # Generate code from proto files")
	fmt.Println("  go mod download         # Download dependencies")
	if cfg.Features.Docker {
		fmt.Println("  docker-compose up -d    # Start PostgreSQL and other services")
	}
	fmt.Println("  make migrate-up         # Run database migrations")
	fmt.Println("  make run                # Start the server")
	fmt.Println("\nServer will start on:")
	fmt.Println("  REST: http://localhost:8080/api/v1/")
	fmt.Println("  gRPC: localhost:9090")
	if cfg.Features.Observability {
		fmt.Println("  Metrics: http://localhost:8080/metrics")
	}
	fmt.Println("  Health: http://localhost:8080/api/v1/health")
	fmt.Println()

	return nil
}
