package config

import "errors"

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrModuleNameRequired  = errors.New("module name is required")
	ErrInvalidProjectName  = errors.New("project name must contain only lowercase letters, numbers, and hyphens")
	ErrInvalidModuleName   = errors.New("module name must be a valid Go module path")
)
