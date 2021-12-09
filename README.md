# go-monkill

[![Go Reference](https://pkg.go.dev/badge/github.com/jtprogru/go-monkill.svg)](https://pkg.go.dev/github.com/jtprogru/go-monkill)
[![GolangCI-lint](https://github.com/jtprogru/go-monkill/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/jtprogru/go-monkill/actions/workflows/golangci-lint.yml)
[![build](https://github.com/jtprogru/go-monkill/actions/workflows/build.yml/badge.svg)](https://github.com/jtprogru/go-monkill/actions/workflows/build.yml)
[![publish](https://github.com/jtprogru/go-monkill/actions/workflows/publish.yml/badge.svg)](https://github.com/jtprogru/go-monkill/actions/workflows/publish.yml)
[![GitHub stars](https://img.shields.io/github/stars/jtprogru/go-monkill.svg)](https://github.com/jtprogru/go-monkill/stargazers)
[![GitHub issues](https://img.shields.io/github/issues-raw/jtprogru/go-monkill)](https://github.com/jtprogru/go-monkill/issues)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/jtprogru/go-monkill)](https://github.com/jtprogru/go-monkill/releases/latest)
[![Go report](https://goreportcard.com/badge/github.com/jtprogru/go-monkill)](https://goreportcard.com/report/github.com/jtprogru/go-monkill)
[![GitHub](https://img.shields.io/github/license/jtprogru/go-monkill)](LICENSE)
[![Linux](https://img.shields.io/badge/-Linux-grey?logo=linux)](https://en.wikipedia.org/wiki/Linux)
[![LoC](https://tokei.rs/b1/github/jtprogru/go-monkill)](https://github.com/jtprogru/go-monkill)
[![Donate](https://img.shields.io/badge/-Donate-yellow?logo=paypal)](https://paypal.me/jtprogru)

A very simple utility that allows you to run the desired command or script as soon as a certain process with a known PID completes correctly or with an error.


## Example

Example running:
```shell
go-monkill watch --pid=12345 --command="ping jtprog.ru -c 4"
```

When process with PID `12345` will finish or be killed, `go-monkill` will run command `ping jtprog.ru -c 4`

## Install

You can install this utility in the following ways:

### From pkg.go.dev
```shell
go install github.com/jtprogru/go-monkill
```

### Get binary
```shell
VERSION=v0.2.0 wget "https://github.com/jtprogru/go-monkill/releases/download/${VERSION}/go-monkill-darwin-amd64" -O /tmp/go-monkill
sudo mv /tmp/go-monkill /usr/local/bin/go-monkill
sudo chmod a+x /usr/local/bin/go-monkill
```

### Get from source code
```shell
git clone https://github.com/jtprogru/go-monkill
cd go-monkill
go install .
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
