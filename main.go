package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
)

// SSHConfig represents the SSH configuration parameters.
type SSHConfig struct {
	Host         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

// Define flag variables
var (
	copyFlagEnabled    bool
	restartFlagEnabled bool
	installFlagEnabled bool
	copyCmd            string
	restartCmd         string
	installCmd         string
	sshHost            string
	username           string
	privateKey         string
	sshPort            string
	sshConfigHost      string
	remoteActive       bool
	showHelp           bool
	envFile            string

	// Color formatting functions
	green     = color.New(color.FgGreen).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	white     = color.New(color.FgWhite).SprintFunc()
	yellow    = color.New(color.FgYellow).SprintFunc()
	// red       = color.New(color.FgRed).SprintFunc()
)

func main() {
	flag.BoolVar(&copyFlagEnabled, "copy", true, "Enable copy command")
	flag.BoolVar(&restartFlagEnabled, "restart", true, "Enable restart command")
	flag.BoolVar(&installFlagEnabled, "install", true, "Enable install command")
	flag.StringVar(&copyCmd, "copy-cmd", "docker cp Koha koha-koha-1:/var/lib/koha/kohadev/plugins/", "Copy command")
	flag.StringVar(&restartCmd, "restart-cmd", "docker exec -ti koha-koha-1 bash -c 'koha-plack --restart kohadev'", "Restart command")
	flag.StringVar(&installCmd, "install-cmd", "docker exec -ti koha-koha-1 /kohadevbox/koha/misc/devel/install_plugins.pl", "Install command")
	flag.StringVar(&sshHost, "ssh-host", "", "The SSH host")
	flag.StringVar(&username, "username", "", "The SSH username")
	flag.StringVar(&privateKey, "private-key", "", "The path to SSH private key")
	flag.StringVar(&sshPort, "ssh-port", "22", "The SSH port")
	flag.StringVar(&sshConfigHost, "ssh-config-host", "", "The SSH host from SSH config file")
	flag.BoolVar(&remoteActive, "remote", false, "Execute command remotely")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.StringVar(&envFile, "env", "remote.env", "The path to environment file")
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

	if sshConfigHost != "" {
		// Get SSH configuration from the SSH config file
		sshConfig, err := GetSSHConfig(sshConfigHost)
		if err != nil {
			log.Fatalf("Error getting SSH config: %v", err)
		}
		sshHost = sshConfig.HostName
		username = sshConfig.User
		sshPort = sshConfig.Port
		privateKey = sshConfig.IdentityFile
	}

	if remoteActive {
		// Execute commands remotely
		if sshHost != "" && username != "" && privateKey != "" {
			if copyFlagEnabled {
				// First, copy the file to the remote host
				err := scpFileToRemote(sshHost, username, privateKey, "Koha", "/tmp/Koha")
				if err != nil {
					log.Fatalf("Error copying file to remote host: %v", err)
				}

				// Modify the copyCmd to be consistent with local execution
				copyCmd = "docker cp /tmp/Koha koha-koha-1:/var/lib/koha/kohadev/plugins/"

				if copyCmd != "" {
					err := executeSSHCommand(sshHost, username, privateKey, copyCmd)
					if err != nil {
						log.Fatalf("Error executing copy command: %v", err)
					}
				}
			}

			if installFlagEnabled {
				err := executeSSHCommand(sshHost, username, privateKey, installCmd)
				if err != nil {
					log.Fatalf("Error executing install command: %v", err)
				}
			}

			if restartFlagEnabled {
				err := executeSSHCommand(sshHost, username, privateKey, restartCmd)
				if err != nil {
					log.Fatalf("Error executing restart command: %v", err)
				}
			}
		} else {
			log.Fatal("SSH host, username, and private key must be provided for remote operation")
		}
	} else {
		// Execute commands locally
		if copyFlagEnabled {
			if copyCmd != "" {
				executeLocalCommand(copyCmd)
			}
		}

		if installFlagEnabled {
			if installCmd != "" {
				executeLocalCommand(installCmd)
			}
		}

		if restartFlagEnabled {
			if restartCmd != "" {
				executeLocalCommand(restartCmd)
			}
		}
	}
}

func expandHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return strings.Replace(path, "~", usr.HomeDir, 1), nil
}

