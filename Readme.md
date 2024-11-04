# gh-bulk

A GH CLI extension for performing bulk operations on GitHub repositories.

## Requirements

- [GitHub CLI](https://cli.github.com/)
- Go 1.22.6 or later

## Installation

```sh
gh extension install jepomeroy/gh-bulk
```

## Removal

```sh
gh extension remove gh-bulk
```

## Usage

```sh
gh bulk
```

Follow the prompts to select the operation you would like to perform

## Development

See the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information.

### Running the extension locally

```sh
make
gh bulk
```

Make sure you uninstall the extension before running it locally

```sh
gh extension remove gh-bulk
```

or

```sh
make uninstall
```
