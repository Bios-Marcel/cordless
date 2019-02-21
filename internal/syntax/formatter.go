package syntax

import (
	"fmt"
	"io"
	"math"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
)

type ttyTable struct {
	foreground map[chroma.Colour]string
	background map[chroma.Colour]string
}

var c = chroma.MustParseColour

var ttyTables = map[int]*ttyTable{
	8: {
		foreground: map[chroma.Colour]string{
			c("#000000"): "[#000000]", c("#7f0000"): "[#7f0000]", c("#007f00"): "[#007f00]", c("#7f7fe0"): "[#7f7fe0]",
			c("#00007f"): "[#00007f]", c("#7f007f"): "[#7f007f]", c("#007f7f"): "[#007f7f]", c("#e5e5e5"): "[#e5e5e5]",
			c("#555555"): "[#555555]", c("#ff0000"): "[#ff0000]", c("#00ff00"): "[#00ff00]", c("#ffff00"): "[#ffff00]",
			c("#0000ff"): "[#0000ff]", c("#ff00ff"): "[#ff00ff]", c("#00ffff"): "[#00ffff]", c("#ffffff"): "[#ffffff]",
		},
		background: map[chroma.Colour]string{
			c("#000000"): "[#000000]", c("#7f0000"): "[#7f0000]", c("#007f00"): "[#007f00]", c("#7f7fe0"): "[#7f7fe0]",
			c("#00007f"): "[#00007f]", c("#7f007f"): "[#7f007f]", c("#007f7f"): "[#007f7f]", c("#e5e5e5"): "[#e5e5e5]",
			c("#555555"): "[#555555]", c("#ff0000"): "[#ff0000]", c("#00ff00"): "[#00ff00]", c("#ffff00"): "[#ffff00]",
			c("#0000ff"): "[#0000ff]", c("#ff00ff"): "[#ff00ff]", c("#00ffff"): "[#00ffff]", c("#ffffff"): "[#ffffff]",
		},
	},
	256: {
		foreground: map[chroma.Colour]string{
			c("#000000"): "[#000000]", c("#800000"): "[#800000]", c("#008000"): "[#008000]", c("#808000"): "[#808000]",
			c("#000080"): "[#000080]", c("#800080"): "[#800080]", c("#008080"): "[#008080]", c("#c0c0c0"): "[#c0c0c0]",
			c("#808080"): "[#808080]", c("#ff0000"): "[#ff0000]", c("#00ff00"): "[#00ff00]", c("#ffff00"): "[#ffff00]",
			c("#0000ff"): "[#0000ff]", c("#ff00ff"): "[#ff00ff]", c("#00ffff"): "[#00ffff]", c("#ffffff"): "[#ffffff]",
			c("#000000"): "[#000000]", c("#00005f"): "[#00005f]", c("#000087"): "[#000087]", c("#0000af"): "[#0000af]",
			c("#0000d7"): "[#0000d7]", c("#0000ff"): "[#0000ff]", c("#005f00"): "[#005f00]", c("#005f5f"): "[#005f5f]",
			c("#005f87"): "[#005f87]", c("#005faf"): "[#005faf]", c("#005fd7"): "[#005fd7]", c("#005fff"): "[#005fff]",
			c("#008700"): "[#008700]", c("#00875f"): "[#00875f]", c("#008787"): "[#008787]", c("#0087af"): "[#0087af]",
			c("#0087d7"): "[#0087d7]", c("#0087ff"): "[#0087ff]", c("#00af00"): "[#00af00]", c("#00af5f"): "[#00af5f]",
			c("#00af87"): "[#00af87]", c("#00afaf"): "[#00afaf]", c("#00afd7"): "[#00afd7]", c("#00afff"): "[#00afff]",
			c("#00d700"): "[#00d700]", c("#00d75f"): "[#00d75f]", c("#00d787"): "[#00d787]", c("#00d7af"): "[#00d7af]",
			c("#00d7d7"): "[#00d7d7]", c("#00d7ff"): "[#00d7ff]", c("#00ff00"): "[#00ff00]", c("#00ff5f"): "[#00ff5f]",
			c("#00ff87"): "[#00ff87]", c("#00ffaf"): "[#00ffaf]", c("#00ffd7"): "[#00ffd7]", c("#00ffff"): "[#00ffff]",
			c("#5f0000"): "[#5f0000]", c("#5f005f"): "[#5f005f]", c("#5f0087"): "[#5f0087]", c("#5f00af"): "[#5f00af]",
			c("#5f00d7"): "[#5f00d7]", c("#5f00ff"): "[#5f00ff]", c("#5f5f00"): "[#5f5f00]", c("#5f5f5f"): "[#5f5f5f]",
			c("#5f5f87"): "[#5f5f87]", c("#5f5faf"): "[#5f5faf]", c("#5f5fd7"): "[#5f5fd7]", c("#5f5fff"): "[#5f5fff]",
			c("#5f8700"): "[#5f8700]", c("#5f875f"): "[#5f875f]", c("#5f8787"): "[#5f8787]", c("#5f87af"): "[#5f87af]",
			c("#5f87d7"): "[#5f87d7]", c("#5f87ff"): "[#5f87ff]", c("#5faf00"): "[#5faf00]", c("#5faf5f"): "[#5faf5f]",
			c("#5faf87"): "[#5faf87]", c("#5fafaf"): "[#5fafaf]", c("#5fafd7"): "[#5fafd7]", c("#5fafff"): "[#5fafff]",
			c("#5fd700"): "[#5fd700]", c("#5fd75f"): "[#5fd75f]", c("#5fd787"): "[#5fd787]", c("#5fd7af"): "[#5fd7af]",
			c("#5fd7d7"): "[#5fd7d7]", c("#5fd7ff"): "[#5fd7ff]", c("#5fff00"): "[#5fff00]", c("#5fff5f"): "[#5fff5f]",
			c("#5fff87"): "[#5fff87]", c("#5fffaf"): "[#5fffaf]", c("#5fffd7"): "[#5fffd7]", c("#5fffff"): "[#5fffff]",
			c("#870000"): "[#870000]", c("#87005f"): "[#87005f]", c("#870087"): "[#870087]", c("#8700af"): "[#8700af]",
			c("#8700d7"): "[#8700d7]", c("#8700ff"): "[#8700ff]", c("#875f00"): "[#875f00]", c("#875f5f"): "[#875f5f]",
			c("#875f87"): "[#875f87]", c("#875faf"): "[#875faf]", c("#875fd7"): "[#875fd7]", c("#875fff"): "[#875fff]",
			c("#878700"): "[#878700]", c("#87875f"): "[#87875f]", c("#878787"): "[#878787]", c("#8787af"): "[#8787af]",
			c("#8787d7"): "[#8787d7]", c("#8787ff"): "[#8787ff]", c("#87af00"): "[#87af00]", c("#87af5f"): "[#87af5f]",
			c("#87af87"): "[#87af87]", c("#87afaf"): "[#87afaf]", c("#87afd7"): "[#87afd7]", c("#87afff"): "[#87afff]",
			c("#87d700"): "[#87d700]", c("#87d75f"): "[#87d75f]", c("#87d787"): "[#87d787]", c("#87d7af"): "[#87d7af]",
			c("#87d7d7"): "[#87d7d7]", c("#87d7ff"): "[#87d7ff]", c("#87ff00"): "[#87ff00]", c("#87ff5f"): "[#87ff5f]",
			c("#87ff87"): "[#87ff87]", c("#87ffaf"): "[#87ffaf]", c("#87ffd7"): "[#87ffd7]", c("#87ffff"): "[#87ffff]",
			c("#af0000"): "[#af0000]", c("#af005f"): "[#af005f]", c("#af0087"): "[#af0087]", c("#af00af"): "[#af00af]",
			c("#af00d7"): "[#af00d7]", c("#af00ff"): "[#af00ff]", c("#af5f00"): "[#af5f00]", c("#af5f5f"): "[#af5f5f]",
			c("#af5f87"): "[#af5f87]", c("#af5faf"): "[#af5faf]", c("#af5fd7"): "[#af5fd7]", c("#af5fff"): "[#af5fff]",
			c("#af8700"): "[#af8700]", c("#af875f"): "[#af875f]", c("#af8787"): "[#af8787]", c("#af87af"): "[#af87af]",
			c("#af87d7"): "[#af87d7]", c("#af87ff"): "[#af87ff]", c("#afaf00"): "[#afaf00]", c("#afaf5f"): "[#afaf5f]",
			c("#afaf87"): "[#afaf87]", c("#afafaf"): "[#afafaf]", c("#afafd7"): "[#afafd7]", c("#afafff"): "[#afafff]",
			c("#afd700"): "[#afd700]", c("#afd75f"): "[#afd75f]", c("#afd787"): "[#afd787]", c("#afd7af"): "[#afd7af]",
			c("#afd7d7"): "[#afd7d7]", c("#afd7ff"): "[#afd7ff]", c("#afff00"): "[#afff00]", c("#afff5f"): "[#afff5f]",
			c("#afff87"): "[#afff87]", c("#afffaf"): "[#afffaf]", c("#afffd7"): "[#afffd7]", c("#afffff"): "[#afffff]",
			c("#d70000"): "[#d70000]", c("#d7005f"): "[#d7005f]", c("#d70087"): "[#d70087]", c("#d700af"): "[#d700af]",
			c("#d700d7"): "[#d700d7]", c("#d700ff"): "[#d700ff]", c("#d75f00"): "[#d75f00]", c("#d75f5f"): "[#d75f5f]",
			c("#d75f87"): "[#d75f87]", c("#d75faf"): "[#d75faf]", c("#d75fd7"): "[#d75fd7]", c("#d75fff"): "[#d75fff]",
			c("#d78700"): "[#d78700]", c("#d7875f"): "[#d7875f]", c("#d78787"): "[#d78787]", c("#d787af"): "[#d787af]",
			c("#d787d7"): "[#d787d7]", c("#d787ff"): "[#d787ff]", c("#d7af00"): "[#d7af00]", c("#d7af5f"): "[#d7af5f]",
			c("#d7af87"): "[#d7af87]", c("#d7afaf"): "[#d7afaf]", c("#d7afd7"): "[#d7afd7]", c("#d7afff"): "[#d7afff]",
			c("#d7d700"): "[#d7d700]", c("#d7d75f"): "[#d7d75f]", c("#d7d787"): "[#d7d787]", c("#d7d7af"): "[#d7d7af]",
			c("#d7d7d7"): "[#d7d7d7]", c("#d7d7ff"): "[#d7d7ff]", c("#d7ff00"): "[#d7ff00]", c("#d7ff5f"): "[#d7ff5f]",
			c("#d7ff87"): "[#d7ff87]", c("#d7ffaf"): "[#d7ffaf]", c("#d7ffd7"): "[#d7ffd7]", c("#d7ffff"): "[#d7ffff]",
			c("#ff0000"): "[#ff0000]", c("#ff005f"): "[#ff005f]", c("#ff0087"): "[#ff0087]", c("#ff00af"): "[#ff00af]",
			c("#ff00d7"): "[#ff00d7]", c("#ff00ff"): "[#ff00ff]", c("#ff5f00"): "[#ff5f00]", c("#ff5f5f"): "[#ff5f5f]",
			c("#ff5f87"): "[#ff5f87]", c("#ff5faf"): "[#ff5faf]", c("#ff5fd7"): "[#ff5fd7]", c("#ff5fff"): "[#ff5fff]",
			c("#ff8700"): "[#ff8700]", c("#ff875f"): "[#ff875f]", c("#ff8787"): "[#ff8787]", c("#ff87af"): "[#ff87af]",
			c("#ff87d7"): "[#ff87d7]", c("#ff87ff"): "[#ff87ff]", c("#ffaf00"): "[#ffaf00]", c("#ffaf5f"): "[#ffaf5f]",
			c("#ffaf87"): "[#ffaf87]", c("#ffafaf"): "[#ffafaf]", c("#ffafd7"): "[#ffafd7]", c("#ffafff"): "[#ffafff]",
			c("#ffd700"): "[#ffd700]", c("#ffd75f"): "[#ffd75f]", c("#ffd787"): "[#ffd787]", c("#ffd7af"): "[#ffd7af]",
			c("#ffd7d7"): "[#ffd7d7]", c("#ffd7ff"): "[#ffd7ff]", c("#ffff00"): "[#ffff00]", c("#ffff5f"): "[#ffff5f]",
			c("#ffff87"): "[#ffff87]", c("#ffffaf"): "[#ffffaf]", c("#ffffd7"): "[#ffffd7]", c("#ffffff"): "[#ffffff]",
			c("#080808"): "[#080808]", c("#121212"): "[#121212]", c("#1c1c1c"): "[#1c1c1c]", c("#262626"): "[#262626]",
			c("#303030"): "[#303030]", c("#3a3a3a"): "[#3a3a3a]", c("#444444"): "[#444444]", c("#4e4e4e"): "[#4e4e4e]",
			c("#585858"): "[#585858]", c("#626262"): "[#626262]", c("#6c6c6c"): "[#6c6c6c]", c("#767676"): "[#767676]",
			c("#808080"): "[#808080]", c("#8a8a8a"): "[#8a8a8a]", c("#949494"): "[#949494]", c("#9e9e9e"): "[#9e9e9e]",
			c("#a8a8a8"): "[#a8a8a8]", c("#b2b2b2"): "[#b2b2b2]", c("#bcbcbc"): "[#bcbcbc]", c("#c6c6c6"): "[#c6c6c6]",
			c("#d0d0d0"): "[#d0d0d0]", c("#dadada"): "[#dadada]", c("#e4e4e4"): "[#e4e4e4]", c("#eeeeee"): "[#eeeeee]",
		},
		background: map[chroma.Colour]string{
			c("#000000"): "[#000000]", c("#800000"): "[#800000]", c("#008000"): "[#008000]", c("#808000"): "[#808000]",
			c("#000080"): "[#000080]", c("#800080"): "[#800080]", c("#008080"): "[#008080]", c("#c0c0c0"): "[#c0c0c0]",
			c("#808080"): "[#808080]", c("#ff0000"): "[#ff0000]", c("#00ff00"): "[#00ff00]", c("#ffff00"): "[#ffff00]",
			c("#0000ff"): "[#0000ff]", c("#ff00ff"): "[#ff00ff]", c("#00ffff"): "[#00ffff]", c("#ffffff"): "[#ffffff]",
			c("#000000"): "[#000000]", c("#00005f"): "[#00005f]", c("#000087"): "[#000087]", c("#0000af"): "[#0000af]",
			c("#0000d7"): "[#0000d7]", c("#0000ff"): "[#0000ff]", c("#005f00"): "[#005f00]", c("#005f5f"): "[#005f5f]",
			c("#005f87"): "[#005f87]", c("#005faf"): "[#005faf]", c("#005fd7"): "[#005fd7]", c("#005fff"): "[#005fff]",
			c("#008700"): "[#008700]", c("#00875f"): "[#00875f]", c("#008787"): "[#008787]", c("#0087af"): "[#0087af]",
			c("#0087d7"): "[#0087d7]", c("#0087ff"): "[#0087ff]", c("#00af00"): "[#00af00]", c("#00af5f"): "[#00af5f]",
			c("#00af87"): "[#00af87]", c("#00afaf"): "[#00afaf]", c("#00afd7"): "[#00afd7]", c("#00afff"): "[#00afff]",
			c("#00d700"): "[#00d700]", c("#00d75f"): "[#00d75f]", c("#00d787"): "[#00d787]", c("#00d7af"): "[#00d7af]",
			c("#00d7d7"): "[#00d7d7]", c("#00d7ff"): "[#00d7ff]", c("#00ff00"): "[#00ff00]", c("#00ff5f"): "[#00ff5f]",
			c("#00ff87"): "[#00ff87]", c("#00ffaf"): "[#00ffaf]", c("#00ffd7"): "[#00ffd7]", c("#00ffff"): "[#00ffff]",
			c("#5f0000"): "[#5f0000]", c("#5f005f"): "[#5f005f]", c("#5f0087"): "[#5f0087]", c("#5f00af"): "[#5f00af]",
			c("#5f00d7"): "[#5f00d7]", c("#5f00ff"): "[#5f00ff]", c("#5f5f00"): "[#5f5f00]", c("#5f5f5f"): "[#5f5f5f]",
			c("#5f5f87"): "[#5f5f87]", c("#5f5faf"): "[#5f5faf]", c("#5f5fd7"): "[#5f5fd7]", c("#5f5fff"): "[#5f5fff]",
			c("#5f8700"): "[#5f8700]", c("#5f875f"): "[#5f875f]", c("#5f8787"): "[#5f8787]", c("#5f87af"): "[#5f87af]",
			c("#5f87d7"): "[#5f87d7]", c("#5f87ff"): "[#5f87ff]", c("#5faf00"): "[#5faf00]", c("#5faf5f"): "[#5faf5f]",
			c("#5faf87"): "[#5faf87]", c("#5fafaf"): "[#5fafaf]", c("#5fafd7"): "[#5fafd7]", c("#5fafff"): "[#5fafff]",
			c("#5fd700"): "[#5fd700]", c("#5fd75f"): "[#5fd75f]", c("#5fd787"): "[#5fd787]", c("#5fd7af"): "[#5fd7af]",
			c("#5fd7d7"): "[#5fd7d7]", c("#5fd7ff"): "[#5fd7ff]", c("#5fff00"): "[#5fff00]", c("#5fff5f"): "[#5fff5f]",
			c("#5fff87"): "[#5fff87]", c("#5fffaf"): "[#5fffaf]", c("#5fffd7"): "[#5fffd7]", c("#5fffff"): "[#5fffff]",
			c("#870000"): "[#870000]", c("#87005f"): "[#87005f]", c("#870087"): "[#870087]", c("#8700af"): "[#8700af]",
			c("#8700d7"): "[#8700d7]", c("#8700ff"): "[#8700ff]", c("#875f00"): "[#875f00]", c("#875f5f"): "[#875f5f]",
			c("#875f87"): "[#875f87]", c("#875faf"): "[#875faf]", c("#875fd7"): "[#875fd7]", c("#875fff"): "[#875fff]",
			c("#878700"): "[#878700]", c("#87875f"): "[#87875f]", c("#878787"): "[#878787]", c("#8787af"): "[#8787af]",
			c("#8787d7"): "[#8787d7]", c("#8787ff"): "[#8787ff]", c("#87af00"): "[#87af00]", c("#87af5f"): "[#87af5f]",
			c("#87af87"): "[#87af87]", c("#87afaf"): "[#87afaf]", c("#87afd7"): "[#87afd7]", c("#87afff"): "[#87afff]",
			c("#87d700"): "[#87d700]", c("#87d75f"): "[#87d75f]", c("#87d787"): "[#87d787]", c("#87d7af"): "[#87d7af]",
			c("#87d7d7"): "[#87d7d7]", c("#87d7ff"): "[#87d7ff]", c("#87ff00"): "[#87ff00]", c("#87ff5f"): "[#87ff5f]",
			c("#87ff87"): "[#87ff87]", c("#87ffaf"): "[#87ffaf]", c("#87ffd7"): "[#87ffd7]", c("#87ffff"): "[#87ffff]",
			c("#af0000"): "[#af0000]", c("#af005f"): "[#af005f]", c("#af0087"): "[#af0087]", c("#af00af"): "[#af00af]",
			c("#af00d7"): "[#af00d7]", c("#af00ff"): "[#af00ff]", c("#af5f00"): "[#af5f00]", c("#af5f5f"): "[#af5f5f]",
			c("#af5f87"): "[#af5f87]", c("#af5faf"): "[#af5faf]", c("#af5fd7"): "[#af5fd7]", c("#af5fff"): "[#af5fff]",
			c("#af8700"): "[#af8700]", c("#af875f"): "[#af875f]", c("#af8787"): "[#af8787]", c("#af87af"): "[#af87af]",
			c("#af87d7"): "[#af87d7]", c("#af87ff"): "[#af87ff]", c("#afaf00"): "[#afaf00]", c("#afaf5f"): "[#afaf5f]",
			c("#afaf87"): "[#afaf87]", c("#afafaf"): "[#afafaf]", c("#afafd7"): "[#afafd7]", c("#afafff"): "[#afafff]",
			c("#afd700"): "[#afd700]", c("#afd75f"): "[#afd75f]", c("#afd787"): "[#afd787]", c("#afd7af"): "[#afd7af]",
			c("#afd7d7"): "[#afd7d7]", c("#afd7ff"): "[#afd7ff]", c("#afff00"): "[#afff00]", c("#afff5f"): "[#afff5f]",
			c("#afff87"): "[#afff87]", c("#afffaf"): "[#afffaf]", c("#afffd7"): "[#afffd7]", c("#afffff"): "[#afffff]",
			c("#d70000"): "[#d70000]", c("#d7005f"): "[#d7005f]", c("#d70087"): "[#d70087]", c("#d700af"): "[#d700af]",
			c("#d700d7"): "[#d700d7]", c("#d700ff"): "[#d700ff]", c("#d75f00"): "[#d75f00]", c("#d75f5f"): "[#d75f5f]",
			c("#d75f87"): "[#d75f87]", c("#d75faf"): "[#d75faf]", c("#d75fd7"): "[#d75fd7]", c("#d75fff"): "[#d75fff]",
			c("#d78700"): "[#d78700]", c("#d7875f"): "[#d7875f]", c("#d78787"): "[#d78787]", c("#d787af"): "[#d787af]",
			c("#d787d7"): "[#d787d7]", c("#d787ff"): "[#d787ff]", c("#d7af00"): "[#d7af00]", c("#d7af5f"): "[#d7af5f]",
			c("#d7af87"): "[#d7af87]", c("#d7afaf"): "[#d7afaf]", c("#d7afd7"): "[#d7afd7]", c("#d7afff"): "[#d7afff]",
			c("#d7d700"): "[#d7d700]", c("#d7d75f"): "[#d7d75f]", c("#d7d787"): "[#d7d787]", c("#d7d7af"): "[#d7d7af]",
			c("#d7d7d7"): "[#d7d7d7]", c("#d7d7ff"): "[#d7d7ff]", c("#d7ff00"): "[#d7ff00]", c("#d7ff5f"): "[#d7ff5f]",
			c("#d7ff87"): "[#d7ff87]", c("#d7ffaf"): "[#d7ffaf]", c("#d7ffd7"): "[#d7ffd7]", c("#d7ffff"): "[#d7ffff]",
			c("#ff0000"): "[#ff0000]", c("#ff005f"): "[#ff005f]", c("#ff0087"): "[#ff0087]", c("#ff00af"): "[#ff00af]",
			c("#ff00d7"): "[#ff00d7]", c("#ff00ff"): "[#ff00ff]", c("#ff5f00"): "[#ff5f00]", c("#ff5f5f"): "[#ff5f5f]",
			c("#ff5f87"): "[#ff5f87]", c("#ff5faf"): "[#ff5faf]", c("#ff5fd7"): "[#ff5fd7]", c("#ff5fff"): "[#ff5fff]",
			c("#ff8700"): "[#ff8700]", c("#ff875f"): "[#ff875f]", c("#ff8787"): "[#ff8787]", c("#ff87af"): "[#ff87af]",
			c("#ff87d7"): "[#ff87d7]", c("#ff87ff"): "[#ff87ff]", c("#ffaf00"): "[#ffaf00]", c("#ffaf5f"): "[#ffaf5f]",
			c("#ffaf87"): "[#ffaf87]", c("#ffafaf"): "[#ffafaf]", c("#ffafd7"): "[#ffafd7]", c("#ffafff"): "[#ffafff]",
			c("#ffd700"): "[#ffd700]", c("#ffd75f"): "[#ffd75f]", c("#ffd787"): "[#ffd787]", c("#ffd7af"): "[#ffd7af]",
			c("#ffd7d7"): "[#ffd7d7]", c("#ffd7ff"): "[#ffd7ff]", c("#ffff00"): "[#ffff00]", c("#ffff5f"): "[#ffff5f]",
			c("#ffff87"): "[#ffff87]", c("#ffffaf"): "[#ffffaf]", c("#ffffd7"): "[#ffffd7]", c("#ffffff"): "[#ffffff]",
			c("#080808"): "[#080808]", c("#121212"): "[#121212]", c("#1c1c1c"): "[#1c1c1c]", c("#262626"): "[#262626]",
			c("#303030"): "[#303030]", c("#3a3a3a"): "[#3a3a3a]", c("#444444"): "[#444444]", c("#4e4e4e"): "[#4e4e4e]",
			c("#585858"): "[#585858]", c("#626262"): "[#626262]", c("#6c6c6c"): "[#6c6c6c]", c("#767676"): "[#767676]",
			c("#808080"): "[#808080]", c("#8a8a8a"): "[#8a8a8a]", c("#949494"): "[#949494]", c("#9e9e9e"): "[#9e9e9e]",
			c("#a8a8a8"): "[#a8a8a8]", c("#b2b2b2"): "[#b2b2b2]", c("#bcbcbc"): "[#bcbcbc]", c("#c6c6c6"): "[#c6c6c6]",
			c("#d0d0d0"): "[#d0d0d0]", c("#dadada"): "[#dadada]", c("#e4e4e4"): "[#e4e4e4]", c("#eeeeee"): "[#eeeeee]",
		},
	},
}

