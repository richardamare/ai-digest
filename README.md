# AI Digest 📚

AI Digest is a powerful command-line tool that aggregates your codebase into a single markdown file, making it easier to share and discuss your code with AI assistants. It intelligently handles both text and binary files, respects ignore patterns, and provides various formatting options.

## Features ✨

- **Smart File Processing**
    - Handles both text and binary files
    - Automatically detects file types
    - Preserves whitespace in sensitive files (Python, YAML, etc.)
    - Supports SVG files with special handling

- **Flexible Configuration**
    - Customizable ignore patterns
    - Default ignore patterns for common development files
    - Support for `.aidigestignore` file (similar to .gitignore)
    - JSON-based configuration file

- **Performance Optimized**
    - Concurrent file processing
    - Efficient memory usage
    - Handles large codebases

- **Developer Friendly**
    - Detailed processing statistics
    - Custom file listings

## Installation 🚀

### Using Go Install
```bash
go install github.com/richardamare/ai-digest@latest
```

### From Source
```bash
git clone https://github.com/richardamare/ai-digest.git
cd ai-digest
go build
```

## Usage 💡

### Basic Usage
```bash
# Process current directory
ai-digest digest

# Process specific directory
ai-digest digest -i /path/to/project -o output.md
```

### Advanced Options
```bash
# Remove unnecessary whitespace
ai-digest digest --whitespace-removal

# Show list of processed files
ai-digest digest --show-output-files

# Use custom ignore file
ai-digest digest --ignore-file .customignore

# Disable default ignore patterns
ai-digest digest --no-default-ignores
```

### Configuration Management
```bash
# Initialize config file
ai-digest config init

# Show current configuration
ai-digest config show
```

### Version Information
```bash
ai-digest version
```

## Configuration File 📝

AI Digest uses a JSON configuration file (`ai-digest.json`) for persistent settings:

```json
{
  "defaultIgnores": [
    "node_modules",
    ".git",
    "*.log",
    "*.swp",
    ".DS_Store",
    "Thumbs.db",
    "*.tmp",
    "*.temp",
    ".idea",
    ".vscode"
  ],
  "ignoreFile": ".aidigestignore"
}
```

## Ignore File Format 🚫

Create a `.aidigestignore` file in your project root to specify files and directories to ignore:

```gitignore
# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/

# IDE files
.vscode/
.idea/

# Logs
*.log

# Environment files
.env*
```

## Output Format 📄

The generated markdown file includes:
- File paths as headers
- Language-specific syntax highlighting
- Special handling for binary files
- Preserved formatting for whitespace-sensitive files

Example output:
```markdown
# main.go

package main

func main() {
    // Your code here
}

# image.png

This is a binary file of the type: Image
```

## Contributing 🤝

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License 📜

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support 💬

If you encounter any problems or have suggestions, please [open an issue](https://github.com/richardamare/ai-digest/issues) on GitHub.