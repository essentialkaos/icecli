<p align="center"><a href="#readme"><img src="https://gh.kaos.st/icecli.svg"/></a></p>

<p align="center">
  <a href="https://github.com/essentialkaos/icecli/actions"><img src="https://github.com/essentialkaos/icecli/workflows/CI/badge.svg" alt="GitHub Actions Status" /></a>
  <a href="https://github.com/essentialkaos/icecli/actions?query=workflow%3ACodeQL"><img src="https://github.com/essentialkaos/icecli/workflows/CodeQL/badge.svg" /></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-icecli-master"><img alt="codebeat badge" src="https://codebeat.co/badges/455126f6-4d86-4c9f-af47-6a4c180bb5e7" /></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/icecli"><img src="https://goreportcard.com/badge/github.com/essentialkaos/icecli" alt="GoReportCard" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#command-line-completion">Command-line completion</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

</br>

`icecli` is a command-line tools for working with [Icecast Admin API](https://icecast.org/docs/icecast-2.4.1/admin-interface.html).

### Installation

#### From source

Before the initial install, allow git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (_reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)_):

```
git config --global http.https://pkg.re.followRedirects true
```

To build the `icecli` from scratch, make sure you have a working Go 1.13+ workspace (_[instructions](https://golang.org/doc/install)_), then:

```
go get github.com/essentialkaos/icecli
```

If you want to update `icecli` to latest stable release, do:

```
go get -u github.com/essentialkaos/icecli
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux from [EK Apps Repository](https://apps.kaos.st/icecli/latest):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) icecli
```

### Command-line completion

You can generate completion for `bash`, `zsh` or `fish` shell.

Bash:
```
[sudo] icecli --completion=bash 1> /etc/bash_completion.d/icecli
```


ZSH:
```
[sudo] icecli --completion=zsh 1> /usr/share/zsh/site-functions/icecli
```


Fish:
```
[sudo] icecli --completion=fish 1> /usr/share/fish/vendor_completions.d/icecli.fish
```

### Usage

```
Usage: icecli {options} {command} arguments…

Commands

  stats                               Show Icecast statistics
  list-mounts                         List mount points
  list-clients mount                  List clients
  move-clients from-mount to-mount    Move clients between mounts
  update-meta mount artist title      Update meta for mount
  kill-client mount client-id         Kill client connection
  kill-source mount                   Kill source connection
  help command                        Show detailed info about command usage

Options

  --host, -H host            URL of Icecast instance (default: http://127.0.0.1:8000)
  --user, -U username        Admin username (default: admin)
  --password, -P password    Admin password (default: hackme)
  --no-color, -nc            Disable colors in output
  --help, -h                 Show this help message
  --version, -v              Show version

Examples

  icecli stats -H 127.0.0.1:10000
  Show stats for server on 127.0.0.1:10000

  icecli kill-client -P mYsUpPaPaSs /stream3 361
  Detach client with ID 361 from /stream3

  icecli list-clients -H 127.0.0.1:10000 -U super_admin -P mYsUpPaPaSs /stream3
  List clients on /stream3

```

### Build Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://github.com/essentialkaos/icecli/workflows/CI/badge.svg?branch=master)](https://github.com/essentialkaos/icecli/actions) |
| `develop` | [![CI](https://github.com/essentialkaos/icecli/workflows/CI/badge.svg?branch=develop)](https://github.com/essentialkaos/icecli/actions) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
