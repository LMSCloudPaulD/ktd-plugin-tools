package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/LMSCloudPaulD/ktd-plugin-tools/pkg/archiving"
	"github.com/LMSCloudPaulD/ktd-plugin-tools/pkg/versioning"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	CopyCmd    string
	InstallCmd string
	RestartCmd string
}

type Flags struct {
	Copy    bool
	Install bool
	Restart bool
}

func main() {
	var config Config
	var flags Flags

	rootCmd := &cobra.Command{
		Use:   "koha-plugin-tools",
		Short: "A helper for Koha plugin development with koha-testing-docker",
	}

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy commands",
		Run: func(cmd *cobra.Command, args []string) {
			if err := deploy(config, flags); err != nil {
				fmt.Printf("Error deploying: %v\n", err)
				os.Exit(1)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "bump {major, minor, patch}",
		Short: "Increment version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := bump(args); err != nil {
				fmt.Printf("Error bumping: %v\n", err)
				os.Exit(1)
			}
		},
	}

	archiveCmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive all but the latest kpz bundle in a dir",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := archive(); err != nil {
				fmt.Printf("Error archiving: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVarP(&flags.Copy, "copy", "c", false, "Enable copy command")
	deployCmd.Flags().BoolVarP(&flags.Install, "install", "i", false, "Enable install command")
	deployCmd.Flags().BoolVarP(&flags.Restart, "restart", "r", false, "Enable restart command")

	rootCmd.AddCommand(versionCmd)

	rootCmd.AddCommand(archiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func deploy(config Config, flags Flags) error {
	if err := initConfig(&config); err != nil {
		return fmt.Errorf("Error initializing config: %w", err)
	}

	if flags.Copy {
		if err := executeCommand(config.CopyCmd); err != nil {
			return fmt.Errorf("Error executing copy command: %w", err)
		}
	}
	if flags.Install {
		if err := executeCommand(config.InstallCmd); err != nil {
			return fmt.Errorf("Error executing install command: %w", err)
		}
	}
	if flags.Restart {
		if err := executeCommand(config.RestartCmd); err != nil {
			return fmt.Errorf("Error executing restart command: %w", err)
		}
	}
	return nil
}

func bump(args []string) error {
	updateType := args[0]

	manager := &versioning.VersionManager{}
	err := manager.BumpVersion("package.json", updateType)
	if err != nil {
		return fmt.Errorf("Error during version bump: %w", err)
	}

	return nil
}

func archive() error {
	archiver := &archiving.Archiver{}
	err := archiver.ArchiveFiles(`-v(\d+\.\d+\.\d+)\.kpz`, "archive")
	if err != nil {
		return fmt.Errorf("Error archiving files: %w", err)
	}

	return nil
}

func executeCommand(cmd string) error {
	fmt.Printf("Executing command: %s\n", cmd)

	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error executing command: %v\nOutput was: %s", err, out)
	}
	fmt.Printf("Command output:\n%s\n", out)
	return nil
}

func initConfig(config *Config) error {
	viper.SetConfigName("deploy") // name of config file (without extension)
	viper.AddConfigPath(".")      // look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Error reading config file: %w", err)
	}

	config.CopyCmd = viper.GetString("copyCmd")
	config.InstallCmd = viper.GetString("installCmd")
	config.RestartCmd = viper.GetString("restartCmd")
	return nil
}
