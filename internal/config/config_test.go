package config

import (
	"testing"
)

func TestHasEntry(t *testing.T) {
	c := &Config{ConfigEntries: []ConfigEntry{{Name: "alice"}}}

	if !c.HasEntry("alice") {
		t.Error("HasEntry(\"alice\"): expected true")
	}
	if c.HasEntry("bob") {
		t.Error("HasEntry(\"bob\"): expected false")
	}
}

func TestGetAuthUser_found(t *testing.T) {
	c := &Config{ConfigEntries: []ConfigEntry{{Name: "alice", AuthUser: "alice-org"}}}

	got, err := c.GetAuthUser("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "alice-org" {
		t.Errorf("got %q, want %q", got, "alice-org")
	}
}

func TestGetAuthUser_notFound(t *testing.T) {
	c := &Config{}

	_, err := c.GetAuthUser("nobody")
	if err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestLoadConfig_empty(t *testing.T) {
	t.Setenv("GH_CONFIG_DIR", t.TempDir())

	c, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(c.ConfigEntries) != 0 {
		t.Errorf("expected empty config, got %d entries", len(c.ConfigEntries))
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	t.Setenv("GH_CONFIG_DIR", t.TempDir())

	if err := makeConfigDir(); err != nil {
		t.Fatal(err)
	}

	c := &Config{
		ConfigEntries: []ConfigEntry{
			{Name: "alice", Type: IndividualType, AuthUser: "alice"},
			{Name: "bob", Type: OrganizationType, AuthUser: "bob-org"},
		},
	}
	if err := c.writeConfig(); err != nil {
		t.Fatalf("writeConfig: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig after write: %v", err)
	}
	if !loaded.HasEntry("alice") {
		t.Error("expected alice entry after round-trip")
	}
	if !loaded.HasEntry("bob") {
		t.Error("expected bob entry after round-trip")
	}
	authUser, err := loaded.GetAuthUser("bob")
	if err != nil {
		t.Fatalf("GetAuthUser: %v", err)
	}
	if authUser != "bob-org" {
		t.Errorf("got %q, want %q", authUser, "bob-org")
	}
}
