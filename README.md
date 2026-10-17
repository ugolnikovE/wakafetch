# wakafetch

A simple utility to get your WakaTime or Wakapi statistics from the terminal.
I had two reasons to write it: to get my stats quickly without leaving the
terminal, and to try Go on a real project.

wakafetch supports two output formats:

```shell
$ wakafetch -t 5 week

  wakafetch | week
--------------------------------
  time      10h 44m
  lang      Go            ██████░░░░  61.23%
            Markdown      █░░░░░░░░░  14.20%
            Bash          █░░░░░░░░░  12.30%
            Unknown       ░░░░░░░░░░   6.45%
            Shell script  ░░░░░░░░░░   1.99%
  os        Macos
  editor    Zed
  project   wakafetch
  ai        41.13%
```

```shell
$ wakafetch -1 week
10h 44m | Go | ai 41.13%
```

## Usage

```
wakafetch [flags] [range]
```

Range is one of `today` (default), `yesterday`, `week`, `month`, `half`,
`year`, `all`.

| Flag | Description |
|------|-------------|
| `-1`, `--one-line` | print statistics on a single line |
| `-t`, `--lang-top int\|all` | number of top languages to show, or `all` (default 3) |
| `-u`, `--api-url string` | base URL of the WakaTime-compatible API |
| `-k`, `--api-key string` | API key. For debugging only: the key stays in your shell history |
| `-h`, `--help` | show help |

## Configuration

wakafetch reads the same settings as the WakaTime plugins. Each source
overrides the previous one:

1. `api_key` and `api_url` in the `[settings]` section of
   `$WAKATIME_HOME/.wakatime.cfg` or `~/.wakatime.cfg`
2. `WAKATIME_API_KEY` and `WAKATIME_API_URL` environment variables
3. command-line flags

If `api_url` is not set, `https://api.wakatime.com/api/v1` is used.

## Build

```bash
git clone https://github.com/ugolnikovE/wakafetch
cd wakafetch
go build .
./wakafetch
```

## Install

```bash
go install github.com/ugolnikovE/wakafetch@latest
```

## Contributing

Ideas, bug reports and any contributions are welcome — feel free to open an issue
or a pull request.

## License

MIT, see [LICENSE](LICENSE).
