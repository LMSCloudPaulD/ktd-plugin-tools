package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

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

	green     = color.New(color.FgGreen).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	white     = color.New(color.FgWhite).SprintFunc()
	yellow    = color.New(color.FgYellow).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
)

func init() {
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
}

func main() {
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	if _, err := os.Stat("remote.env"); err == nil {
		file, err := os.Open("remote.env")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				switch key {
				case "REMOTE_SSH_HOST":
					if sshHost == "" {
						sshHost = value
					}
				case "REMOTE_SSH_PORT":
					if sshPort == "22" {
						sshPort = value
					}
				case "REMOTE_USER":
					if username == "" {
						username = value
					}
				case "REMOTE_ACTIVE":
					if !remoteActive {
						remoteActive = value == "true"
					}
				case "COPY_CMD":
					if copyCmd == "docker cp Koha koha-koha-1:/var/lib/koha/kohadev/plugins/" {
						copyCmd = value
					}
				case "RESTART_CMD":
					if restartCmd == "docker exec -ti koha-koha-1 bash -c 'koha-plack --restart kohadev'" {
						restartCmd = value
					}
				case "INSTALL_CMD":
					if installCmd == "docker exec -ti koha-koha-1 /kohadevbox/koha/misc/devel/install_plugins.pl" {
						installCmd = value
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	if sshConfigHost != "" {
		sshConfig, err := GetSSHConfig(sshConfigHost)
		if err != nil {
			log.Fatalf("Error getting SSH config: %v", err)
		}
		sshHost = sshConfig.HostName
		username = sshConfig.User
		sshPort = sshConfig.Port
		// Override private key path with default path for the user
		privateKey = fmt.Sprintf("/home/%s/.ssh/id_rsa", username)
	}

	if remoteActive {
		if sshHost != "" && username != "" && privateKey != "" {
			if copyFlagEnabled {
				err := executeSSHCommand(sshHost, username, privateKey, copyCmd)
				if err != nil {
					log.Fatalf("Error executing copy command: %v", err)
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

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
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

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	_, err = session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}

func executeLocalCommand(cmd string) {
	fmt.Printf("Executing command: %s\n", blue(cmd))
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to execute command: %s, error: %v", red(cmd), err)
	}
	fmt.Printf("Command output:\n%s\n", green(output))
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

	identityFile := ssh_config.Get(host, "IdentityFile")
	if identityFile == "" {
		identityFile = fmt.Sprintf("/home/%s/.ssh/id_rsa", user)
	}

	return &SSHConfig{
		Host:         host,
		HostName:     hostname,
		User:         user,
		Port:         port,
		IdentityFile: identityFile,
	}, nil

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
