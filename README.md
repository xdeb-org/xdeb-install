[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![devcontainer Status](https://img.shields.io/github/actions/workflow/status/thetredev/xdeb-install/docker.yml?label=devcontainer
)](https://github.com/thetredev/xdeb-install/actions/workflows/docker.yml)
[![Release Status](https://img.shields.io/github/actions/workflow/status/thetredev/xdeb-install/release.yml?label=release
)](https://github.com/thetredev/xdeb-install/actions/workflows/release.yml)

# xdeb-install

Simple tool to automatically download, convert, and install DEB packages on [Void Linux](https://voidlinux.org) via the awesome [`xdeb`](https://github.com/toluschr/xdeb) tool. Basically just a wrapper to automate the process.

## Table of Contents

- [Help Page](#help-page)
- [Installation](#installation)
  - [Using XBPS](#using-xbps)
  - [Using Go](#using-go)
  - [Manually](#manually)
- [Listing available providers](#listing-available-providers)
- [Managing package repositories](#managing-package-repositories)
  - [Syncing package repositories](#syncing-package-repositories)
  - [Supported package repositories](#supported-package-repositories)
- [Searching for DEB packages](#searching-for-deb-packages)
  - [General instructions](#general-instructions)
  - [Search filtering by provider/distribution](#search-filtering-by-providerdistribution)
  - [Inexact matches](#inexact-matches)
- [Installing DEB packages](#installing-deb-packages)
  - [From remote repositories](#from-remote-repositories)
  - [Directly from a URL](#directly-from-a-url)
  - [Directly from a local file](#directly-from-a-local-file)

## Help Page

To display the help page, type:
```
$ xdeb-install -h
```

Output:
```
NAME:
   xdeb-install - Automation wrapper for the xdeb utility

USAGE:
   xdeb-install [global options] command [command options] [arguments...]

DESCRIPTION:
   Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb utility.
   Basically just a wrapper to automate the process.

COMMANDS:
   xdeb          install the xdeb utility to the system along with its dependencies
   providers, p  list available providers
   sync, S       synchronize remote repositories
   search, s     search remote repositories for a package
   deb, apt, i   install a package from a remote repository
   url, u        install a package from a URL directly
   file, f       install a package from a local DEB file
   clean, c      cleanup temporary xdeb context root path, optionally the repository lists as well
   version, v    print the version of this tool
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --options value, -o value  override XDEB_OPTS, '-i' will be removed if provided (default: "-Sde")
   --temp value, -t value     set the temporary xdeb context root path (default: "/tmp/xdeb")
   --help, -h                 show help
```

#### xdeb

```
$ xdeb-install xdeb
NAME:
   xdeb-install xdeb [release version] - install the xdeb utility to the system along with its dependencies

USAGE:
   xdeb-install xdeb [release version] [command options] [arguments...]

OPTIONS:
   --help, -h  show help
```

To install the current `master` version of the `xdeb` utility, type:
```
$ xdeb-install xdeb
```

To install a specific version of the `xdeb` utility, type:
```
$ xdeb-install xdeb <version>
```

For example:
```
$ xdeb-install xdeb 1.3
```

The tool also supports to figure out the latest release of `xdeb` when `latest` is provided as `<version>`. Example:
```
$ xdeb-install xdeb latest
```

As of today, 2023-10-04, this will install release `1.3`.

#### providers

```
$ xdeb-install providers -h
NAME:
   xdeb-install providers - list available providers

USAGE:
   xdeb-install providers [command options] [arguments...]

OPTIONS:
   --help, -h  show help
```

See [Listing available providers](#listing-available-providers)

#### sync

```
$ xdeb-install sync -h
NAME:
   xdeb-install sync [provider list] - synchronize remote repositories

USAGE:
   xdeb-install sync [provider list] [command options] [arguments...]

OPTIONS:
   --help, -h  show help
```

See [Syncing package repositories](#syncing-package-repositories)

#### search

```
$ xdeb-install search -h
NAME:
   xdeb-install search - search remote repositories for a package

USAGE:
   xdeb-install search [command options] [arguments...]

OPTIONS:
   --exact, -e                                   perform an exact match of the package name provided (default: false)
   --provider value, -p value                    limit search results to a specific provider
   --distribution value, --dist value, -d value  limit search results to a specific distribution (requires --provider)
   --help, -h                                    show help
```

See [Searching for DEB packages](#searching-for-deb-packages)

#### deb

```
$ xdeb-install deb -h
NAME:
   xdeb-install deb - install a package from a remote repository

USAGE:
   xdeb-install deb [command options] [arguments...]

OPTIONS:
   --provider value, -p value                    limit search results to a specific provider
   --distribution value, --dist value, -d value  limit search results to a specific distribution (requires --provider)
   --help, -h                                    show help
```

See [Installing DEB packages/From remote repositories](#from-remote-repositories)

#### url

```
$ xdeb-install url -h
NAME:
   xdeb-install url [URL] - install a package from a URL directly

USAGE:
   xdeb-install url [URL] [command options] [arguments...]

OPTIONS:
   --help, -h  show help
```

See [Installing DEB packages/Directly from a URL](#directly-from-a-url)

#### file

```
$ xdeb-install file -h
NAME:
   xdeb-install file [path] - install a package from a local DEB file

USAGE:
   xdeb-install file [path] [command options] [arguments...]

OPTIONS:
   --help, -h  show help
```

See [Installing DEB packages/Directly from a local file](#directly-from-a-local-file)

#### clean

```
$ xdeb-install clean -h
NAME:
   xdeb-install clean - cleanup temporary xdeb context root path, optionally the repository lists as well

USAGE:
   xdeb-install clean [command options] [arguments...]

OPTIONS:
   --lists, -l  cleanup repository lists as well (default: false)
   --help, -h   show help
```

#### version

```
$ xdeb-install version -h
NAME:
   xdeb-install version - print the version of this tool

USAGE:
   xdeb-install version [command options] [arguments...]

OPTIONS:
   --help, -h  show help
```

## Installation

There are three ways you can install the tool:
  - [using XBPS](#using-xbps)
  - [using Go](#using-go)
  - [manually downloading a release binary](#manually)

You can install `xdeb` using `xdeb-install` later, see [Help Page](#help-page).

### Using XBPS

*Before you continue reading this section*, read up on https://docs.voidlinux.org/xbps/repositories/custom.html. You have been warned.

Since [my PR over at void-linux/void-packages](https://github.com/void-linux/void-packages/pull/46352) didn't make it, you can't install the tool using any official XBPS repositories.

To work around that problem, I created my own unofficial XBPS repository at https://thetredev.github.io/voidlinux-repository. See https://github.com/thetredev/voidlinux-repository for instructions on how to install it to your system.

Afterwards, you can execute `xbps-install xdeb-install` to install the tool.

### Using Go

If you have [Go](https://go.dev) installed, simply execute:
```
go install github.com/thetredev/xdeb-install/v2@latest
```

As long as the `GOPATH` is within your `PATH`, that's it.

### Manually

Head over to the [releases](https://github.com/thetredev/xdeb-install/releases) page and download a release binary. Then move it to some place within your `PATH`, like `/usr/local/bin`. Make sure to make it executable afterwards: `sudo chmod +x /usr/local/bin/xdeb-install`.

## Listing available providers

To check which providers are available, type:
```
$ xdeb-install providers
```

Output:
```
[xdeb-install] Syncing lists: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/v1.0.0/repositories/x86_64/lists.yaml
debian.org
  architecture: amd64
  url: http://ftp.debian.org/debian
    distribution: bookworm
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: bookworm-backports
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: bullseye
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: bullseye-backports
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: buster
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: buster-backports
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: sid
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: testing
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware
    distribution: testing-backports
      component: main
      component: contrib
      component: non-free
      component: non-free-firmware

linuxmint.com
  architecture: amd64
  url: http://packages.linuxmint.com
    distribution: victoria
      component: main
      component: backport
      component: import
      component: upstream
    distribution: vera
      component: main
      component: backport
      component: import
      component: upstream
    distribution: vanessa
      component: main
      component: backport
      component: import
      component: upstream
    distribution: faye
      component: main
      component: backport
      component: import
      component: upstream

ubuntu.com
  architecture: amd64
  url: http://archive.ubuntu.com/ubuntu
    distribution: bionic
      component: main
      component: multiverse
      component: restricted
      component: universe
    distribution: focal
      component: main
      component: multiverse
      component: restricted
      component: universe
    distribution: jammy
      component: main
      component: multiverse
      component: restricted
      component: universe

microsoft.com
  architecture: amd64
  url: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/main/repositories/x86_64/microsoft.com
    distribution: current
      component: vscode.yaml

google.com
  architecture: amd64
  url: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/main/repositories/x86_64/google.com
    distribution: current
      component: google-chrome.yaml
```

## Managing package repositories

### Syncing package repositories

To sync package repositories, type:
```
$ xdeb-install sync
```

Output:
```
[xdeb-install] Syncing lists: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/v1.0.0/repositories/x86_64/lists.yaml
[xdeb-install] Syncing repository debian.org/bullseye: contrib
[xdeb-install] Syncing repository debian.org/bookworm: contrib
[xdeb-install] Syncing repository debian.org/bullseye: non-free
[xdeb-install] Syncing repository debian.org/bookworm: non-free
...
[xdeb-install] Syncing repository linuxmint.com/victoria: import
[xdeb-install] Syncing repository linuxmint.com/vanessa: import
[xdeb-install] Syncing repository linuxmint.com/vera: import
[xdeb-install] Syncing repository linuxmint.com/vanessa: mai
...
[xdeb-install] Syncing repository ubuntu.com/bionic: restricted
[xdeb-install] Syncing repository ubuntu.com/focal: restricted
[xdeb-install] Syncing repository ubuntu.com/jammy: restricted
[xdeb-install] Syncing repository ubuntu.com/focal: multiverse
...
[xdeb-install] Syncing repository microsoft.com/current: vscode.yaml
[xdeb-install] Syncing repository google.com/current: google-chrome.yaml
[xdeb-install] Finished syncing: ~/.config/xdeb-install/repositories/x86_64
```

The log output is not in order because syncing is parallelized.

You can also filter the providers to sync, like so:
```
$ xdeb-install sync ubuntu.com
```

Output:
```
[xdeb-install] Syncing lists: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/v1.0.0/repositories/x86_64/lists.yaml
[xdeb-install] Syncing repository ubuntu.com/bionic: restricted
[xdeb-install] Syncing repository ubuntu.com/focal: restricted
[xdeb-install] Syncing repository ubuntu.com/jammy: restricted
[xdeb-install] Syncing repository ubuntu.com/bionic: multiverse
[xdeb-install] Syncing repository ubuntu.com/focal: multiverse
[xdeb-install] Syncing repository ubuntu.com/jammy: multiverse
[xdeb-install] Syncing repository ubuntu.com/bionic: main
[xdeb-install] Syncing repository ubuntu.com/focal: main
[xdeb-install] Syncing repository ubuntu.com/jammy: main
[xdeb-install] Syncing repository ubuntu.com/focal: universe
[xdeb-install] Syncing repository ubuntu.com/bionic: universe
[xdeb-install] Syncing repository ubuntu.com/jammy: universe
[xdeb-install] Finished syncing: ~/.config/xdeb-install/repositories/x86_64
```

The package repository lists are stored at `$XDG_CONFIG_HOME/xdeb-install/repositories/<arch>`, where `$XDG_CONFIG_HOME` typically translates to `$HOME/.config`.

### Supported package repositories

See https://github.com/thetredev/xdeb-install-repositories for details.

## Searching for DEB packages

### General instructions
You can search for a specific package by its name, let's stay with `speedcrunch`:
```
$ xdeb-install search speedcrunch
```

Output:
```
[xdeb-install] Looking for package speedcrunch (exact: false) via provider * and distribution * ...
debian.org/main
  package: speedcrunch
  distribution: bookworm
  version: 0.12.0-6
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
  sha256: a306a478bdf923ad1206a1a76fdc1b2d6a745939663419b360febfa6350e96b6

debian.org/main
  package: speedcrunch
  distribution: sid
  version: 0.12.0-6
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
  sha256: a306a478bdf923ad1206a1a76fdc1b2d6a745939663419b360febfa6350e96b6

debian.org/main
  package: speedcrunch
  distribution: testing
  version: 0.12.0-6
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
  sha256: a306a478bdf923ad1206a1a76fdc1b2d6a745939663419b360febfa6350e96b6

debian.org/main
  package: speedcrunch
  distribution: bullseye
  version: 0.12.0-5
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-5_amd64.deb
  sha256: 0c108597debfbc47e6eb384cfff5539627d0f0652202a63f82aa3c3e8f56aa5c

ubuntu.com/universe
  package: speedcrunch
  distribution: jammy
  version: 0.12.0-5
  url: http://archive.ubuntu.com/ubuntu/pool/universe/s/speedcrunch/speedcrunch_0.12.0-5_amd64.deb
  sha256: 241d302af8d696032d11abbc6e46d045934c23461786c4876fcc82e1743eec33

ubuntu.com/universe
  package: speedcrunch
  distribution: focal
  version: 0.12.0-4build1
  url: http://archive.ubuntu.com/ubuntu/pool/universe/s/speedcrunch/speedcrunch_0.12.0-4build1_amd64.deb
  sha256: 79c0075eea11b172d17963da185a0dffb9d2ab368fd5c64c812c695127579922

debian.org/main
  package: speedcrunch
  distribution: buster
  version: 0.12.0-4
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-4_amd64.deb
  sha256: 8681da5ca651a6a7f5abb479c673d33ce3525212e34a2a33afcec7ad75c28aea

ubuntu.com/universe
  package: speedcrunch
  distribution: bionic
  version: 0.12.0-3
  url: http://archive.ubuntu.com/ubuntu/pool/universe/s/speedcrunch/speedcrunch_0.12.0-3_amd64.deb
  sha256: 0206f112ac503393c984088817488aa21589c1c5f16f67df8d8836612f27f81
```

### Search filtering by provider/distribution
Filtering search results is also supported via `--provider <provider> [--distribution <distribution>]`:
```
$ xdeb-install search --provider ubuntu.com --distribution bionic speedcrunch
```

Output:
```
[xdeb-install] Looking for package speedcrunch  (exact: false) via provider ubuntu.com and distribution bionic ...
ubuntu.com/universe
  package: speedcrunch
  distribution: bionic
  version: 0.12.0-3
  url: http://archive.ubuntu.com/ubuntu/pool/universe/s/speedcrunch/speedcrunch_0.12.0-3_amd64.deb
  sha256: 0206f112ac503393c984088817488aa21589c1c5f16f67df8d8836612f27f810
```

### Inexact matches
Futhermore, the flag `--exact` (or `-e`) specifies whether the search will look for a package of the exact name as provided:
```
$ xdeb-install search --exact google-chrome
```

Output:
```
[xdeb-install] Looking for package google-chrome (exact: true) via provider * and distribution * ...
google.com/google-chrome
  package: google-chrome
  distribution: current
  url: https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
```

Omitting `--exact` yields:
```
$ xdeb-install search google-chrome
[xdeb-install] Looking for package google-chrome (exact: false) via provider * and distribution * ...
google.com/google-chrome
  package: google-chrome
  distribution: current
  url: https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb

google.com/google-chrome
  package: google-chrome-unstable
  distribution: current
  url: https://dl.google.com/linux/direct/google-chrome-unstable_current_amd64.deb
```

Currently, the only pattern available is `startsWith`, effectively matching `google-chrome*` in the example above.

## Installing DEB packages

### From remote repositories

To install `speedcrunch`, for example, type:
```
$ xdeb-install deb speedcrunch
```

This will install the most recent version of the package from the first provider and distribution it can find.

You can also specify the provider and distribution, for example `debian.org` and `bookworm`, respectively:
```
$ xdeb-install deb --provider debian.org --distribution bookworm speedcrunch
```

### Directly from a URL

Let's stay with the `speedcrunch` example:
```
$ xdeb-install url http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
```

### Directly from a local file

First, obviously download a DEB file from a remote location. Let's stay it's stored at `$HOME/Downloads/speedcrunch.deb`:
```
$ xdeb-install file $HOME/Downloads/speedcrunch.deb
```
