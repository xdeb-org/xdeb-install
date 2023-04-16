# xdeb-install

Simple script to automatically download, convert, and install packages via the awesome [`xdeb`](https://github.com/toluschr/xdeb) tool. Basically just a wrapper to automate the process.

## Installation

1. Make sure the two dependencies are installed on your system: `curl` and `xdeb`.
2. Clone the repository to a desired location on your host
3. Adjust your `PATH` accordingly

Like so:
```
# step 1 - curl
sudo xbps-install -S curl

# step 1 - xdeb
mkdir -p ~/.local/bin
curl -fsSL -o ~/.local/bin/xdeb https://github.com/toluschr/xdeb/releases/download/1.3/xdeb
chmod +x ~/.local/bin/xdeb

# step 2
git clone https://github.com/thetredev/xdeb-install.git ~/.local/share/xdeb-install
ln -sf ~/.local/share/xdeb-install/xdeb-install ~/.local/bin/xdeb-install

# step 3 - bash
echo 'export PATH=${PATH}:${HOME}/.local/bin' >> ~/.bashrc

# step 3 - zsh
echo 'export PATH=${PATH}:${HOME}/.local/bin' >> ~/.zshrc
```

If you want to use links instead of adjusting your `PATH`, you can do this instead:
```
# step 3 - links
sudo ln -sf ~/.local/bin/xdeb /usr/local/bin/xdeb
sudo ln -sf ~/.local/bin/xdeb-install /usr/local/bin/xdeb-install
```

If necessary, checkout a specific tag of the repository:
```
cd ~/.local/share/xdeb-install
git fetch origin <tag>
git checkout <tag>
```

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

Package specifications are located in this repository under [`repositories`](./repositories). They are used to set package-specific definitions to use both for `xdeb` and for `xdeb-install`. All definitions from the [`repositories/default`](./repositories/default) specifications can and should be overwritten in other specifications.

If the flag `<package name>` corresponds to a package specification, then the definitions from `repositories/default` + `repositories/<arch>/<repository>/<package name>` are used. Otherwise, only those from `repositories/default` are used.

## Overwriting Package Specifications

All package definitions can be overwritten by setting them before executing the `xdeb-install` script. For example:
```
# defaults
XDEB_OPTS="-Sdei" xdeb-install vscode

# package-specific
XDEB_PACKAGE_URL="https://some-different-url.deb" xdeb-install vscode

# or both
XDEB_OPTS="-Sdei" XDEB_PACKAGE_URL="https://some-different-url.deb" xdeb-install vscode
```
Or, of course via `export`.
