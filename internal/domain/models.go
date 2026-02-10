// Package domain contains shared domain models and types used across the application.
package domain

// Project represents a project to be scaffolded.
type Project struct {
	Language  string
	Framework string
	Name      string
	Slug      string
	Module    string
	Dir       string
	Libraries []string
}

// Library represents an optional library that can be added to a project.
type Library struct {
	Name        string
	Description string
}

// Template represents a file template to be generated.
type Template struct {
	RelativePath string
	Content      string
}

// Framework represents a project framework option.
type Framework struct {
	Language  string
	Name      string
	Templates []Template
	Generator string
	Libraries []Library
}

// Action represents a file system action to be performed.
type Action struct {
	Path    string
	Content string
}

// Plan represents the complete scaffolding plan.
type Plan struct {
	ProjectDir string
	Actions    []Action
	Generator  string
}
