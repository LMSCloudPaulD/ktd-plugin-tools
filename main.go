package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

// Define flag variables
type settings struct {
	copyFlagEnabled    bool
	restartFlagEnabled bool
	installFlagEnabled bool
	copyCmd            string
	restartCmd         string
	installCmd         string
	showHelp           bool
	envFile            string
	quiet              bool
}

// Color formatting functions
var (
	green     = color.New(color.FgGreen).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	white     = color.New(color.FgWhite).SprintFunc()
	yellow    = color.New(color.FgYellow).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
)

func main() {
	set := &settings{}

	// First, parse flags
	flag.BoolVar(&set.copyFlagEnabled, "copy", false, "Enable copy command")
	flag.BoolVar(&set.copyFlagEnabled, "c", false, "(shorthand for --copy)")
	flag.BoolVar(&set.restartFlagEnabled, "restart", false, "Enable restart command")
	flag.BoolVar(&set.restartFlagEnabled, "r", false, "(shorthand for --restart)")
	flag.BoolVar(&set.installFlagEnabled, "install", false, "Enable install command")
	flag.BoolVar(&set.installFlagEnabled, "i", false, "(shorthand for --install)")
	flag.StringVar(&set.copyCmd, "copy-cmd", "", "Copy command")
	flag.StringVar(&set.installCmd, "install-cmd", "", "Install command")
	flag.StringVar(&set.restartCmd, "restart-cmd", "", "Restart command")
	flag.BoolVar(&set.showHelp, "help", false, "Show help")
	flag.BoolVar(&set.showHelp, "h", false, "(shorthand for --help)")
	flag.StringVar(&set.envFile, "env", "kpt.env", "The path to environment file")
	flag.BoolVar(&set.quiet, "quiet", false, "Silence all output")
	flag.BoolVar(&set.quiet, "q", false, "(shorthand for --quiet)")
	flag.Parse()

	// Check if flags were set
	flagsWereSet := set.copyCmd != "" || set.restartCmd != "" || set.installCmd != ""

	// Then, load environment variables if the file exists
	if _, err := os.Stat(set.envFile); !os.IsNotExist(err) {
		err := loadEnvFile(set, flagsWereSet)
		if err != nil {
			log.Fatal(err)
		}
	} else if !flagsWereSet {
		// If no flags were passed and the env file does not exist, display the help message and exit
		printHelp()
		os.Exit(1)
	}

	if set.showHelp {
		printHelp()
		return
	}

	// Define the order of execution based on dependencies
	commands := []string{}
	if set.copyFlagEnabled {
		commands = append(commands, set.copyCmd)
	}
	if set.installFlagEnabled {
		commands = append(commands, set.installCmd)
	}
	if set.restartFlagEnabled {
		commands = append(commands, set.restartCmd)
	}

	// Execute commands in the defined order
	for _, cmd := range commands {
		executeIfEnabled(cmd, set.quiet)
	}

}

func loadEnvFile(set *settings, flagsWereSet bool) error {
	// Check if a remote environment file exists and load its values
	if _, err := os.Stat(set.envFile); os.IsNotExist(err) {
		return fmt.Errorf("environment file does not exist: %v", set.envFile)
	} else if err != nil {
		return fmt.Errorf("error checking environment file: %v", err)
	}

	env, err := godotenv.Read(set.envFile)
	if err != nil {
		return fmt.Errorf("unable to read environment file: %v", err)
	}

	// Populate settings from the environment if present
	if !flagsWereSet || set.copyCmd == "" {
		set.copyCmd = envOrDefault(env, "COPY_COMMAND", set.copyCmd)
	}

	if !flagsWereSet || set.installCmd == "" {
		set.installCmd = envOrDefault(env, "INSTALL_COMMAND", set.installCmd)
	}

	if !flagsWereSet || set.restartCmd == "" {
		set.restartCmd = envOrDefault(env, "RESTART_COMMAND", set.restartCmd)
	}

	return nil
}

func envOrDefault(env map[string]string, key string, fallback string) string {
	if value, exists := env[key]; exists {
		return value
	}
	return fallback
}

func executeIfEnabled(cmd string, quiet bool) {
	if cmd == "" {
		return
	}

	if !quiet {
		fmt.Printf("Executing command: %s\n", blue(cmd))
	}

	// Execute the command synchronously and wait for its completion
	cmdParts := []string{"bash", "-c", cmd}
	execCmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	err := execCmd.Run()
	if err != nil {
		if !quiet {
			log.Fatalf("Error executing command: %s\nError: %s\n", cmd, err)
		}
		return
	}
}

func printHelp() {
	fmt.Printf("%s\n", boldGreen("ktd-plugin-tools - A helper for Koha plugin development with koha-testing-docker"))
	fmt.Printf("%s\n", white("Usage: ktd-plugin-tools [options]"))
	fmt.Printf("%s\n", white("Options:"))

	flag.VisitAll(func(f *flag.Flag) {
		optionLine := fmt.Sprintf("--%s", f.Name)
		defaultLine := ""
		if f.DefValue != "" {
			defaultLine = fmt.Sprintf("Default: %s", yellow(f.DefValue))
		}
		line := fmt.Sprintf("\t%s\t%s\t\t%s", optionLine, f.Usage, defaultLine)
		fmt.Println(line)
	})
}
