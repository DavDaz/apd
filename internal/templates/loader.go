package templates

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Load decodes and validates one YAML template.
func Load(r io.Reader) (Template, error) {
	var t Template
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&t); err != nil {
		return Template{}, fmt.Errorf("decode template: %w", err)
	}
	return t, Validate(t)
}

// LoadFile reads and validates one YAML template file from a filesystem.
func LoadFile(fsys fs.FS, name string) (Template, error) {
	data, err := fs.ReadFile(fsys, filepath.ToSlash(name))
	if err != nil {
		return Template{}, fmt.Errorf("read template %q: %w", name, err)
	}
	t, err := Load(bytes.NewReader(data))
	if err != nil {
		return Template{}, fmt.Errorf("load template %q: %w", name, err)
	}
	return t, nil
}

// LoadDir reads every .yaml template in dir in filename order.
func LoadDir(fsys fs.FS, dir string) ([]Template, error) {
	entries, err := fs.ReadDir(fsys, filepath.ToSlash(dir))
	if err != nil {
		return nil, fmt.Errorf("read template directory %q: %w", dir, err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			names = append(names, filepath.ToSlash(filepath.Join(dir, entry.Name())))
		}
	}
	sort.Strings(names)
	items := make([]Template, 0, len(names))
	for _, name := range names {
		t, err := LoadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}
