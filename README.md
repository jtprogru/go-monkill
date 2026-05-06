# go-monkill

[![Go Reference](https://pkg.go.dev/badge/github.com/jtprogru/go-monkill.svg)](https://pkg.go.dev/github.com/jtprogru/go-monkill)
[![Go Report Card](https://goreportcard.com/badge/github.com/jtprogru/go-monkill)](https://goreportcard.com/report/github.com/jtprogru/go-monkill)
[![GolangCI-lint](https://github.com/jtprogru/go-monkill/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/jtprogru/go-monkill/actions/workflows/golangci-lint.yml)
[![build](https://github.com/jtprogru/go-monkill/actions/workflows/build.yml/badge.svg)](https://github.com/jtprogru/go-monkill/actions/workflows/build.yml)
[![release](https://github.com/jtprogru/go-monkill/actions/workflows/release.yml/badge.svg)](https://github.com/jtprogru/go-monkill/actions/workflows/release.yml)
[![GitHub stars](https://img.shields.io/github/stars/jtprogru/go-monkill.svg?color=gold)](https://github.com/jtprogru/go-monkill/stargazers)
[![GitHub issues](https://img.shields.io/github/issues-raw/jtprogru/go-monkill?color=blue)](https://github.com/jtprogru/go-monkill/issues)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/jtprogru/go-monkill)](https://github.com/jtprogru/go-monkill/releases/latest)
[![Go report](https://goreportcard.com/badge/github.com/jtprogru/go-monkill)](https://goreportcard.com/report/github.com/jtprogru/go-monkill)
[![GitHub](https://img.shields.io/github/license/jtprogru/go-monkill?color=gold)](LICENSE)
[![Linux](https://img.shields.io/badge/-Linux-grey?logo=linux)](https://en.wikipedia.org/wiki/Linux)
[![LoC](https://tokei.rs/b1/github/jtprogru/go-monkill)](https://github.com/jtprogru/go-monkill)
[![Donate](https://img.shields.io/badge/PayPal-Donate-green?logo=paypal)](https://paypal.me/jtprogru)

Very simple utility that allows you to run the desired command or script as soon as a certain process with a known PID completes correctly or with an error.

## Features

- Watch a PID and run an arbitrary command once the process exits.
- Shell-style command parsing (single/double quotes, escapes) via [shlex](https://github.com/google/shlex).
- Verbose debug logging of every poll cycle, watch duration and the command's exit code.
- Optional JSON-formatted external log file (`--logfile`), parallel with text output to stderr.
- Graceful shutdown on `SIGINT`/`SIGTERM` — polling stops cleanly, the dependent command is skipped, exit code `130`.
- Propagates the user-command's exit code as the utility's own exit code.

## Example

```shell
go-monkill watch --pid=12345 --command="ping jtprog.ru -c 4"
```

When process with PID `12345` finishes or is killed, `go-monkill` runs `ping jtprog.ru -c 4` and exits with the command's exit code.

Quoted arguments work as in a shell:

```shell
go-monkill watch --pid=12345 --command="sh -c 'echo done >> /var/log/notify.log'"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--pid` | _required_ | PID to watch |
| `--command` | _required_ | command to run after the process exits (shell-style quoting supported) |
| `--timeout` | `250` | poll interval in milliseconds |
| `-v`, `--verbose` | `false` | verbose (debug-level) output |
| `--logfile` | _empty_ | path to a log file (JSON format); empty disables file logging |

### Exit codes

| Code | Meaning |
|------|---------|
| `0` | watched process exited and the user-command succeeded |
| `1` | invalid input (PID, empty `--command`, parse error, waiter failure) |
| _N_ | exit code of the user-command if it returned non-zero |
| `130` | interrupted by `SIGINT`/`SIGTERM` before the watched process exited |

## Install

### Homebrew (macOS / Linux)

```shell
brew tap jtprogru/tap
brew install go-monkill
```

### Pre-built binaries

Download from the [releases page](https://github.com/jtprogru/go-monkill/releases/latest), e.g.:

```shell
VERSION=v0.3.0
OS=Darwin   # or Linux
ARCH=arm64  # or x86_64
curl -L "https://github.com/jtprogru/go-monkill/releases/download/${VERSION}/go-monkill_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C /tmp go-monkill
sudo install -m 0755 /tmp/go-monkill /usr/local/bin/go-monkill
```

Verify the checksum signature:

```shell
gpg --verify checksums.txt.sig checksums.txt
sha256sum -c checksums.txt
```

### From source

```shell
go install github.com/jtprogru/go-monkill@latest
```

## Development

The project uses [Task](https://taskfile.dev) instead of `make`. Run `task --list` to see all targets:

```shell
task install-deps    # go mod tidy
task fmt             # gofmt -s -w .
task vet             # go vet ./...
task lint            # golangci-lint run
task test            # tests + coverage summary
task build           # local binary into ./dist
task release:snapshot  # goreleaser snapshot (no publish, no signing)
task ci              # vet + lint + test (used by CI)
```

## Feedback

If it happened that you started using this utility, and you have feedback, please create [issues](https://github.com/jtprogru/go-monkill/issues) or contact the Telegram chat [jtprogru_chat](https://t.me/jtprogru_chat).

## Authors

- Michael Savin
    - :octocat: [@jtprogru](https://www.github.com/jtprogru)
    - :bird: [@jtprogru](https://www.twitter.com/jtprogru)
    - :moneybag: [savinmi.ru](https://savinmi.ru)
- Ivan Anfilatov:
    - :octocat: [@t0pep0](https://github.com/t0pep0)

## License

See [LICENSE](LICENSE)
