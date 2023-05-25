# 🐳 ktd-plugin-tools

ktd-plugin-tools is a helper for Koha plugin development with koha-testing-docker.

## Usage

```sh
ktd-plugin-tools [options]
```

## Options

| Option            | Description                       | Default                                                                    | Shorthand |
| ----------------- | --------------------------------- | -------------------------------------------------------------------------- | --------- |
| --env             | Path to environment file          | kpt.env                                                                    |           |
| --copy            | Enable copy command               | true                                                                       | -c        |
| --restart         | Enable restart command            | true                                                                       | -r        |
| --install         | Enable install command            | true                                                                       | -i        |
| --copy-cmd        | Copy command                      | docker cp Koha koha-koha-1:/var/lib/koha/kohadev/plugins/                  |           |
| --restart-cmd     | Restart command                   | docker exec -ti koha-koha-1 bash -c 'koha-plack --restart kohadev'         |           |
| --install-cmd     | Install command                   | docker exec -ti koha-koha-1 /kohadevbox/koha/misc/devel/install_plugins.pl |           |
| --help            | Show help                         | false                                                                      | -h        |
| --quiet           | Disable output                    | false                                                                      | -q        |

Note: Command line flags take precedence over environment file variables.

## Dependencies

- [github.com/fatih/color](https://github.com/fatih/color)
- [golang.org/x/crypto/ssh](https://golang.org/x/crypto/ssh)
- [golang.org/x/sync](golang.org/x/sync)
