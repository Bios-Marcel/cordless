# Cordless

| OS | Build-Status |
| - |:- |
| linux | [![CircleCI](https://circleci.com/gh/Bios-Marcel/cordless.svg?style=svg)](https://circleci.com/gh/Bios-Marcel/cordless) |
| darwin | [![Build Status](https://travis-ci.org/Bios-Marcel/cordless.svg?branch=master)](https://travis-ci.org/Bios-Marcel/cordless) |
| windows | [![Build status](https://ci.appveyor.com/api/projects/status/svv866htsr33hdoh/branch/master?svg=true)](https://ci.appveyor.com/project/Bios-Marcel/cordless/branch/master) |
| freebsd | [![builds.sr.ht status](https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml.svg)](https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml?) |

[![Donate using librepay](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/biosmarcel/donate)

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
any third party. However, this application is not an product official product
by Discord (Hammer & Chisel).

This application is currently a WIP and will change rather fast.

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

```shell
yaourt -S cordless-git
```

Depending on your installation, you might not have `yaourt` installed or even
have a different AUR package manager.

### Installing on Windows

In order to install the latest version on Windows, you first need
[scoop](https://scoop.sh/#installs-in-seconds).

After installing scoop, run the following:

```ps1
scoop install https://raw.githubusercontent.com/Bios-Marcel/cordless/master/cordless.json
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

### Using pre-built binaries

**UPDATES HAVE TO BE INSTALLED MANUALLY**

You can always find the latest release in this repositories
[release page](../../releases/latest).

### Building it from source

**UPDATES HAVE TO BE INSTALLED MANUALLY**

First you have to grab the code via:

```shell
go get github.com/Bios-Marcel/cordless
```

In order to execute this command
[you need to have go installed](https://golang.org/doc/install).

In order to execute the application, simply run the executable, which lies at
`$GOPATH/bin/cordless`. In order to be able to run this from your terminal,
`$GOPATH/bin` has to be in your `PATH` variable.

### Login

On launch, cordless will offer you two login methods:

1. Using an authentication token
2. Using email and password

I recommend the first way, as the second one won't work anyway in case you have
two-factor authentication enabled. After logging on using either method, your
token is stored locally on your machine. The token will not be encrypted, so be
careful with your configuration file.

If you are logging in with a bot token, you have to append `Bot ` (yes, with
the space) in front of the token.

**THIS APPLICATION NEVER SAVES YOUR PASSWORD LOCALLY.**

#### Retrieving your token via discord or chrome

* Press `CTRL+SHIFT+I` (or `COMMAND+SHIFT+I` if you are on Mac OS)

![Default window](https://user-images.githubusercontent.com/19377618/53696114-2b3fda00-3dc4-11e9-9111-50a1e77ca838.png)

* Click `Network` section

![Network section](https://user-images.githubusercontent.com/19377618/53696066-ac4aa180-3dc3-11e9-8585-df9f579a44a6.png)

* Click `XHR` Section

![XHR section](https://user-images.githubusercontent.com/19377618/53696115-2d099d80-3dc4-11e9-914f-6bb3769853f9.png)

* Reload it by clicking `F5` and choose `access-token`

![XHR section after reloading it](https://user-images.githubusercontent.com/19377618/53696068-aeacfb80-3dc3-11e9-8af7-8f93fd226eff.png) 

* Get into `Headers` section and scroll down until you find `authorization: token` and copy it
 
![Headers section of access token](https://user-images.githubusercontent.com/19377618/53696070-afde2880-3dc3-11e9-8859-d307677f51de.png)

* Paste the token into the terminal and hit enter

![Terminal after pasting the token](https://user-images.githubusercontent.com/19377618/53696072-b1a7ec00-3dc3-11e9-9ce5-d4d2534602de.png)

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
donate something via https://liberapay.com/biosmarcel/donate. It may not
directly have an impact on the project, but it will surely motivate me to
keep working on this project, as it shows that people care about it.

## Why should or shouldn't you use this project

Reason to use it:
  * Your PC is not very powerful
  * Your on a mobile device and value your battery life
  * You just like terminal applications
  * You are scared that the discord client communicates too much with the HQ

Reasons not to use it:
  * You like fancy GUI
  * You want to see images, videos and whatnot inside of the application itself
  * You need the voice/video calling features
  * You need to administrate a server (no adminsitration features yet)

## Similar projects

Here is a list of similar projects:

* [terminal-discord](https://github.com/xynxynxyn/terminal-discord)
* [Discurses](https://github.com/topisani/Discurses)
* [Discline](https://github.com/MitchWeaver/Discline)
* [discord-term](https://github.com/cloudrex/discord-term)
* [6cord](https://github.com/cloudrex/6cord)

Hit me up if you have a similar project as well :D

## Troubleshooting

If you happen to encounter a crash or a bug, please submit a bug request.

In case that you simply can't use any shortcuts that the application has, this
might be due to your terminal emulator accepting those instead of letting
cordless handle them.

Or message me on Discord at `Marcel#7299`.
