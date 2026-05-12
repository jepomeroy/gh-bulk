// Package config manages persistent per-user configuration for gh-bulk.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/cli/go-gh/v2/pkg/config"
	"gopkg.in/yaml.v3"
)

const (
	// IndividualType indicates the user operates as an individual GitHub user.
	IndividualType UserType = iota
	// OrganizationType indicates the user operates as a GitHub organization.
	OrganizationType
)

// UserType distinguishes whether the authenticated user operates as an individual or an organization.
type UserType int

// ConfigEntry represents a single user's configuration in the gh-bulk config file.
type ConfigEntry struct {
	Name     string   `yaml:"name"`
	Type     UserType `yaml:"type"`
	AuthUser string   `yaml:"authUser"`
}

// Config holds all configuration entries loaded from the gh-bulk config file.
type Config struct {
	ConfigEntries []ConfigEntry
}

// LoadConfig reads the gh-bulk configuration from disk and returns it.
func LoadConfig() (*Config, error) {
	entries, err := readConfig()
	if err != nil {
		return nil, err
	}

	return &Config{ConfigEntries: entries}, nil
}

func readConfig() ([]ConfigEntry, error) {
	err := makeConfigDir()
	if err != nil {
		return []ConfigEntry{}, err
	}

	data, err := os.ReadFile(filepath.Join(config.ConfigDir(), "gh-bulk", "config.yaml"))
	if err != nil {
		return []ConfigEntry{}, nil
	}

	var config []ConfigEntry
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return []ConfigEntry{}, err
	}

	return config, nil
}

// AddEntry prompts for a new config entry for entryName, appends it, and writes the config to disk.
func (c *Config) AddEntry(entryName string) (string, error) {
	configEntry, err := makeEntry(entryName)
	if err != nil {
		return "", err
	}

	c.ConfigEntries = append(c.ConfigEntries, configEntry)

	err = c.writeConfig()
	if err != nil {
		return "", err
	}

	return configEntry.AuthUser, nil
}

// HasEntry reports whether a config entry with the given name exists.
func (c *Config) HasEntry(entryName string) bool {
	for _, entry := range c.ConfigEntries {
		if entry.Name == entryName {
			return true
		}
	}

	return false
}

func (c *Config) writeConfig() error {
	data, err := yaml.Marshal(c.ConfigEntries)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(config.ConfigDir(), "gh-bulk", "config.yaml"), data, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func makeConfigDir() error {
	return os.MkdirAll(filepath.Join(config.ConfigDir(), "gh-bulk"), 0o755)
}

func makeEntry(entryName string) (ConfigEntry, error) {
	var entryType UserType
	var authUser string

	fmt.Printf("Current GitHub User: %s\n\n", entryName)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[UserType]().
				Title("Select User type").
				Options(
					huh.NewOption("Individual", IndividualType),
					huh.NewOption("Organization", OrganizationType),
				).
				Value(&entryType),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Enter Organization name").
				Placeholder("Org name").
				Value(&authUser),
		).WithHideFunc(func() bool { return entryType == IndividualType }),
	).WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return ConfigEntry{}, err
	}

	if entryType == IndividualType {
		authUser = entryName
	}

	return ConfigEntry{
		Name:     entryName,
		Type:     entryType,
		AuthUser: authUser,
	}, nil
}

// GetAuthUser returns the authUser value for the config entry matching entryName.
func (c *Config) GetAuthUser(entryName string) (string, error) {
	for _, entry := range c.ConfigEntries {
		if entry.Name == entryName {
			return entry.AuthUser, nil
		}
	}

	return "", fmt.Errorf("entry %s not found", entryName)
}
