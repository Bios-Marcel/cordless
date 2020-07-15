package discordutil

import (
	"fmt"

	"github.com/Bios-Marcel/discordgo"
)

// GetRoleColor returns the roles color in the format #RRGGBB
// or an empty string if there's no color.
func GetRoleColor(role *discordgo.Role) string {
	//Apparently discord doesn't like black. Black just counts as "no color", so we can skip it.
	if role.Color == 0 {
		return ""
	}

	r := role.Color >> 16 & 255
	g := role.Color >> 8 & 255
	b := role.Color & 255

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)

}
