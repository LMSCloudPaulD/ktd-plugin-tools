# 🐳 ktd-plugin-tools

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
| --help            | Show help                         | false                                                                      |

## Dependencies

- [github.com/fatih/color](https://github.com/fatih/color)
- [golang.org/x/crypto/ssh](https://golang.org/x/crypto/ssh)

