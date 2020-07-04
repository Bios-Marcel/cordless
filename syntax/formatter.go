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
			c("#000000"): "[#000000]", c("#7f0000"): "[#7f0000]", c("#007f00"): "[#3baf3b]", c("#7f7fe0"): "[#7f7fe0]",
			c("#00007f"): "[#2d2db7]", c("#7f007f"): "[#7f007f]", c("#007f7f"): "[#3ea8a8]", c("#e5e5e5"): "[#e5e5e5]",
			c("#555555"): "[#555555]", c("#ff0000"): "[#d16666]", c("#00ff00"): "[#80dd80]", c("#ffff00"): "[#efef8b]",
			c("#0000ff"): "[#5757f2]", c("#ff00ff"): "[#d36bd3]", c("#00ffff"): "[#7ed3d3]", c("#ffffff"): "[#ffffff]",
		},
		background: map[chroma.Colour]string{
			c("#000000"): "[#000000]", c("#7f0000"): "[#7f0000]", c("#007f00"): "[#3baf3b]", c("#7f7fe0"): "[#7f7fe0]",
			c("#00007f"): "[#2d2db7]", c("#7f007f"): "[#7f007f]", c("#007f7f"): "[#3ea8a8]", c("#e5e5e5"): "[#e5e5e5]",
			c("#555555"): "[#555555]", c("#ff0000"): "[#d16666]", c("#00ff00"): "[#80dd80]", c("#ffff00"): "[#efef8b]",
			c("#0000ff"): "[#5757f2]", c("#ff00ff"): "[#d36bd3]", c("#00ffff"): "[#7ed3d3]", c("#ffffff"): "[#ffffff]",
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
	closest := math.MaxFloat64
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
}
