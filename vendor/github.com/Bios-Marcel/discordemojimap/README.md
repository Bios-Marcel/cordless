# Discord Emoji Map

[![builds.sr.ht status](https://builds.sr.ht/~biosmarcel/discordemojimap/arch.yml.svg)](https://builds.sr.ht/~biosmarcel/discordemojimap/arch.yml?)
[![GoDoc](https://godoc.org/github.com/Bios-Marcel/discordemojimap?status.svg)](https://godoc.org/github.com/Bios-Marcel/discordemojimap)
[![codecov](https://codecov.io/gh/Bios-Marcel/discordemojimap/branch/master/graph/badge.svg)](https://codecov.io/gh/Bios-Marcel/discordemojimap)

This is the map of emojis that discord uses. However, I have left out
different skin tones and such. A complete map might follow at some
point.

## Usage

The usage is quite simple, you just pass your input string and it replaces all
valid emoji sequences with their respective emojis.

```go
package main

import (
    "fmt"
    "github.com/Bios-Marcel/discordemojimap"
)

func main() {
    fmt.Println(discordemojimap.Replace("What a wonderful day :sun_with_face:, am I right?"))
}
```