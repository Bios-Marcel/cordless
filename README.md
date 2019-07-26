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

- [Cordless](#cordless)
  - [Overview](#overview)
  - [How to install it](#how-to-install-it)
    - [Installing on Linux](#installing-on-linux)
      - [Snap](#snap)
      - [Arch based Linux distributions](#arch-based-linux-distributions)
    - [Installing on Windows](#installing-on-windows)
    - [Installing on macOS](#installing-on-macos)
    - [Using pre-built binaries](#using-pre-built-binaries)
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
by Discord (Hammer & Chisel).

[Small demo video made by me](https://peertube.social/videos/watch/15ae8076-2de6-4f97-8947-93d8b356ad08)

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

### Using pre-built binaries

**UPDATES HAVE TO BE INSTALLED MANUALLY**

You can always find the latest release on this repositories
[release page](../../releases/latest).

For your information, since May 22th 2019 all binaries will be build using
the following parameters: `-ldflags="-w -s"`. So checksums of binaries built
without those parameters will not align with the ones built with those
parameters. Meaning that if you want to compare checksums, please keep this
in mind.

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

If you are logging in with a bot token, you have to append `Bot ` (yes, with
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
intersted in it.**

## Contributing

All kinds of contributions are welcome. Whether it's correcting typos, fixing
bugs, adding features or whatever else might be good for the project. If you
want to contribute code, please create a new branch and commit only changes
relevant to your planned pull request onto that branch. This will help
isolating new changes and make merging those into `master` easier.

I also encourage you to report anything you deem a bug, because that means
that there might problems with the UX that could still be worked on. Obviously
feature requests are welcome as well, no matter if those are features that the
official discord client has or not.

Oh and please try to keep things pragmatic and foul-language free ;)

If none of those ways of contributing are your kind of thing, feel free to
donate something via [Liberapay](https://liberapay.com/biosmarcel/donate).
It may not directly have an impact on the project, but it will surely motivate
me to keep working on this project, as it shows that people care about it.

## Why should or shouldn't you use this project

Reasons to use it:

- Your PC is not very powerful
- You're on a mobile device and value your battery life
- You just like terminal applications
- You are scared that the discord client communicates too much with the HQ

Reasons not to use it:

- You like fancy GUI
- You want to see images, videos and whatnot inside of the application itself
- You need the voice/video calling features
- You need to administrate a server (no adminsitration features yet)

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
the projects issue tracker.

In case that you simply can't use any shortcuts that the application has, this
might be due to your terminal emulator accepting those instead of letting
cordless handle them.

If you need help or have questions that you don't want to create an issue for.
feel free to hit me up on Discord: `Marcel#7299`.
