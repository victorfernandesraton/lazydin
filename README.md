# Lazydin

## A tool for automatization boring stuff in linkedin

CLI for interacting with Linkedin

Usage:
  lazydin [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  create-credentials Start proccess to define credentials in config credentials file
  create-storage     Start proccess to define path to storage file
  help               Help about any command
  post-comment       Post a comment on a Linkedin post
  search-posts       Search for posts on Linkedin

Flags:
  -c, --config string        Configguration path
      --credentials string   Credential file storage in toml
  -h, --help                 help for lazydin
  -p, --password string      Linkedin Password
  -u, --user string          Linkedin Username

Use "lazydin [command] --help" for more information about a command.

## Features

- Search for job posts related
- Touch in contact with recruiter who publish related job posts

## Requiements

- Golang 1.20+
- Google chrome or chromiun avaliable for current user and installs as normal host sofware (not support for flatpacks , distrobox , snap or any container format)

## Before starting
- Make sure your credntials is stored correctly and update with
- Make sure you __disable__ MFA security in linkedin (but not forgot to put back when you finish)

### Warning

I am writing to inform you that the software I have developed is intended for personal use only and should not be used to automate activities that may be considered harmful or inappropriate on LinkedIn. By using this software, you acknowledge and agree that you will not use it to engage in any activities that may be considered spamming, harassment, or other forms of abuse. Additionally, you understand and agree that you are solely responsible for any damages or liabilities that may arise from your use of this software, and that you will not hold me for any such damages or liabilities.

Please note that LinkedIn has strict policies against using automated software to access or interact with their platform, and any violation of these policies may result in your account being suspended or terminated. By using this software, you acknowledge and agree that you are aware of these policies and will comply with them.
