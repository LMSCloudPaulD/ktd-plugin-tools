package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

// Define flag variables
var (
	copyFlagEnabled    bool
	restartFlagEnabled bool
	installFlagEnabled bool
	copyCmd            string
	restartCmd         string
	installCmd         string
	showHelp           bool
	envFile            string

	// Color formatting functions
	green     = color.New(color.FgGreen).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	white     = color.New(color.FgWhite).SprintFunc()
	yellow    = color.New(color.FgYellow).SprintFunc()
)

func main() {
	flag.BoolVar(&copyFlagEnabled, "copy", true, "Enable copy command")
	flag.BoolVar(&restartFlagEnabled, "restart", true, "Enable restart command")
	flag.BoolVar(&installFlagEnabled, "install", true, "Enable install command")
	flag.StringVar(&copyCmd, "copy-cmd", "docker cp Koha koha-koha-1:/var/lib/koha/kohadev/plugins/", "Copy command")
	flag.StringVar(&restartCmd, "restart-cmd", "docker exec -ti koha-koha-1 bash -c 'koha-plack --restart kohadev'", "Restart command")
	flag.StringVar(&installCmd, "install-cmd", "docker exec -ti koha-koha-1 /kohadevbox/koha/misc/devel/install_plugins.pl", "Install command")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.StringVar(&envFile, "env", "kpt.env", "The path to environment file")
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	// Check if a remote environment file exists and load its values
	err := loadEnvFile(envFile)

	if err != nil {
		log.Fatal(err)
	}

	// Execute commands locally
	if copyFlagEnabled {
		if copyCmd != "" {
			executeCommand(copyCmd)
		}
	}

	if installFlagEnabled {
		if installCmd != "" {
			executeCommand(installCmd)
		}
	}

	if restartFlagEnabled {
		if restartCmd != "" {
			executeCommand(restartCmd)
		}
	}
}

func loadEnvFile(file string) error {
	// Check if a remote environment file exists and load its values
	if _, err := os.Stat(file); err == nil {
		env, err := godotenv.Read(file)
		if err != nil {
			return fmt.Errorf("unable to read environment file: %v", err)
		}

		// Populate global variables from the environment if present
		copyCmd = env["COPY_COMMAND"]
		restartCmd = env["RESTART_COMMAND"]
		installCmd = env["INSTALL_COMMAND"]

	} else {
		return fmt.Errorf("environment file does not exist: %v", file)
	}
	return nil
}

func executeCommand(cmd string) {
	fmt.Printf("Executing command: %s\n", blue(cmd))
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
	fmt.Printf("Command output:\n%s\n", green(output))
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
