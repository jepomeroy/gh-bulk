package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/cli/go-gh/v2/pkg/config"
	"gopkg.in/yaml.v3"
)

const (
	IndividualType UserType = iota
	OrganizationType
)

type UserType int

type ConfigEntry struct {
	Name     string   `yaml:"name"`
	Type     UserType `yaml:"type"`
	AuthUser string   `yaml:"authUser"`
}

type Config struct {
	ConfigEntries []ConfigEntry
}

func LoadConfig() (*Config, error) {
	// Load the config from the config file
	entries, err := readConfig()

	if err != nil {
		return nil, err
	}

	return &Config{ConfigEntries: entries}, nil
}

func readConfig() ([]ConfigEntry, error) {
	// Make sure the config directory exists
	err := makeConfigDir()
	if err != nil {
		return []ConfigEntry{}, err
	}

	// Read the config file
	data, err := os.ReadFile(filepath.Join(config.ConfigDir(), "gh-bulk", "config.yaml"))
	if err != nil {
		return []ConfigEntry{}, nil
	}

	// Unmarshal the config file
	var config []ConfigEntry
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return []ConfigEntry{}, err
	}

	return config, nil
}

func (c *Config) AddEntry(entryName string) (string, error) {
	// prompt the user for the entry type and auth user
	configEntry, err := makeEntry(entryName)
	if err != nil {
		return "", err
	}

	c.ConfigEntries = append(c.ConfigEntries, configEntry)

	// write the config to the config file
	err = c.writeConfig()
	if err != nil {
		return "", err
	}

	return configEntry.AuthUser, nil
}

func (c *Config) HasEntry(entryName string) bool {
	for _, entry := range c.ConfigEntries {
		if entry.Name == entryName {
			return true
		}
	}

	return false
}

func (c *Config) writeConfig() error {
	// Marshal the config entries
	data, err := yaml.Marshal(c.ConfigEntries)
	if err != nil {
		return err
	}

	// Write the config file
	err = os.WriteFile(filepath.Join(config.ConfigDir(), "gh-bulk", "config.yaml"), data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func makeConfigDir() error {
	return os.MkdirAll(filepath.Join(config.ConfigDir(), "gh-bulk"), 0755)
}

func makeEntry(entryName string) (ConfigEntry, error) {
	var entryType UserType
	var authUser string

	// prompt the user for the entry type and auth user
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

func (c *Config) GetAuthUser(entryName string) (string, error) {
	for _, entry := range c.ConfigEntries {
		if entry.Name == entryName {
			return entry.AuthUser, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Entry %s not found", entryName))
}