func entryToEscapeSequence(table *ttyTable, entry chroma.StyleEntry) string {
	var out string
	if entry.Colour.IsSet() {
		out += table.foreground[findClosest(table, entry.Colour)]
	}
	return out
}

func findClosest(table *ttyTable, seeking chroma.Colour) chroma.Colour {
	closestColour := chroma.Colour(0)
	closest := float64(math.MaxFloat64)
	for colour := range table.foreground {
		distance := colour.Distance(seeking)
		if distance < closest {
			closest = distance
			closestColour = colour
		}
	}
	return closestColour
}

func styleToEscapeSequence(table *ttyTable, style *chroma.Style) map[chroma.TokenType]string {
	out := map[chroma.TokenType]string{}
	for _, ttype := range style.Types() {
		entry := style.Get(ttype)
		out[ttype] = entryToEscapeSequence(table, entry)
	}
	return out
}

type indexedTTYFormatter struct {
	table *ttyTable
}

func (c *indexedTTYFormatter) Format(w io.Writer, style *chroma.Style, it chroma.Iterator) (err error) {
	defer func() {
		if perr := recover(); perr != nil {
			err = perr.(error)
		}
	}()
	theme := styleToEscapeSequence(c.table, style)
	for token := it(); token != chroma.EOF; token = it() {
		clr, ok := theme[token.Type]
		if !ok {
			clr, ok = theme[token.Type.SubCategory()]
			if !ok {
				clr = theme[token.Type.Category()]
			}
		}
		if clr != "" {
			fmt.Fprint(w, clr)
		}
		fmt.Fprint(w, token.Value)
	}
	return nil
}

func init() {
	formatters.Register("tview-8bit", &indexedTTYFormatter{ttyTables[8]})
	formatters.Register("tview-256bit", &indexedTTYFormatter{ttyTables[256]})
}
