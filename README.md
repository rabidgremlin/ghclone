# ghclone

Clone a GitHub repository with `gh` and set local git user config from your GitHub account.

## Install

### macOS and Linux

Replace `vX.Y.Z` with the desired release tag:

```bash
curl -fsSL -o ghclone "https://github.com/rabidgremlin/ghclone/releases/download/vX.Y.Z/ghclone-vX.Y.Z-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64" && chmod +x ghclone
```

### Windows

Download `ghclone-vX.Y.Z-windows-amd64.exe` from the [releases page](https://github.com/rabidgremlin/ghclone/releases).

## Usage

```bash
ghclone <owner/repo>
ghclone --help
ghclone --version
```

`--version` prints the release tag for release builds, and `dev-build` for local builds.
