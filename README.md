# xdeb-install

Simple tool to automatically download, convert, and install DEB packages on [Void Linux](https://voidlinux.org) via the awesome [`xdeb`](https://github.com/toluschr/xdeb) tool. Basically just a wrapper to automate the process.

## Installation

There are three ways you can install the tool:
  - [using XBPS](#xbps)
  - [using Go](#using-go)
  - [manually downloading a release binary](#manually)

You can install `xdeb` using `xdeb-install` later, see [Help Page](#help-page).

## XBPS

*Before you continue reading this section*, read up on https://docs.voidlinux.org/xbps/repositories/custom.html. You have been warned.

Since [my PR over at void-linux/void-packages](https://github.com/void-linux/void-packages/pull/46352) didn't make it, you can't install the tool using any official XBPS repositories.

To work around that problem, I created my own unofficial XBPS repository at https://thetredev.github.io/voidlinux-repository. See https://github.com/thetredev/voidlinux-repository for instructions on how to install it to your system.

Afterwards, you can execute `xbps-install xdeb-install` to install the tool.

**Notes**:
  - All binary architectures for release 1.0.1 are available: `x86_64`, `i686`, `aarch64`

### Using Go

If you have [Go](https://go.dev) installed, simply execute:
```
go install github.com/thetredev/xdeb-install
```

As long as the `GOPATH` is within your `PATH`, that's it.

### Manually

Head over to the [releases](https://github.com/thetredev/xdeb-install/releases) page and download a release binary. Then move it to some place within your `PATH`, like `/usr/local/bin`. Make sure to make it executable afterwards: `sudo chmod +x /usr/local/bin/xdeb-install`.

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
   Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb tool.
   Basically just a wrapper to automate the process.

COMMANDS:
   xdeb           installs the xdeb utility to the system along with its dependencies
   providers      lists available providers
   sync           sync remote repositories
   search, s      search remote repositories for a package
   repository, r  install a package from an remote repository
   url, u         install a package from a URL directly
   file, f        install a package from a local DEB file
   clean, c       cleanup temporary xdeb context root path, optionally the repository lists as well
   version, v     prints the version of this tool
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --options value, -o value  override XDEB_OPTS, '-i' will be removed if provided (default: "-Sde")
   --temp value, -t value     temporary xdeb context root path (default: "/tmp/xdeb")
   --help, -h                 show help
```

## List available providers

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
  architecture:
  url: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/main/repositories/x86_64/microsoft.com
    distribution: current
      component: vscode.yaml

google.com
  architecture:
  url: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/main/repositories/x86_64/google.com
    distribution: current
      component: google-chrome.yaml
```

## Sync package repositories

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

## Supported repositories

See https://github.com/thetredev/xdeb-install-repositories for details.

## Install DEB packages from remote repositories

To install `speedcrunch`, for example, type:
```
$ xdeb-install repository speedcrunch
```

This will install the most recent version of the package from the first provider and distribution it can find.

You can also specify the provider and distribution, for example `debian.org` and `bookworm`, respectively:
```
$ xdeb-install repository --provider debian.org --distribution bookworm speedcrunch
```

## Install DEB packages directly from a URL

Let's stay with the `speedcrunch` example:
```
$ xdeb-install url http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
```

## Install DEB packages from a local DEB file

First, obviously download a DEB file from a remote location. Let's stay it's stored at `$HOME/Downloads/speedcrunch.deb`:
```
$ xdeb-install file $HOME/Downloads/speedcrunch.deb
```

## Search for a DEB package

You can search for a specific package by its name, let's stay with `speedcrunch`:

```
$ xdeb-install search speedcrunch
```

Output:
```
[xdeb-install] Looking for package speedcrunch via provider * and distribution * ...
debian.org/main
  distribution: bookworm
  version: 0.12.0-6
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
  sha256: a306a478bdf923ad1206a1a76fdc1b2d6a745939663419b360febfa6350e96b6

debian.org/main
  distribution: sid
  version: 0.12.0-6
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
  sha256: a306a478bdf923ad1206a1a76fdc1b2d6a745939663419b360febfa6350e96b6

debian.org/main
  distribution: testing
  version: 0.12.0-6
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-6_amd64.deb
  sha256: a306a478bdf923ad1206a1a76fdc1b2d6a745939663419b360febfa6350e96b6

debian.org/main
  distribution: bullseye
  version: 0.12.0-5
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-5_amd64.deb
  sha256: 0c108597debfbc47e6eb384cfff5539627d0f0652202a63f82aa3c3e8f56aa5c

debian.org/main
  distribution: buster
  version: 0.12.0-4
  url: http://ftp.debian.org/debian/pool/main/s/speedcrunch/speedcrunch_0.12.0-4_amd64.deb
  sha256: 8681da5ca651a6a7f5abb479c673d33ce3525212e34a2a33afcec7ad75c28aea
```

Filtering search results is also supported via `--provider <provider> [--distribution <distribution>]`:

```
$ xdeb-install search --provider ubuntu.com --distribution bionic speedcrunch
```

Output:
```
[xdeb-install] Looking for package speedcrunch via provider ubuntu.com and distribution bionic ...
ubuntu.com/universe
  distribution: bionic
  version: 0.12.0-3
  url: http://archive.ubuntu.com/ubuntu/pool/universe/s/speedcrunch/speedcrunch_0.12.0-3_amd64.deb
  sha256: 0206f112ac503393c984088817488aa21589c1c5f16f67df8d8836612f27f810
```
