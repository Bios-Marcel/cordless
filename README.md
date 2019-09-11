# Cordless

| OS | Build-Status |
| - |:- |
| linux | [![CircleCI](https://circleci.com/gh/Bios-Marcel/cordless.svg?style=svg)](https://circleci.com/gh/Bios-Marcel/cordless) |
| darwin | [![Build Status](https://travis-ci.org/Bios-Marcel/cordless.svg?branch=master)](https://travis-ci.org/Bios-Marcel/cordless) |
| windows | [![Build status](https://ci.appveyor.com/api/projects/status/svv866htsr33hdoh/branch/master?svg=true)](https://ci.appveyor.com/project/Bios-Marcel/cordless/branch/master) |
| freebsd | [![builds.sr.ht status](https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml.svg)](https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml?) |

[![codecov](https://codecov.io/gh/Bios-Marcel/cordless/branch/master/graph/badge.svg)](https://codecov.io/gh/Bios-Marcel/cordless)
[![Discord](https://img.shields.io/discord/600329866558308373.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2)](https://discord.gg/fxFqszu)

## Overview

- [Credits](#credits)
- [How to install it](#how-to-install-it)
- [Installing on Linux](#installing-on-linux)
  - [Snap](#snap)
  - [Arch based Linux distributions](#arch-based-linux-distributions)
- [Installing on Windows](#installing-on-windows)
- [Installing on macOS](#installing-on-macos)
- [Building it from source](#building-it-from-source)
- [Login](#login)
  - [Retrieving your token via discord or chrome](#retrieving-your-token-via-discord-or-chrome)
- [Quick overview - Navigation (switching between boxes / containers)](#quick-overview---navigation-switching-between-boxes--containers)
- [Extending Cordless via the scripting interface](#extending-cordless-via-the-scripting-interface)
- [Contributing](#contributing)
- [Why should or shouldn't you use this project](#why-should-or-shouldnt-you-use-this-project)
- [Similar projects](#similar-projects)
- [Troubleshooting](#troubleshooting)

Cordless is supposed to be a custom [Discord](https://discordapp.com) client
that aims to have a low memory footprint and be aimed at power-users.

**WARNING: Self-bots are discouraged and against Discords TOS.**

This project was mainly inspired by [Southclaws](https://github.com/Southclaws)
[Cordless](https://github.com/Southclaws/cordless-old), which he sadly didn't
develop any further.

The application only uses the official Discord API and doesn't send data to
any third party. However, this application is not a official product
by Discord Inc.

![Demo Screenshot](.github/images/chat-demo.png)

## Credits

Big thanks to [JetBrains](https://www.jetbrains.com/?from=cordless) for providing the
cordless project with free licenses!

## How to install it

### Installing on Linux

#### Snap

On linux the recommended way of installation is the snap.

Simply run (Might require sudo):

```shell
snap install cordless
```

Snap will automatically install updates.

#### Arch based Linux distributions

If you are on an arch based distribution, you could use the AUR package:

Manually:

```shell
$ git clone https://aur.archlinux.org/cordless-git.git
$ cd cordless-git
$ makepkg -sric
```

With yay:

```shell
$ yay -Syu cordless-git
```

#### Manual Linux installation

If you are installing manually via:

```sh
go get -u github.com/Bios-Marcel/cordless
```

then you also need xclip in order to be able to copy and paste properly.

There is currently no wayland support for copy and paste.

### Installing on Windows

In order to install the latest version on Windows, you first need
[scoop](https://scoop.sh/#installs-in-seconds).

After installing scoop, run the following:

```ps1
scoop bucket add biosmarcel https://github.com/Bios-Marcel/scoopbucket.git
```

This adds the bucket (repository) to your local index, allowing you to
install any package from that bucket.

Install cordless via

```ps1
scoop install cordless
```

In order to install updates, run:

```ps1
scoop update cordless
```

### Installing on macOS

Use [Homebrew](https://brew.sh) to install `cordless` on macOS:

```shell
brew tap Bios-Marcel/cordless
brew install cordless
```

If you don't install via homebrew, then you should get `pngpaste`, since it's
what allows you to paste images.

### Building it from source

In order to execute this command
[you need to have go 1.12 or a more recent version installed](https://golang.org/doc/install).

**UPDATES HAVE TO BE INSTALLED MANUALLY**

First you have to grab the code via:

```shell
go get -u github.com/Bios-Marcel/cordless
```

In order to execute the application, simply run the executable, which lies at
`$GOPATH/bin/cordless`. In order to be able to run this from your terminal,
`$GOPATH/bin` has to be in your `PATH` variable.

### Login

On launch, cordless will offer you two login methods:

1. Using an authentication token
2. Using email and password

I recommend the first way, as the second one won't work anyway in case you have
two-factor authentication enabled. After logging in using either method, your
token is stored locally on your machine. The token will not be encrypted, so be
careful with your configuration file.

If you are logging in with a bot token, you have to prepend `Bot ` (yes, with
the space) in front of the token.

**THIS APPLICATION NEVER SAVES YOUR ACTUAL PASSWORD LOCALLY.**

If you need to find out how to retrieve your token, [check the wiki at](https://github.com/Bios-Marcel/cordless/wiki/Retrieving-your-token-via-the-discord-app-or-a-chromium-based-browser).

## Quick overview - Navigation (switching between boxes / containers)

| Shortcut | Action |
| - |:- |
| <kbd>Alt</kbd> + <kbd>S</kbd> | Sets the focus on the servers (guilds) container |
| <kbd>Alt</kbd> + <kbd>C</kbd> | Sets the focus on the channels container |
| <kbd>Alt</kbd> + <kbd>T</kbd> | Sets the focus on the messages container |
| <kbd>Alt</kbd> + <kbd>M</kbd> | Sets the focus on the messages input field |
| <kbd>Alt</kbd> + <kbd>U</kbd> | Sets the focus on the users container |
| <kbd>Alt</kbd> + <kbd>P</kbd> | Opens the direct messages container |
| <kbd>Alt</kbd> + <kbd>.</kbd> | Toggles the internal console view |

Further shortcuts / key-bindings can be found in the manual on the internal
console with the command `manual`.

If any of the default commands don't work for you, open the keyboard shortcut
changer via <kbd>Alt</kbd> + <kbd>Shift</kbd> + <kbd>S</kbd>.

## Extending Cordless via the scripting interface

Cordless has a very basic scripting interface that exposes predefined events.
Scripts can simply be dumped into the subfolder `scripts` of the cordless
configuration folder.

An example can be found here:
[Kaomoji](https://github.com/Bios-Marcel/cordless-kaomoji)

**Currently the scripting interface only offers a single event, if people show
interest in this feature, I might add more, as I am currently not very
interested in it.**

## Contributing

All kinds of contributions are welcome. Whether it's correcting typos, fixing
bugs, adding features or whatever else might be good for the project. If you
want to contribute code, please create a new branch and commit only changes
relevant to your planned pull request onto that branch. This will help
isolating new changes and make merging those into `master` easier.

If you encounter any issues, whether it's bugs or the lack of certain features,
don't hesitate to create a new GitHub issue.

If there are specific issues you want to be solved quickly, you can set a
bounty on those via [IssueHunt](https://issuehunt.io/r/Bios-Marcel/cordless).
The full 100% of the bounty goes to whoever solves the issue, no matter
whether that's me or someone else.

If none of those ways of contributing are your kind of thing, feel free to
donate something via [Liberapay](https://liberapay.com/biosmarcel/donate).
It may not directly have an impact on the project, but it will surely motivate
me to keep working on this project, as it shows that people care about it.

## Why should or shouldn't you use this project

Reasons to use it:

- Your PC is not very powerful
- You're on a mobile device and value your battery life
- You want to reduce your bandwidth usage
- You just like terminal applications

Reasons not to use it:

- You like fancy GUI
- You want to see images, videos and whatnot inside of the application itself
- You need the voice/video calling features (This might soon change!)
- You need to administrate a server (no administration features yet)

## Similar projects

Here is a list of similar projects:

- [terminal-discord](https://github.com/xynxynxyn/terminal-discord)
- [Discurses](https://github.com/topisani/Discurses)
- [Discline](https://github.com/MitchWeaver/Discline)
- [discord-term](https://github.com/cloudrex/discord-term)
- [6cord](https://gitlab.com/diamondburned/6cord)

Hit me up if you have a similar project and I'll gladly add it to the list.

## Troubleshooting

If you happen to encounter a crash or a bug, please submit a bug report via
the projects GitHub issue tracker.

For general problems faced by cordless users, check out the wiki at:
https://github.com/Bios-Marcel/cordless/wiki/Troubleshooting

If you need help or have questions that you don't want to create an issue for.
feel free to hit me up on Discord: `Marcel#7299`. Alternatively, just join the
public Cordless server linked at the top of the Readme.