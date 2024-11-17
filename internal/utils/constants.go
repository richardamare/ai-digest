package utils

// WhitespaceDependentExtensions defines file extensions where whitespace is significant
var WhitespaceDependentExtensions = map[string]bool{
	".py":     true, // Python
	".yaml":   true, // YAML
	".yml":    true, // YAML
	".jade":   true, // Jade/Pug
	".haml":   true, // Haml
	".slim":   true, // Slim
	".coffee": true, // CoffeeScript
	".pug":    true, // Pug
	".styl":   true, // Stylus
	".gd":     true, // Godot
}

// DefaultIgnores defines patterns to ignore by default
var DefaultIgnores = []string{
	// Node.js
	"node_modules",
	"package-lock.json",
	"npm-debug.log",
	// Yarn
	"yarn.lock",
	"yarn-error.log",
	// pnpm
	"pnpm-lock.yaml",
	// Bun
	"bun.lockb",
	// Deno
	"deno.lock",
	// PHP
	"vendor",
	"composer.lock",
	// Python
	"__pycache__",
	"*.pyc",
	"*.pyo",
	"*.pyd",
	".Python",
	"pip-log.txt",
	"pip-delete-this-directory.txt",
	".venv",
	"venv",
	"ENV",
	"env",
	// Godot
	".godot",
	"*.import",
	// Ruby
	"Gemfile.lock",
	".bundle",
	// Java
	"target",
	"*.class",
	// Gradle
	".gradle",
	"build",
	// Maven
	"pom.xml.tag",
	"pom.xml.releaseBackup",
	"pom.xml.versionsBackup",
	"pom.xml.next",
	// .NET
	"bin",
	"obj",
	"*.suo",
	"*.user",
	// Go
	"go.sum",
	// Rust
	"Cargo.lock",
	"target",
	// General
	".git",
	".svn",
	".hg",
	".DS_Store",
	"Thumbs.db",
	// Environment variables
	".env",
	".env.local",
	".env.development.local",
	".env.test.local",
	".env.production.local",
	"*.env",
	"*.env.*",
	// Common framework directories
	".svelte-kit",
	".next",
	".nuxt",
	".vuepress",
	".cache",
	"dist",
	"tmp",
	// Output file
	"codebase.md",
	// Turborepo
	".turbo",
}
