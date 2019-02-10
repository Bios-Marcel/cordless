# Cordless

| OS | Build-Status |
| - |:- |
| linux | [![CircleCI](https://circleci.com/gh/Bios-Marcel/cordless.svg?style=svg)](https://circleci.com/gh/Bios-Marcel/cordless) |
| darwin | [![Build Status](https://travis-ci.org/Bios-Marcel/cordless.svg?branch=master)](https://travis-ci.org/Bios-Marcel/cordless) |
| windows | [![Build status](https://ci.appveyor.com/api/projects/status/svv866htsr33hdoh/branch/master?svg=true)](https://ci.appveyor.com/project/Bios-Marcel/cordless/branch/master) |
| freebsd | [![builds.sr.ht status](https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml.svg)](https://builds.sr.ht/~biosmarcel/cordless/freebsd.yml?) |

## Overview

* How to install it
  * [Using pre-built binaries](https://github.com/Bios-Marcel/cordless#using-pre-built-binaries)
  * [Building it from source](https://github.com/Bios-Marcel/cordless#building-it-from-source)
* [Login](https://github.com/Bios-Marcel/cordless#login)
* [Features](https://github.com/Bios-Marcel/cordless#features)
* [Extending Cordless via the scripting interface](https://github.com/Bios-Marcel/cordless#extending-cordless-via-the-scripting-interface)
* [Similar projects](https://github.com/Bios-Marcel/cordless#similar-projects)
* [Troubleshooting](https://github.com/Bios-Marcel/cordless#troubleshooting)

Cordless is supposed to be a custom [Discord](https://discordapp.com) client
that aims to have a low memory footprint and be aimed at powerusers.

**WARNING: Self-bots are discouraged and against Discords TOS.**

This project was mainly inspired by [Southclaws](https://github.com/Southclaws)
[Cordless](https://github.com/Southclaws/cordless-old), which he sadly didn't
develop any further.

The application only uses the official Discord API and doesn't send data to
any third party. However, this application is not an product official product
of Discord (Hammer & Chisel).

This application is currently a WIP and will change rather fast.

![Cordless 26th January 2019](https://i.imgur.com/xX7dVCw.png)

## How to install it

### Using pre-built binaries

Currently every commit triggers a build for windows and linux, those builds
each produce ready to use binaries. The builds can be found at the top, by
clicking on the respective builds badges.

### Building it from source

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
2. Using username and password

I recommend the first way, as the second one won't work anyway in case you have
two-factor authentication enabled. After logging on using either method, your
token is stored locally on your machine. The token will not be encrypted, so be
careful with your configuration file.

**THIS APPLICATION NEVER SAVES YOUR PASSWORD LOCALLY.**

### In order to retrieve your token from browser or from discord client, simply follow the steps below:

* Press **CTRL+SHIFT+I** or **COMMAND+SHIFT+I** depends on your [OS](https://en.wikipedia.org/wiki/Operating_system).

![Imgur](https://i.imgur.com/38UF1h5.png))

* Click **Network** section.

![Imgur](https://i.imgur.com/k6OhJHt.png)

* Click **XHR** Section

![Imgur](https://i.imgur.com/HXqL7Pp.png)

* Reload it by clicking **F5** and choose **access-token**

![Imgur](https://i.imgur.com/Rkb2krO.png) 

* Get into **Headers** section and scroll down until you find **authorization: <token>**, then *copy* it
 
![Imgur](https://i.imgur.com/PEox6bP.png)

* Paste into terminal, hit enter and there you go !

![Imgur](https://i.imgur.com/UpsrGJt.png)

## Features

*This list might be incomplete.*

* Guilds
  * Enter channels
  * See channels
  * See members
* Channels
  * See NSFW flag
  * See group
  * See topic
* Members
  * See Nickname
  * See hoist group
* Chatting
  * conversation with a friend
  * Groupchats
  * Talk in a channel
  * Timestamps `HH:MM(:SS)`
* See all your friends
* Messages
  * Send messages
  * Mention people using their full username or nickname
  * Mention a channel
  * Edit last message
  * Edit any message
  * Delete last message
  * Delete any message
  * Quoting
  * Copying
  * Spoiler
  * Multiline
  * Highlighting
    * Code
    * Mentions (Channels/Users)
  * Special highlighting when you get mentioned
  * Send Emojis with `:name:`
* Custom layout to a specific point
* Notifications 
  * System notifications on mention (not considering muted servers and such)
  * Mark channels read when unread messages in the current session exist
  * Prefix channel with `(@You)` if you got mentioned
* Basic scripting interface
  * Languages
    * JavaScript
    * More to come ...
  * Hooks
    * onMessageSend Event that can manipulate the message before sending

## Extending Cordless via the scripting interface

Cordless has a very basic scripting interface that exposes predefined events.
Scripts can simply be dumped into the subfolder `scripts` of the cordless
config folder.

An example can be found here:
[Kaomoji](https://github.com/Bios-Marcel/cordless-kaomoji)

## Similar projects

Here is a list of similar projects:

* [terminal-discord](https://github.com/xynxynxyn/terminal-discord)
* [Discurses](https://github.com/topisani/Discurses)
* [Discline](https://github.com/MitchWeaver/Discline)

## Troubleshooting

If you happen to encounter a crash or a bug, please submit a bug request.

In case that you simply can't use any shortcuts that the application has, this
might be due to your terminal emulator accepting those instead of letting
cordless handle them.
