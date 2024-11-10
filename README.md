# gh-bulk

A GH CLI extension for performing bulk operations on GitHub repositories.

## Requirements

- [GitHub CLI](https://cli.github.com/)
- Go 1.21.13 or later

## Installation

```sh
gh extension install jepomeroy/gh-bulk
```

## Removal

```sh
gh extension remove bulk
```

## Usage

```sh
gh bulk
```

### First time setup

When running the extension for the first time, you are prompted to enter prompted for user information.

- If you are and individual user, select `Individual` and your `Current GitHub User` is stored and used to retrieve repositories.
  ![Individual setup](./images/individual.png)

- If you are member of an organization or an external collaborator, select `Organization` and enter the organization name at next prompt. The organization name is stored and used to retrieve repositories.
  ![Organization setup](./images/organization.png)
  ![Organization name](./images/org-name.png)

The username/organization information is stored in the `~/.config/gh/gh-bulk/config.json` file on Linux and MacOS and `%USERPROFILE%\.config\gh\gh-bulk\config.json` on Windows. If you change GitHub accounts by running `gh auth login`, you are prompted to enter the username/organization information again. Once the information is entered, it is stored and used for subsequent runs.

#### Sample configuation

```yaml
# Individual account where user_1 is used for repository access
- name: user_1
  type: 0
  authUser: user_1
# Organization account where my_org_name is used for repository access
- name: user_2
  type: 1
  authUser: my_org_name
```

## Development

### Running the extension locally

```sh
make
gh bulk
```

> [!NOTE]
> Make sure you uninstall the extension before running it locally

```sh
gh extension remove bulk
```

or

```sh
make uninstall
```
