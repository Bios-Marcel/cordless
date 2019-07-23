package discordutil

import (
	"testing"

	"github.com/Bios-Marcel/discordgo"
)

func TestGetUserColor(t *testing.T) {
	tests := []struct {
		name        string
		user        *discordgo.User
		want        string
		doNotEquals bool
	}{
		{
			name: "Verify that same ID gets the same color.",
			user: &discordgo.User{
				ID: "1398541219874",
			},
			want: GetUserColor(&discordgo.User{
				ID: "1398541219874",
			}),
		},
		{
			name: "Verify that same ID gets the same color.",
			user: &discordgo.User{
				ID: "1398541219874",
			},
			want:        "",
			doNotEquals: true,
		},
		{
			name: "Verify that bot gets a specific color, no matter which ID.",
			user: &discordgo.User{
				ID:  "1398541219874",
				Bot: true,
			},
			want: botColor,
		},
		{
			name: "Verify that a new ID doesn't get the same color as a specific previous ID.",
			user: &discordgo.User{
				ID: "0183587135982",
			},
			want: GetUserColor(&discordgo.User{
				ID: "1398541219874",
			}),
			doNotEquals: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserColor(tt.user); (tt.doNotEquals && got == tt.want) || (!tt.doNotEquals && got != tt.want) {
				t.Errorf("GetUserColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMemberName(t *testing.T) {
	tests := []struct {
		name   string
		member *discordgo.Member
		want   string
	}{
		{
			name: "nickname and no bot",
			member: &discordgo.Member{
				Nick: "Hello",
				User: &discordgo.User{
					Username: "World",
				},
			},
			want: "Hello",
		},
		{
			name: "nickname and bot",
			member: &discordgo.Member{
				Nick: "Hello",
				User: &discordgo.User{
					Username: "World",
					Bot:      true,
				},
			},
			want: botPrefix + "Hello",
		},
		{
			name: "no nickname and bot",
			member: &discordgo.Member{
				User: &discordgo.User{
					Username: "World",
					Bot:      true,
				},
			},
			want: botPrefix + "World",
		},
		{
			name: "no nickname and no bot",
			member: &discordgo.Member{
				User: &discordgo.User{
					Username: "World",
				},
			},
			want: "World",
		},
		{
			name: "no nickname and no bot and contains tview color sequenec",
			member: &discordgo.Member{
				User: &discordgo.User{
					Username: "[red]World",
				},
			},
			//Might break if the way tview escapes breaks
			want: "[red[]World",
		},
		{
			name: "no nickname and no bot and contains tview region",
			member: &discordgo.Member{
				User: &discordgo.User{
					Username: "[\"hello\"]World",
				},
			},
			//Might break if the way tview escapes breaks
			want: "[\"hello\"[]World",
		},
		{
			name: "no nickname and no bot and contains tview region close",
			member: &discordgo.Member{
				User: &discordgo.User{
					Username: "[\"\"]World",
				},
			},
			//Might break if the way tview escapes breaks
			want: "[\"\"[]World",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMemberName(tt.member); got != tt.want {
				t.Errorf("GetMemberName() = %v, want %v", got, tt.want)
			}
		})
	}
}
