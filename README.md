<h1 align="center">Cordless</h1>

<p align="center">
  <a href="https://circleci.com/gh/Bios-Marcel/cordless">
    <img src="https://img.shields.io/circleci/build/gh/Bios-Marcel/cordless?label=linux&logo=linux&logoColor=white">
  </a>
  <a href="https://travis-ci.org/Bios-Marcel/cordless">
    <img src="https://img.shields.io/travis/Bios-Marcel/cordless?label=darwin&logo=apple&logoColor=white">
  </a>
  <a href="https://ci.appveyor.com/project/Bios-Marcel/cordless/branch/master">
    <img src="https://img.shields.io/appveyor/ci/Bios-Marcel/cordless?label=windows&logo=windows&logoColor=white">
  </a>
  <a href="https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml?">
    <img src="https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml.svg">
  </a>
  <a href="https://codecov.io/gh/Bios-Marcel/cordless">
     <img src="https://codecov.io/gh/Bios-Marcel/cordless/branch/master/graph/badge.svg">
  </a>
  <a href="https://discord.gg/fxFqszu">
     <img src="https://img.shields.io/discord/600329866558308373.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2">
  </a>
</p>

## Overview

- [Credits](#credits)
- [Introduction](#introduction)
- [Installing](#installing)
  - [Linux](#linux)
  - [Windows](#installing-on-windows)
  - [macOS](#installing-on-macos)
  - [Building from source](#building-from-source)
- [Login](#login)
- [Quick overview - Navigation (switching between boxes / containers)](#quick-overview---navigation-switching-between-boxes--containers)
- [Contributing](#contributing)
- [Similar projects](#similar-projects)
- [Troubleshooting](#troubleshooting)

## Credits
**Special thanks to**
- [JetBrains](https://www.jetbrains.com/?from=cordless) for providing the cordless project with free licenses.
- [Southclaws](https://github.com/Southclaws) for inspiring the project with [Cordless](https://github.com/Southclaws/cordless-old).

## Introduction

Cordless is a custom third-party [Discord](https://discordapp.com) client
aimed towards a low memory footprint and power-users.

**WARNING: Self-bots are discouraged and against Discords TOS. The application only uses the official Discord API and doesn't send data to any third party. However, this application is not a official product by Discord Inc.**
 
Cordless is:  
- Lightweight (can reduce usage of battery and bandwidth)
- Tweakable (comes with [scripting](https://github.com/Bios-Marcel/cordless/wiki/Extending-Cordless-via-the-scripting-interface) and [theming](https://github.com/Bios-Marcel/cordless/wiki/Themes) (coming soon) out of the box)
- An interesting TUI

However, cordless doesn't have:
- A fancy GUI (no image or video playback support)
- Voice/video calling features (This might soon change!)
- Server Administration features (This might soon change!)

![Demo Screenshot](.github/images/chat-demo.png)

## Installing

### Linux

<details>
  <summary>Snap (Recommended)</summary>
  
   Simply run (Might require sudo):
   ```shell
   snap install cordless
   ```
   Snap will automatically install updates.  
  
</details>

<details>
  <summary>AUR</summary>
  
  On arch based distributions, you can use the AUR package to install cordless
  
  <details>
    <summary>Manually</summary>
  
   ```shell
   $ git clone https://aur.archlinux.org/cordless-git.git
   $ cd cordless-git
   $ makepkg -sric
   ```
    
  </details>
  <details>
    <summary>AUR Helpers</summary>
    
   ###### yay:
   ```shell
   $ yay -Syu cordless-git
   ```
   ```shell
   $ yay -S cordless-git
   ```

   ###### trizen:
   ```shell
   $ trizen -S cordless-git
   ```

   ###### pacaur
   ```shell
   $ pacuar -S cordless-git
   ```
    
  </details>
</details>

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

### Building from source

In order to execute the following commands
[you need to have go 1.12 or a more recent version installed](https://golang.org/doc/install).

**UPDATES HAVE TO BE INSTALLED MANUALLY**

You can either install the binary into your `$GOPATH/bin` by running:

```shell
go get -u github.com/Bios-Marcel/cordless
```

Which you can then execute by running the executable, which lies at
`$GOPATH/bin/cordless`. In order to be able to run this from your terminal,
`$GOPATH/bin` has to be in your `PATH` variable. The very same command can
be used for updating.

or you manually grab the source:

```shell
git clone https://github.com/Bios-Marcel/cordless
cd cordless
go build .
```

If done this way, updates have to be installed via:

```shell
cd cordless
git pull
go build .
```

Note:
* X11 users need `xclip` in order to copy and paste.
* Wayland users need `wl-clipboard` in order to copy and paste.

### Login

Logging in works via the UI on first startup of the application.

If you are logging in with a bot token, you have to prepend `Bot ` (yes, with
the space) in front of the token. Pasting only works via your terminals normal pasting shortcut.

**THIS APPLICATION NEVER SAVES YOUR ACTUAL PASSWORD LOCALLY.**

If you need to find out how to retrieve your token, [check the wiki at](https://github.com/Bios-Marcel/cordless/wiki/Retrieving-your-token).

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
changer via <kbd>Ctrl</kbd> + <kbd>K</kbd>.

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
donate something via [Liberapay](https://liberapay.com/biosmarcel/donate) or use 
my ETH public key 0x49939106563a9de8a777Cf5394149423b1dFd970.
It may not directly have an impact on the project, but it will surely motivate
me to keep working on this project, as it shows that people care about it.

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
