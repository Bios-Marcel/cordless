package discordutil

import (
	"testing"

	"github.com/Bios-Marcel/discordgo"
)

func Test_MentionsCurrentUserExplicitly(t *testing.T) {
	state := &discordgo.State{
		Ready: discordgo.Ready{
			User: &discordgo.User{
				ID: "123",
			},
		},
	}

	tests := []struct {
		name    string
		message *discordgo.Message
		want    bool
	}{
		{
			name:    "no mentions",
			message: &discordgo.Message{},
			want:    false,
		}, {
			name: "one foreign mention",
			message: &discordgo.Message{
				Mentions: []*discordgo.User{
					{ID: "1"},
				},
			},
			want: false,
		}, {
			name: "multiple foreign mention",
			message: &discordgo.Message{
				Mentions: []*discordgo.User{
					{ID: "1"},
					{ID: "2"},
					{ID: "3"},
					{ID: "4"},
				},
			},
			want: false,
		}, {
			name: "one mention for current user",
			message: &discordgo.Message{
				Mentions: []*discordgo.User{
					{ID: "123"},
				},
			},
			want: true,
		}, {
			name: "multiple mentions containing one for current user in the middle",
			message: &discordgo.Message{
				Mentions: []*discordgo.User{
					{ID: "1"},
					{ID: "2"},
					{ID: "123"},
					{ID: "4"},
					{ID: "5"},
				},
			},
			want: true,
		}, {
			name: "multiple mentions containing one for current user in the end",
			message: &discordgo.Message{
				Mentions: []*discordgo.User{
					{ID: "1"},
					{ID: "2"},
					{ID: "3"},
					{ID: "4"},
					{ID: "123"},
				},
			},
			want: true,
		}, {
			name: "multiple mentions containing one for current user in the beginning",
			message: &discordgo.Message{
				Mentions: []*discordgo.User{
					{ID: "123"},
					{ID: "2"},
					{ID: "3"},
					{ID: "4"},
					{ID: "5"},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MentionsCurrentUserExplicitly(state, tt.message); got != tt.want {
				t.Errorf("MentionsCurrentUserExplicitly() = %v, want %v", got, tt.want)
			}
		})
	}
}

type messageSupplier struct {
	requestAmount int
}

func (l *messageSupplier) ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
	l.requestAmount++
	return nil, nil
}

func Test_LoadMessages_CacheAccess(t *testing.T) {
	t.Run("Test empty channel", func(t *testing.T) {
		channelEmpty := &discordgo.Channel{
			ID:            "1",
			LastMessageID: "",
		}
		supplier := &messageSupplier{}
		loader := CreateMessageLoader(supplier)

		_, loadError := loader.LoadMessages(channelEmpty)
		if loadError != nil {
			t.Errorf("Error loading channels: %s", loadError)
		}

		if supplier.requestAmount != 0 {
			t.Errorf("Cache was accessed %d times instead of 1 time.", supplier.requestAmount)
		}

		_, loadError = loader.LoadMessages(channelEmpty)
		if loadError != nil {
			t.Errorf("Error loading channels: %s", loadError)
		}

		if supplier.requestAmount != 0 {
			t.Errorf("Cache was accessed %d times instead of 1 time.", supplier.requestAmount)
		}
	})

	t.Run("Test cache access", func(t *testing.T) {
		channelNotEmpty := &discordgo.Channel{
			ID:            "1",
			LastMessageID: "123",
		}
		supplier := &messageSupplier{}
		loader := CreateMessageLoader(supplier)

		_, loadError := loader.LoadMessages(channelNotEmpty)
		if loadError != nil {
			t.Errorf("Error loading channels: %s", loadError)
		}
		_, loadError = loader.LoadMessages(channelNotEmpty)
		if loadError != nil {
			t.Errorf("Error loading channels: %s", loadError)
		}

		if supplier.requestAmount != 1 {
			t.Errorf("Cache was accessed %d times instead of 1 time.", supplier.requestAmount)
		}
	})

	t.Run("Test cache access with cacheclear", func(t *testing.T) {
		channelNotEmpty := &discordgo.Channel{
			ID:            "1",
			LastMessageID: "123",
		}
		supplier := &messageSupplier{}
		loader := CreateMessageLoader(supplier)

		if loader.IsCached(channelNotEmpty.ID) {
			t.Error("Channel cache should still be clear.")
		}

		_, loadError := loader.LoadMessages(channelNotEmpty)
		if loadError != nil {
			t.Errorf("Error loading channels: %s", loadError)
		}
		_, loadError = loader.LoadMessages(channelNotEmpty)
		if loadError != nil {
			t.Errorf("Error loading channels: %s", loadError)
		}

		if supplier.requestAmount != 1 {
			t.Errorf("Cache was accessed %d times instead of 1 time.", supplier.requestAmount)
		}

		if !loader.IsCached(channelNotEmpty.ID) {
			t.Error("Channel should've been cached but wasn't.")
		}

		loader.DeleteFromCache(channelNotEmpty.ID)

		if loader.IsCached(channelNotEmpty.ID) {
			t.Error("Cache should've been cleared, but wasn't.")
		}

	})
}

