package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

// Test for deploy function
func TestDeploy(t *testing.T) {
	// Mock the Config
	config := Config{
		CopyCmd:    "echo copy",
		InstallCmd: "echo install",
		RestartCmd: "echo restart",
	}

	t.Run("All flags set", func(t *testing.T) {
		flags := Flags{
			Copy:    true,
			Install: true,
			Restart: true,
		}

		err := deploy(config, flags)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("No flags set", func(t *testing.T) {
		flags := Flags{
			Copy:    false,
			Install: false,
			Restart: false,
		}

		err := deploy(config, flags)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Some flags set", func(t *testing.T) {
		flags := Flags{
			Copy:    true,
			Install: false,
			Restart: true,
		}

		err := deploy(config, flags)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Empty command", func(t *testing.T) {
		config := Config{
			CopyCmd:    "",
			InstallCmd: "echo install",
			RestartCmd: "echo restart",
		}
		flags := Flags{
			Copy:    true,
			Install: true,
			Restart: true,
		}

		err := deploy(config, flags)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Invalid command", func(t *testing.T) {
		config := Config{
			CopyCmd:    "invalid_command",
			InstallCmd: "echo install",
			RestartCmd: "echo restart",
		}
		flags := Flags{
			Copy:    true,
			Install: true,
			Restart: true,
		}

		err := deploy(config, flags)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

// Test for initConfig function
func TestInitConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	text := []byte("copyCmd: echo copy\ninstallCmd: echo install\nrestartCmd: echo restart")
	if _, err := tmpfile.Write(text); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	t.Run("Config file exists", func(t *testing.T) {
		config := &Config{}
		viper.SetConfigFile(tmpfile.Name())
		err := initConfig(config)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Config file does not exist", func(t *testing.T) {
		config := &Config{}
		viper.SetConfigFile("nonexistent.yaml")
		err := initConfig(config)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

// Test for executeCommand function
func TestExecuteCommand(t *testing.T) {
	t.Run("Valid command", func(t *testing.T) {
		err := executeCommand("echo Hello, world!")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Invalid command", func(t *testing.T) {
		err := executeCommand("invalid_command")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Empty command", func(t *testing.T) {
		err := executeCommand("")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
