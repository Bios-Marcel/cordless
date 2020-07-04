package ui

import (
	"testing"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/config"
	_ "github.com/Bios-Marcel/cordless/syntax"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

func TestParseBoldAndUnderline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple bold",
			input: "**Hallo Welt**",
			want:  "[::b]Hallo Welt[::-]",
		},
		{
			name:  "Useless bold",
			input: "****Hallo Welt",
			want:  "****Hallo Welt",
		},
		{
			name:  "Non closed bold",
			input: "**Hallo Welt",
			want:  "**Hallo Welt",
		},
		{
			name:  "Bold newline",
			input: "**Hallo\nWelt**",
			want:  "[::b]Hallo\n[::b]Welt[::-]",
		},
		{
			name:  "Bold newline2",
			input: "Hallo**\nWelt**",
			want:  "Hallo[::b]\n[::b]Welt[::-]",
		},
		{
			name:  "Bold newline3",
			input: "Hal**lo\nWelt**",
			want:  "Hal[::b]lo\n[::b]Welt[::-]",
		},
		{
			name:  "Simple underline",
			input: "__Hallo Welt__",
			want:  "[::u]Hallo Welt[::-]",
		},
		{
			name:  "Useless underline",
			input: "____Hallo Welt",
			want:  "____Hallo Welt",
		},
		{
			name:  "Non closed underline",
			input: "__Hallo Welt",
			want:  "__Hallo Welt",
		},
		{
			name:  "Underline newline",
			input: "__Hallo\nWelt__",
			want:  "[::u]Hallo\n[::u]Welt[::-]",
		},
		{
			name:  "Underline newline2",
			input: "Hallo__\nWelt__",
			want:  "Hallo[::u]\n[::u]Welt[::-]",
		},
		{
			name:  "Underline newline3",
			input: "Hal__lo\nWelt__",
			want:  "Hal[::u]lo\n[::u]Welt[::-]",
		},
		{
			name:  "Underline and bold",
			input: "**__Hallo Welt__**",
			want:  "[::b][::bu]Hallo Welt[::b][::-]",
		},
		{
			name:  "Underline and bold2",
			input: "** OwO__Hallo Welt__**",
			want:  "[::b] OwO[::bu]Hallo Welt[::b][::-]",
		},
		{
			name:  "Underline and bold3",
			input: "** OwO__Hallo Welt__** What",
			want:  "[::b] OwO[::bu]Hallo Welt[::b][::-] What",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBoldAndUnderline(tt.input); got != tt.want {
				t.Errorf("ParseBoldAndUnderline() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestChatView_formatMessageText(t *testing.T) {
	defaultChatView := &ChatView{
		showSpoilerContent: make(map[string]bool),
		state:              &discordgo.State{},
		shortenLinks:       false,
	}
	tests := []struct {
		name     string
		input    *discordgo.Message
		want     string
		chatView *ChatView
	}{
		{
			name: "empty message",
			input: &discordgo.Message{
				Content: "",
			},
			want:     "",
			chatView: defaultChatView,
		}, {
			name: "super simple message",
			input: &discordgo.Message{
				Content: "simple",
			},
			want:     "simple",
			chatView: defaultChatView,
		}, {
			name: "super simple multiline message",
			input: &discordgo.Message{
				Content: "simple\nsimple",
			},
			want:     "simple\nsimple",
			chatView: defaultChatView,
		}, {
			name: "simple bold message",
			input: &discordgo.Message{
				Content: "**simple**",
			},
			want:     "[::b]simple[::-]",
			chatView: defaultChatView,
		}, {
			name: "simple underlined message",
			input: &discordgo.Message{
				Content: "__simple__",
			},
			want:     "[::u]simple[::-]",
			chatView: defaultChatView,
		}, {
			name: "simple message with bold part",
			input: &discordgo.Message{
				Content: "a **simple** b",
			},
			want:     "a [::b]simple[::-] b",
			chatView: defaultChatView,
		}, {
			name: "simple message with underlined part",
			input: &discordgo.Message{
				Content: "a __simple__ b",
			},
			want:     "a [::u]simple[::-] b",
			chatView: defaultChatView,
		}, {
			name: "simple message with bold and underlined part",
			input: &discordgo.Message{
				Content: "a **__simple__** b",
			},
			want:     "a [::b][::bu]simple[::b][::-] b",
			chatView: defaultChatView,
		}, {
			name: "simple message with bold and partially underlined part",
			input: &discordgo.Message{
				Content: "a **fat__simple__fat** b",
			},
			want:     "a [::b]fat[::bu]simple[::b]fat[::-] b",
			chatView: defaultChatView,
		}, {
			name: "simple message with underlined and partially bold part",
			input: &discordgo.Message{
				Content: "a __underline**fat**underline__ b",
			},
			want:     "a [::u]underline[::ub]fat[::u]underline[::-] b",
			chatView: defaultChatView,
		}, {
			name: "simple message with underlined and unclosed bold part",
			input: &discordgo.Message{
				Content: "a __underline**fatunderline__ b",
			},
			want:     "a [::u]underline**fatunderline[::-] b",
			chatView: defaultChatView,
		}, {
			name: "simple spoiler",
			input: &discordgo.Message{
				Content: "||simple||",
			},
			want:     "[" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]!SPOILER![#ffffff]",
			chatView: defaultChatView,
		}, {
			name: "simple spoiler in between",
			input: &discordgo.Message{
				Content: "gimme ||simple|| pls",
			},
			want:     "gimme [" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]!SPOILER![#ffffff] pls",
			chatView: defaultChatView,
		}, {
			name: "formatted spoiler in between",
			input: &discordgo.Message{
				Content: "gimme ||**simple**|| pls",
			},
			want:     "gimme [" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]!SPOILER![#ffffff] pls",
			chatView: defaultChatView,
		}, {
			name: "formatted spoiler in between",
			input: &discordgo.Message{
				Content: "gimme ||**simple**|| pls",
			},
			want:     "gimme [" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]!SPOILER![#ffffff] pls",
			chatView: defaultChatView,
		}, {
			name: "unclosed spoiler",
			input: &discordgo.Message{
				Content: "owo ||spoiler",
			},
			want:     "owo ||spoiler",
			chatView: defaultChatView,
		}, {
			name: "spoiler with formatting around",
			input: &discordgo.Message{
				Content: "gimme **||simple||** pls",
			},
			//FIXME Not sure whether this is correct, but it's the
			//current state, so i'll be specifying it for now.
			want:     "gimme [::b][" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]!SPOILER![#ffffff][::-] pls",
			chatView: defaultChatView,
		}, {
			name: "codeblock without specified language",
			input: &discordgo.Message{
				Content: "```\none\ntwo\nthree\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one\n[#c9dddc]▐ [#ffffff]two\n[#c9dddc]▐ [#ffffff]three",
			chatView: defaultChatView,
		}, {
			name: "one line codeblock with text around",
			input: &discordgo.Message{
				Content: "test\n```\none\n```\ntest",
			},
			want:     "test\n[#c9dddc]▐ [#ffffff]one\ntest",
			chatView: defaultChatView,
		}, {
			name: "one line codeblock with text around, but without newlines in between",
			input: &discordgo.Message{
				Content: "test```\none\n```test",
			},
			want:     "test\n[#c9dddc]▐ [#ffffff]one\ntest",
			chatView: defaultChatView,
		}, {
			name: "codeblock at start of message",
			input: &discordgo.Message{
				Content: "```\none\n```test",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one\ntest",
			chatView: defaultChatView,
		}, {
			name: "simple oneline codeblock with language",
			input: &discordgo.Message{
				Content: "```go\none\n```",
			},
			want:     "\n[#c9dddc]▐ [#efef8b]one",
			chatView: defaultChatView,
		}, {
			name: "simple oneline codeblock without language",
			input: &discordgo.Message{
				Content: "```\none\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one",
			chatView: defaultChatView,
		}, {
			name: "simple oneline codeblock with nonexistent language",
			input: &discordgo.Message{
				Content: "```owowhatsthis\none\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one",
			chatView: defaultChatView,
		}, {
			name: "simple oneline codeblock with cpp as language",
			input: &discordgo.Message{
				Content: "```cpp\none\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one[#ffffff]",
			chatView: defaultChatView,
		}, {
			name: "simple oneline codeblock with cpp as language and some unnecessary trailing newlines",
			input: &discordgo.Message{
				Content: "```cpp\none\n\n\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one[#ffffff]\n[#c9dddc]▐ [#ffffff]\n[#c9dddc]▐ [#ffffff]",
			chatView: defaultChatView,
		}, {
			name: "two simple oneline codeblocks without language after another",
			input: &discordgo.Message{
				Content: "```\none\n```\n```\none\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]one\n[#c9dddc]▐ [#ffffff]one",
			chatView: defaultChatView,
		}, {
			name: "codeblock with spoiler inside",
			input: &discordgo.Message{
				Content: "```\nowo ||Spoiler|| owo\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]owo ||Spoiler|| owo",
			chatView: defaultChatView,
		}, {
			name: "codeblock with bold text inside",
			input: &discordgo.Message{
				Content: "```\nowo **bold** owo\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]owo **bold** owo",
			chatView: defaultChatView,
		}, {
			name: "codeblock with underlined text inside",
			input: &discordgo.Message{
				Content: "```\nowo __underline__ owo\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]owo __underline__ owo",
			chatView: defaultChatView,
		}, {
			name: "codeblock with spoiler around",
			input: &discordgo.Message{
				Content: "||```\nowo\n```||",
			},
			want:     "[" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]!SPOILER![#ffffff]",
			chatView: defaultChatView,
		}, {
			name: "codeblock with revelaed spoiler around",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "||```\nowo\n```||",
			},
			want: "||\n[#c9dddc]▐ [#ffffff]owo\n||",
			chatView: &ChatView{
				showSpoilerContent: map[string]bool{
					"OwO": true,
				},
				state:        &discordgo.State{},
				shortenLinks: false,
			},
		}, {
			name: "two codeblocks without newlines but character inbeteween",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "```\nowo\n```f```\nowo\n```",
			},
			want:     "\n[#c9dddc]▐ [#ffffff]owo\nf\n[#c9dddc]▐ [#ffffff]owo",
			chatView: defaultChatView,
		}, {
			name: "Remove escape characters",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "\\`\\*\\_",
			},
			want:     "`*_",
			chatView: defaultChatView,
		}, {
			name: "Remove escape characters, but not any additional backslashes",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "\\\\`\\*\\_",
			},
			want:     "\\`*_",
			chatView: defaultChatView,
		}, {
			name: "single custom emoji",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "<:owo:123>",
			},
			want:     "[owo[]( https://cdn.discordapp.com/emojis/123 )",
			chatView: defaultChatView,
		}, {
			name: "single animated custom emoji",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "<a:owo:123>",
			},
			want:     "[a:owo[]( https://cdn.discordapp.com/emojis/123 )",
			chatView: defaultChatView,
		}, {
			//FIXME Remove space, it's useless
			name: "two custom emoji without space",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "<:owo:123><:owo:123>",
			},
			want:     "[owo[]( https://cdn.discordapp.com/emojis/123 )\n [owo[]( https://cdn.discordapp.com/emojis/123 )",
			chatView: defaultChatView,
		}, {
			//FIXME Remove space, it's useless
			name: "two custom emoji with space",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "<:owo:123> <:owo:123>",
			},
			want:     "[owo[]( https://cdn.discordapp.com/emojis/123 )\n [owo[]( https://cdn.discordapp.com/emojis/123 )",
			chatView: defaultChatView,
		}, {
			//FIXME Remove space, it's useless
			name: "multiple successive emoji without spaces",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "<:owo:123><:owo:124><:owo:125><:owo:126>",
			},
			want:     "[owo[]( https://cdn.discordapp.com/emojis/123 )\n [owo[]( https://cdn.discordapp.com/emojis/124 )\n [owo[]( https://cdn.discordapp.com/emojis/125 )\n [owo[]( https://cdn.discordapp.com/emojis/126 )",
			chatView: defaultChatView,
		}, {
			//FIXME Remove space, it's useless
			name: "multiple successive emoji with spaces",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "<:owo:123> <:owo:124> <:owo:125> <:owo:126>",
			},
			want:     "[owo[]( https://cdn.discordapp.com/emojis/123 )\n [owo[]( https://cdn.discordapp.com/emojis/124 )\n [owo[]( https://cdn.discordapp.com/emojis/125 )\n [owo[]( https://cdn.discordapp.com/emojis/126 )",
			chatView: defaultChatView,
		}, {
			//FIXME Remove spaces behind prefix and suffix of emoji
			name: "message with custom emoji",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "Look, <:owo:123> what's this?",
			},
			want:     "Look, \n[owo[]( https://cdn.discordapp.com/emojis/123 )\n what's this?",
			chatView: defaultChatView,
		}, {
			//FIXME Remove spaces behind prefix and suffix of emoji
			name: "message with multiple custom emoji with spaces",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "Look, <:owo:123> <:owo:123> what's this?",
			},
			want:     "Look, \n[owo[]( https://cdn.discordapp.com/emojis/123 )\n [owo[]( https://cdn.discordapp.com/emojis/123 )\n what's this?",
			chatView: defaultChatView,
		}, {
			//FIXME Remove spaces behind prefix and suffix of emoji
			name: "message with multiple custom emoji without spaces",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "Look, <:owo:123><:owo:123> what's this?",
			},
			want:     "Look, \n[owo[]( https://cdn.discordapp.com/emojis/123 )\n [owo[]( https://cdn.discordapp.com/emojis/123 )\n what's this?",
			chatView: defaultChatView,
		}, {
			name: "message with custom animated emoji",
			input: &discordgo.Message{
				ID:      "OwO",
				Content: "Hello <a:owo:123>",
			},
			want:     "Hello \n[a:owo[]( https://cdn.discordapp.com/emojis/123 )",
			chatView: defaultChatView,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.chatView.formatMessageText(tt.input); got != tt.want {
				t.Errorf("ChatView.formatMessageText() = '%v', want: '%v'", got, tt.want)
			}
		})
	}
}

func Test_removeLeadingWhitespaceInCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "no whitespace; single line",
			code: "1",
			want: "1",
		}, {
			name: "no whitespace; two lines",
			code: "1\n2",
			want: "1\n2",
		}, {
			name: "single space on single line",
			code: " 1",
			want: "1",
		}, {
			name: "two spaces on single line",
			code: "  1",
			want: "1",
		}, {
			name: "multiple spaces on single line",
			code: "    1",
			want: "1",
		}, {
			name: "multiple spaces on multiple lines with different amounts",
			code: "    1\n  2\n 3",
			want: "   1\n 2\n3",
		}, {
			name: "single tab on single line",
			code: "	1",
			want: "1",
		}, {
			name: "single tab on multiple lines",
			code: "	1\n	2",
			want: "1\n2",
		}, {
			name: "multiple tabs on multiple lines",
			code: "		1\n		2",
			want: "1\n2",
		}, {
			name: "multiple tabs on multiple lines; each line different tab amount",
			code: "				1\n		2",
			want: "		1\n2",
		}, {
			name: "only one line with a tab",
			code: "	1\n2",
			want: "	1\n2",
		}, {
			name: "multiple tabs on single line",
			code: "			1",
			want: "1",
		}, {
			name: "mixed tabs and spaces of multiple lines",
			code: " 	1\n  	2\n 	3",
			want: "	1\n 	2\n	3",
		}, {
			name: "multiple lines with an empty line in between",
			code: "	1\n\n	2\n	3",
			want: "1\n\n2\n3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeLeadingWhitespaceInCode(tt.code); got != tt.want {
				t.Errorf("removeLeadingWhitespaceInCode() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
