package discordutil

import (
	"testing"

	"github.com/Bios-Marcel/discordgo"
)

func TestGetPrivateChannelName(t *testing.T) {
	tests := []struct {
		name    string
		channel *discordgo.Channel
		want    string
	}{
		{
			name: "DM channel with one recipient and a name",
			channel: &discordgo.Channel{
				Name: "test",
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					},
				},
				Type: discordgo.ChannelTypeDM,
			},
			want: "Maruseru",
		}, {
			name: "DM channel with one recipient",
			channel: &discordgo.Channel{
				Type: discordgo.ChannelTypeDM,
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					},
				},
			},
			want: "Maruseru",
		}, {
			name: "DM channel with two recipient",
			channel: &discordgo.Channel{
				Type: discordgo.ChannelTypeDM,
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					}, {
						Username: "Numbah two",
					},
				},
			},
			want: "Maruseru",
		}, {
			name: "Group channel with one recipient",
			channel: &discordgo.Channel{
				Type: discordgo.ChannelTypeGroupDM,
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					},
				},
			},
			want: "Maruseru",
		}, {
			name: "Group channel with two recipients",
			channel: &discordgo.Channel{
				Type: discordgo.ChannelTypeGroupDM,
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					}, {
						Username: "Numbah two",
					},
				},
			},
			want: "Maruseru, Numbah two",
		}, {
			name: "Group channel with three recipients",
			channel: &discordgo.Channel{
				Type: discordgo.ChannelTypeGroupDM,
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					}, {
						Username: "Numbah two",
					}, {
						Username: "Numbah three",
					},
				},
			},
			want: "Maruseru, Numbah two, Numbah three",
		}, {
			name: "Group channel with three recipients and a name",
			channel: &discordgo.Channel{
				Name: "name",
				Type: discordgo.ChannelTypeGroupDM,
				Recipients: []*discordgo.User{
					{
						Username: "Maruseru",
					}, {
						Username: "Numbah two",
					}, {
						Username: "Numbah three",
					},
				},
			},
			want: "name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPrivateChannelName(tt.channel); got != tt.want {
				t.Errorf("GetPrivateChannelName() = %v, want %v", got, tt.want)
			}
		})
	}
}
