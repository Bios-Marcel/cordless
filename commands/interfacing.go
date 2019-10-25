package commands

import "github.com/Bios-Marcel/discordgo"

type ClientState interface {
	GetSelectedGuild() *discordgo.Guild
	GetSelectedChannel() *discordgo.Channel
}