func TestGenerateQuote(t *testing.T) {
	type args struct {
		message           string
		author            string
		time              discordgo.Timestamp
		attachments       []*discordgo.MessageAttachment
		messageAfterQuote string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple line",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "",
			},
			want:    "> **humaN** 21:30:57 UTC:\n> Hello World\n",
			wantErr: false,
		}, {
			name: "simple line; non UTC - positive",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+03:00"),
				messageAfterQuote: "",
			},
			want:    "> **humaN** 18:30:57 UTC:\n> Hello World\n",
			wantErr: false,
		}, {
			name: "simple line; non UTC - negative",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000-02:00"),
				messageAfterQuote: "",
			},
			want:    "> **humaN** 23:30:57 UTC:\n> Hello World\n",
			wantErr: false,
		}, {
			name: "multi line",
			args: args{
				message:           "Hello World\nBye World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "",
			},
			want:    "> **humaN** 21:30:57 UTC:\n> Hello World\n> Bye World\n",
			wantErr: false,
		}, {
			name: "simple line with message after quote",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "Hei",
			},
			want:    "> **humaN** 21:30:57 UTC:\n> Hello World\nHei",
			wantErr: false,
		}, {
			name: "simple line with multline message after quote",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "Hei\nHo",
			},
			want:    "> **humaN** 21:30:57 UTC:\n> Hello World\nHei\nHo",
			wantErr: false,
		}, {
			name: "simple line with whitespace message after quote",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "    \t    ",
			},
			want:    "> **humaN** 21:30:57 UTC:\n> Hello World\n",
			wantErr: false,
		}, {
			name: "simple line with surrounding whitespace message after quote",
			args: args{
				message:           "Hello World",
				author:            "humaN",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "    \t    hei",
			},
			want:    "> **humaN** 21:30:57 UTC:\n> Hello World\nhei",
			wantErr: false,
		}, {
			name: "empty author; we won't handle this, but still specify expected behaviour",
			args: args{
				message:           "Hello World\nBye World",
				author:            "",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "",
			},
			want:    "> **** 21:30:57 UTC:\n> Hello World\n> Bye World\n",
			wantErr: false,
		}, {
			name: "empty message; we won't handle this, but still specify expected behaviour",
			args: args{
				message:           "",
				author:            "author",
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "",
			},
			want:    "> **author** 21:30:57 UTC:\n> \n",
			wantErr: false,
		}, {
			name: "Invalid timestamps should cause an error",
			args: args{
				message:           "",
				author:            "",
				time:              discordgo.Timestamp("OwO, an invalid timestamp"),
				messageAfterQuote: "",
			},
			want:    "",
			wantErr: true,
		}, {
			name: "single line plus single attachment, no sender message",
			args: args{
				message: "f",
				author:  "author",
				attachments: []*discordgo.MessageAttachment{
					{
						URL: "https://download/this.zip",
					},
				},
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "",
			},
			want: "> **author** 21:30:57 UTC:\n" +
				"> f\n" +
				"> https://download/this.zip\n",
			wantErr: false,
		}, {
			name: "single line plus two attachment, no sender message",
			args: args{
				message: "f",
				author:  "author",
				attachments: []*discordgo.MessageAttachment{
					{
						URL: "https://download/this.zip",
					},
					{
						URL: "https://download/thistoo.zip",
					},
				},
				time:              discordgo.Timestamp("2019-10-28T21:30:57.003000+00:00"),
				messageAfterQuote: "",
			},
			want: "> **author** 21:30:57 UTC:\n" +
				"> f\n" +
				"> https://download/this.zip\n" +
				"> https://download/thistoo.zip\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateQuote(tt.args.message, tt.args.author, tt.args.time, tt.args.attachments, tt.args.messageAfterQuote)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateQuote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateQuote() got = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
