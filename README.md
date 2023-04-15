# xdeb-install

Simple script to automatically download, convert, and install packages via the awesome [`xdeb`](https://github.com/toluschr/xdeb) tool. Basically just a wrapper to automate the process.

## Installation

Copy the script to `/usr/local/bin/`. The only dependencies to this script are `xdeb` and `curl`.

## Help Page

```
xdeb-install -h
[xdeb-install] USAGE: xdeb-install <package name> [FILE]
```

## Flags

The first flag `<package name>` has to be provided. For example:
```
xdeb-install vscode
xdeb-install google-chrome
```

The second flag `[FILE]` is optional and if provided must point to a local `.deb` file. This can be used to install a package without downloading it first. For example:
```
xdeb-install vscode ~/Downloads/code.deb
```

## Package Specifications

Package specifications are located in this repository under [`packages`](./packages). They are used to set package-specific definitions to use both for `xdeb` and for `xdeb-install`. All definitions from the [`packages/default`](./packages/default) specifications can and should be overwritten in other specifications.

If the flag `<package name>` corresponds to a package specification, then the definitions from `packages/default` + `packages/<arch>/<package name>` are used. Otherwise, only those from `packages/default` are used.
