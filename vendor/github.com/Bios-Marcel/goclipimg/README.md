# goclipimg

| OS | Build |
| - | - |
| linux | [![builds.sr.ht status](https://builds.sr.ht/~biosmarcel/goclipimg/arch.yml.svg)](https://builds.sr.ht/~biosmarcel/goclipimg/arch.yml?) |
| darwin | [![Build Status](https://travis-ci.org/Bios-Marcel/goclipimg.svg?branch=master)](https://travis-ci.org/Bios-Marcel/goclipimg) |
| windows | [![Build status](https://ci.appveyor.com/api/projects/status/jk8g0q27hle7m98p/branch/master?svg=true)](https://ci.appveyor.com/project/Bios-Marcel/goclipimg/branch/master) |

This is just a tiny library that helps you getting an image from your clipboard into your application.

## Requirements

### Requirements - Linux

If you are running x11 you'll need to have `xclip` installed.

On Wayland you need `wl-clipboard`.

## Example

```go
func main() {
    data, readError := goclipimg.GetImageFromClipboard()
    if readError == nil {
        doSomethingWithYourPNG(data)
    }
}
```