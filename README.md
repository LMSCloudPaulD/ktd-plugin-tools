# ktd-plugin-tools

ktd-plugin-tools is a helper for Koha plugin development with koha-testing-docker.

## Usage

```sh
ktd-plugin-tools [options]
```

## Options

| Option            | Description                       | Default                                                                    |
| ----------------- | --------------------------------- | -------------------------------------------------------------------------- |
| --copy            | Enable copy command               | true                                                                       |
| --restart         | Enable restart command            | true                                                                       |
| --install         | Enable install command            | true                                                                       |
| --copy-cmd        | Copy command                      | docker cp Koha koha-koha-1:/var/lib/koha/kohadev/plugins/                  |
| --restart-cmd     | Restart command                   | docker exec -ti koha-koha-1 bash -c 'koha-plack --restart kohadev'         |
| --install-cmd     | Install command                   | docker exec -ti koha-koha-1 /kohadevbox/koha/misc/devel/install_plugins.pl |
| --ssh-host        | The SSH host                      |                                                                            |
| --username        | The SSH username                  |                                                                            |
| --private-key     | The path to SSH private key       |                                                                            |
| --ssh-port        | The SSH port                      | 22                                                                         |
| --ssh-config-host | The SSH host from SSH config file |                                                                            |
| --remote          | Execute command remotely          | false                                                                      |
| --help            | Show help                         | false                                                                      |

## SSHConfig Struct

The SSHConfig struct represents the SSH configuration.

```go
type SSHConfig struct {
	Host         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
}
```

## Function Reference

### func init()

The `init` function initializes the flags.

### func main()

The `main` function is the entry point of the program.

### func executeSSHCommand(host string, username string, keyPath string, command string) error

The `executeSSHCommand` function executes an SSH command on a remote host.

### func executeLocalCommand(cmd string)

The `executeLocalCommand` function executes a local command.

### func GetSSHConfig(host string) (*SSHConfig, error)

The `GetSSHConfig` function retrieves the SSH configuration for a host from the SSH config file.

### func printHelp()

The `printHelp` function prints the help information.

## Dependencies

- github.com/fatih/color
- github.com/kevinburke/ssh_config
- golang.org/x/crypto/ssh

