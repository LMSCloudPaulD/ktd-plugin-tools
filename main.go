package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"text/tabwriter"

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

	// First, load environment variables
	err := loadEnvFile(set)
	if err != nil {
		log.Fatal(err)
	}

	// Then, parse flags
	flag.BoolVar(&set.copyFlagEnabled, "copy", false, "Enable copy command (shorthand: -c)")
	flag.BoolVar(&set.restartFlagEnabled, "restart", false, "Enable restart command (shorthand: -r)")
	flag.BoolVar(&set.installFlagEnabled, "install", false, "Enable install command (shorthand: -i)")
	flag.StringVar(&set.copyCmd, "copy-cmd", set.copyCmd, "Copy command")
	flag.StringVar(&set.restartCmd, "restart-cmd", set.restartCmd, "Restart command")
	flag.StringVar(&set.installCmd, "install-cmd", set.installCmd, "Install command")
	flag.BoolVar(&set.showHelp, "help", false, "Show help (shorthand: -h)")
	flag.StringVar(&set.envFile, "env", set.envFile, "The path to environment file")
	flag.BoolVar(&set.quiet, "quiet", false, "Silence all output (shorthand: -q)")
	flag.Parse()

	// Flag shorthands
	flag.Visit(func(f *flag.Flag) {
		if len(f.Name) == 1 {
			switch f.Name {
			case "c":
				set.copyFlagEnabled = true
			case "r":
				set.restartFlagEnabled = true
			case "i":
				set.installFlagEnabled = true
			case "h":
				set.showHelp = true
			case "q":
				set.quiet = true
			}
		}
	})

	if set.showHelp {
		printHelp()
		return
	}

	var wg sync.WaitGroup

	// Execute commands locally
	if set.copyFlagEnabled {
		wg.Add(1)
		go executeIfEnabled(set.copyCmd, set.quiet, &wg)
	}

	if set.installFlagEnabled {
		wg.Add(1)
		go executeIfEnabled(set.installCmd, set.quiet, &wg)
	}

	if set.restartFlagEnabled {
		wg.Add(1)
		go executeIfEnabled(set.restartCmd, set.quiet, &wg)
	}

	wg.Wait()
}

func loadEnvFile(set *settings) error {
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
	set.copyCmd = envOrDefault(env, "COPY_COMMAND", set.copyCmd)
	set.restartCmd = envOrDefault(env, "RESTART_COMMAND", set.restartCmd)
	set.installCmd = envOrDefault(env, "INSTALL_COMMAND", set.installCmd)

	return nil
}

func envOrDefault(env map[string]string, key string, fallback string) string {
	if value, exists := env[key]; exists {
		return value
	}
	return fallback
}

func executeIfEnabled(cmd string, quiet bool, wg *sync.WaitGroup) {
	defer wg.Done()

	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		if !quiet {
			log.Fatalf("Error executing command: %v\nOutput:\n%s\n", err, red(string(output)))
		}
		return
	}
	if len(output) > 0 && !quiet {
		fmt.Printf("Command output:\n%s\n", green(string(output)))
	}
}

func printHelp() {
	fmt.Printf("%s\n", boldGreen("ktd-plugin-tools - A helper for Koha plugin development with koha-testing-docker"))
	fmt.Printf("%s\n", white("Usage: ktd-plugin-tools [options]"))
	fmt.Printf("%s\n", white("Options:"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	flag.VisitAll(func(f *flag.Flag) {
		optionLine := fmt.Sprintf("--%s", f.Name)
		defaultLine := ""
		if f.DefValue != "" {
			defaultLine = fmt.Sprintf("Default: %s", yellow(f.DefValue))
		}
		line := fmt.Sprintf("\t%s\t%s\t\t%s", optionLine, f.Usage, defaultLine)
		fmt.Fprintln(w, line)
	})
	w.Flush()
}
