# xdeb-install

Simple tool to automatically download, convert, and install DEB packages via the awesome [`xdeb`](https://github.com/toluschr/xdeb) tool. Basically just a wrapper to automate the process.

## Installation

You can either use [Go](https://go.dev/) or download a release to install `xdeb-install`. You can install `xdeb` using `xdeb-install` later, see [Help Page](#help-page).

### Using Go

If you have Go installed, simply execute:
```
go install github.com/thetredev/xdeb-install
```

As long as the `GOPATH` is within your `PATH`, that's it.

### Manually

Head over to the [releases](https://github.com/thetredev/xdeb-install/releases) page and download a release binary. Then move it to some place within your `PATH`, like `/usr/local/bin`.

## Help Page

```
$ xdeb-install -h
NAME:
   xdeb-install - Automation wrapper for the xdeb utility

USAGE:
   xdeb-install [global options] command [command options] [arguments...]

DESCRIPTION:
   Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb tool.
   Basically just a wrapper to automate the process.

COMMANDS:
   xdeb           installs the xdeb utility to the system along with its dependencies
   repository, r  install a package from an online APT repository
   search, s      search online APT repositories for a package
   sync           sync online APT repositories
   url, u         install a package from a URL directly
   file, f        install a package from a local DEB file
   help, h        Shows a list of commands or help for one command
   clean, c       cleanup temporary xdeb context root path, optionally the repository lists as well

GLOBAL OPTIONS:
   --options value, -o value  override XDEB_OPTS, '-i' will be removed if provided (default: "-Sde")
   --temp value, -t value     temporary xdeb context root path (default: "/tmp/xdeb")
   --help, -h                 show help
```

## Sync package repositories

To sync package repositories, type:
```
$ xdeb-install sync
Syncing lists: https://raw.githubusercontent.com/thetredev/xdeb-install-repositories/main/repositories/x86_64/lists.yaml
[xdeb-install] Syncing repository debian.org/contrib @ bookworm
[xdeb-install] Syncing repository debian.org/main @ bookworm
[xdeb-install] Syncing repository debian.org/non-free @ bookworm
[xdeb-install] Syncing repository debian.org/non-free-firmware @ bookworm
...
[xdeb-install] Syncing repository microsoft.com/vscode.yaml @ current
[xdeb-install] Syncing repository google.com/google-chrome.yaml @ current
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