func loadEnvFile(file string) error {
	// Check if a remote environment file exists and load its values
	if _, err := os.Stat(file); err == nil {
		env, err := godotenv.Read(file)
		if err != nil {
			return fmt.Errorf("unable to read environment file: %v", err)
		}

		// Populate global variables from the environment if present
		sshHost = env["REMOTE_SSH_HOST"]
		sshPort = env["REMOTE_SSH_PORT"]
		username = env["REMOTE_SSH_USER"]
		privateKey = env["REMOTE_SSH_KEY_PATH"]
		sshConfigHost = env["REMOTE_SSH_CONF"]
		copyCmd = env["COPY_COMMAND"]
		restartCmd = env["RESTART_COMMAND"]
		installCmd = env["INSTALL_COMMAND"]

	} else {
		return fmt.Errorf("environment file does not exist: %v", file)
	}
	return nil
}

func getHostKey(host string) (ssh.PublicKey, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get user home directory: %w", err)
	}
	knownHostsPath := filepath.Join(userHomeDir, ".ssh", "known_hosts")

	file, err := os.Open(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open known_hosts file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}

		if strings.Contains(fields[0], host) {
			keyBytes := []byte(fields[1] + " " + fields[2])
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(keyBytes)
			if err != nil {
				return nil, fmt.Errorf("error parsing host key for %q: %v", host, err)
			}
			fmt.Printf("Found key for host %s: %s\n", host, ssh.FingerprintSHA256(hostKey))
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New("no hostkey found for " + host)
	}

	return hostKey, nil
}

func executeSSHCommand(host string, username string, keyPath string, command string) error {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %v", err)
	}

	// Code for host key verification
	hostKey, err := getHostKey(host)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	// Connect to the SSH server.
	conn, err := ssh.Dial("tcp", host+":"+sshPort, config) // use sshPort here
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	// Create a session. It is one session per command.
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	// Print the output of the executed command
	fmt.Printf("Command output:\n%s\n", green(output))

	return nil
}

func executeLocalCommand(cmd string) (string, error) {
	fmt.Printf("Executing command: %s\n", blue(cmd))
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v", err)
	}
	return string(output), nil
}

func GetSSHConfig(host string) (*SSHConfig, error) {
	user := ssh_config.Get(host, "User")
	if user == "" {
		return nil, fmt.Errorf("user not found in SSH config for host: %s", host)
	}

	hostname := ssh_config.Get(host, "HostName")
	if hostname == "" {
		hostname = host
	}

	port := ssh_config.Get(host, "Port")
	if port == "" {
		return nil, fmt.Errorf("port not found in SSH config for host: %s", host)
	}

	identityFiles := ssh_config.GetAll(host, "IdentityFile")
	var identityFile string
	if len(identityFiles) == 0 {
		if runtime.GOOS == "darwin" {
			identityFile = fmt.Sprintf("/Users/%s/.ssh/id_rsa", user)
		} else {
			identityFile = fmt.Sprintf("/home/%s/.ssh/id_rsa", user)
		}
	} else {
		identityFile = identityFiles[0]
		expandedIdentityFile, err := expandHomeDir(identityFile)
		if err != nil {
			return nil, fmt.Errorf("failed to expand home directory in identity file path: %v", err)
		}
		identityFile = expandedIdentityFile
	}

	return &SSHConfig{
		Host:         host,
		HostName:     hostname,
		User:         user,
		Port:         port,
		IdentityFile: identityFile,
	}, nil
}

func scpFileToRemote(host string, username string, keyPath string, localFilePath string, remoteFilePath string) error {
	var cmd string
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to get file information: %v", err)
	}

	if fileInfo.IsDir() {
		cmd = fmt.Sprintf("scp -r -i %s %s %s@%s:%s", keyPath, localFilePath, username, host, remoteFilePath)
	} else {
		cmd = fmt.Sprintf("scp -i %s %s %s@%s:%s", keyPath, localFilePath, username, host, remoteFilePath)
	}

	executeLocalCommand(cmd)
	return nil
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
