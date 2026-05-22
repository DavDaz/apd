// Package templates loads and validates APD document templates.
package templates

// Template describes one guided document route.
type Template struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Version     int       `yaml:"version"`
	Aliases     []string  `yaml:"aliases"`
	Sections    []Section `yaml:"sections"`
}

// Section describes one ordered prompt section in a template.
type Section struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Required    bool     `yaml:"required"`
	Description string   `yaml:"description"`
	Help        string   `yaml:"help"`
	Example     string   `yaml:"example"`
	Questions   []string `yaml:"questions"`
	ContextKeys []string `yaml:"context_keys"`
}
