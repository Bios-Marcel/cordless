# I AM CLOSING DOWN THE CORDLESS PROJECT

Hey, so I know this is somewhat of a bummer, but I got banned because of ToS violation today. This seemed to be connected to creating a new PM channel via the `/users/@me` endpoint. As that's basically a confirmation for what we've believed would never be enforced, I decided to not work on the cordless project anymore. I'll be taking down cordless in package managers in hope that no new users will install it anymore without knowing the risks. I believe that if you manage to build it yourself, you've probably read the README and are aware of the risks.
I'll keep the repository up, but it'll be archived (read-only) and I have vendored the dependencies, meaning that you'll probably always be able to build the project from source as long as you have a compatible go compiler. **And yes, you'll still be able to use existing binaries for as long as discord doesn't introduce any more breaking changes. However, be aware that the risk of getting a ban will only get higher with time!**

<h1 align="center">Cordless</h1>

The discord server still exists and there's still some people talking, so feel free to check it out if you want to:

https://discord.gg/fxFqszu

## Overview

- [How to install it](#installation)
  - [Using prebuilt binaries](#using-prebuilt-binaries)
  - [Building from source](#building-from-source)
- [Login](#login)
- [Quick overview - Navigation (switching between boxes / containers)](#quick-overview---navigation-switching-between-boxes--containers)
- [Extending Cordless via the scripting interface](#extending-cordless-via-the-scripting-interface)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [This project isn't for you, if](#this-project-isnt-for-you-if)
- [Similar projects](#similar-projects)
- [Credits](#credits)

**WARNING: Third party clients are discouraged and against the Discord TOS. There have already been cordless users that got banned, including me (Bios-Marcel, the maker and maintainer)**

Cordless is a custom [Discord](https://discord.com/app) client that aims to
have a low memory footprint and be aimed at power-users.

The application only uses the official Discord API and doesn't send data to
third parties. However, this application is not an official product by
Discord Inc.

![Demo Screenshot](.github/images/chat-demo.png)

## Installation

### Using prebuilt binaries

If you don't want to build the application yourself or use some kind of
package management system, you can get the latest binaries for the three
major systems in the release overview:

https://github.com/Bios-Marcel/cordless/releases/latest

### Building from source

In order to execute the following commands, you need to install **go 1.13 or**
higher. You can find golang packages at https://golang.org/doc/install.
On top of that, you need to have **git** installed. It can be fund at
https://git-scm.com/downloads.

**UPDATES HAVE TO BE INSTALLED MANUALLY**

Open a command line and execute the following commands:

```shell
git clone https://github.com/Bios-Marcel/cordless.git
cd cordless
go build
```

This will create an executable file called `cordless` or `cordless.exe`
depending on whether you are on Windows or not. Move that file anywhere
 that your terminal can find it. I recommend adding a `bin` folder to your
user home and adding it to your systems `PATH` variable. Please search the
internet, using your favourite search engine, for
`how to set an environment variable in XXX` in order to update your `PATH`
variable correctly.

For updateing you simply have to delete the folder you downloaded last
time and repeat the instructions.

Note:

* X11 users need `xclip` in order to copy and paste.
* Wayland users need `wl-clipboard` in order to copy and paste.
* Mac OS users need `pngpaste` in order to copy and paste images.

### Login

**YOUR PASSWORD IS NEVER SAVED LOCALLY.**

Logging in works via the UI on first startup of the application.

If you are logging in with a bot token, you have to prepend `Bot` in front of
the token.

If you need to find out how to retrieve your token, check [the wiki](https://github.com/Bios-Marcel/cordless/wiki/Retrieving-your-token).

**Currently captcha-code login isn't supported. Thanks for your SHIT-API, Google**

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

## Extending Cordless via the scripting interface

[Check the wiki](https://github.com/Bios-Marcel/cordless/wiki/Extending-Cordless-via-the-scripting-interface)

## Troubleshooting

If you happen to encounter a crash or a bug, please submit a bug report via
the projects GitHub issue tracker. Bugs reported via Discord will probably
be forgotten or overseen.

For general problems faced by cordless users, check out the wiki at:
https://github.com/Bios-Marcel/cordless/wiki/Troubleshooting

If you need help or have questions that you don't want to create an issue
for, just join our Discord server: https://discord.gg/fxFqszu

# FAQ

In order to find answers to common questions, check out the FAQ in the wiki:

https://github.com/Bios-Marcel/cordless/wiki/FAQ

## This project isn't for you, if

- You like fancy GUI
- You want to see images, videos and whatnot inside of the application itself
- You need the voice/video calling features
- You need to manage or moderate servers

## Similar projects

Here is a list of similar projects:

- [terminal-discord](https://github.com/xynxynxyn/terminal-discord)
- [Discurses](https://github.com/topisani/Discurses)
- [Discline](https://github.com/MitchWeaver/Discline)
- [discord-term](https://github.com/cloudrex/discord-term)
- [6cord](https://gitlab.com/diamondburned/6cord)

Hit me up if you have a similar project, and I'll gladly add it to the list.

## Credits

Big thanks to [JetBrains](https://www.jetbrains.com/?from=cordless) for providing the
cordless project with free licenses!

This project was mainly inspired by [Southclaws](https://github.com/Southclaws)
[Cordless](https://github.com/Southclaws/cordless-old), which he sadly didn't
develop any further.
