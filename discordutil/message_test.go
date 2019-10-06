package discordutil

import (
	"github.com/Bios-Marcel/discordgo"
	"testing"
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
