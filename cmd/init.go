package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/alecthomas/kong"
	"github.com/legnoh/apple-calendar-server/pkg/applecalendar"
)

const DefaultConfigPath = "~/.apple-calendar-server/config.yml"

// InitCmd sets up a sample YAML configuration file.
// It writes to the path provided by --config-file (global flag) or a custom --output.
// If the file exists, it fails unless --force is specified.
// We keep Calendars sample minimal.
type InitCmd struct {
	Output string `help:"Output file path (uses --config-file value if not specified)" name:"output"`
	Force  bool   `help:"Overwrite existing files"`
}

const configTemplate = `# apple-calendar-server configuration
serve:
  # HTTP listen address (can be overridden by serve command)
  listen: ":8080"
  
  # Calendar names (auto-detected)
  calendars:
{{- range .Calendars}}
    - {{.}}
{{- end}}
`

// generateSampleConfig creates YAML configuration with detected calendars
func generateSampleConfig() (string, error) {
	client := applecalendar.New()
	defer client.Close()

	calendars, err := client.GetCalendarList()
	if err != nil {
		// Use fallback calendars on error
		calendars = []string{"Work", "Personal"}
	}

	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse config template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		Calendars []string
	}{
		Calendars: calendars,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute config template: %w", err)
	}

	return buf.String(), nil
}

func (c *InitCmd) Run() error {
	targetPath := c.Output
	if targetPath == "" {
		targetPath = DefaultConfigPath
	}
	targetPath = kong.ExpandPath(targetPath)

	// Create directory if needed.
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if !c.Force {
		if fi, err := os.Stat(targetPath); err == nil && !fi.IsDir() {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", targetPath)
		} else if err == nil && fi.IsDir() {
			return fmt.Errorf("target path is a directory: %s", targetPath)
		}
	}

	// Generate configuration with actual calendars
	config, err := generateSampleConfig()
	if err != nil {
		return fmt.Errorf("failed to generate sample config: %w", err)
	}

	// Write file atomically-ish.
	tmp := fmt.Sprintf("%s.tmp-%d", targetPath, time.Now().UnixNano())
	if err := os.WriteFile(tmp, []byte(config), fs.FileMode(0o644)); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmp, targetPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename failed: %w", err)
	}

	fmt.Printf("Configuration file created: %s\n", targetPath)
	return nil
}
