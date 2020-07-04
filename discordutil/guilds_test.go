package discordutil

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/Bios-Marcel/discordgo"
)

func TestSortGuilds(t *testing.T) {
	//The IDs have random length and don't mean much.
	guildOneID := "98122541287"
	guildTwoID := "12450501965"
	guildThreeID := "1086518z963"
	guildFourID := "19651241842"
	settings := &discordgo.Settings{
		GuildPositions: []string{
			guildOneID,
			guildTwoID,
			guildThreeID,
			guildFourID,
		},
	}

	guilds := []*discordgo.Guild{
		{ID: guildThreeID},
		{ID: guildOneID},
		{ID: guildFourID},
		{ID: guildTwoID},
	}

	SortGuilds(settings, guilds)

	if guilds[0].ID != guildOneID {
		t.Errorf("The first guild should've been %s, but was %s", guildOneID, guilds[0].ID)
	}

	if guilds[1].ID != guildTwoID {
		t.Errorf("The second guild should've been %s, but was %s", guildTwoID, guilds[1].ID)
	}

	if guilds[2].ID != guildThreeID {
		t.Errorf("The third guild should've been %s, but was %s", guildThreeID, guilds[2].ID)
	}

	if guilds[3].ID != guildFourID {
		t.Errorf("The fourth guild should've been %s, but was %s", guildFourID, guilds[3].ID)
	}
}

type testGuildLoader struct {
	loadFunction func(int, string, string) ([]*discordgo.UserGuild, error)
}

func (loader testGuildLoader) UserGuilds(amount int, beforeID, afterID string) ([]*discordgo.UserGuild, error) {
	return loader.loadFunction(amount, beforeID, afterID)
}

func generateGuilds(start, amount int) []*discordgo.UserGuild {
	guilds := make([]*discordgo.UserGuild, 0, amount)
	for i := start; i < start+amount; i++ {
		fmt.Println(i)
		guilds = append(guilds, &discordgo.UserGuild{ID: fmt.Sprintf("%d", i)})
	}

	return guilds
}

func TestLoadGuilds(t *testing.T) {
	tests := []struct {
		name        string
		guildLoader GuildLoader
		want        []*discordgo.UserGuild
		wantErr     bool
	}{
		{
			name: "forward error",
			guildLoader: testGuildLoader{func(amount int, beforeID, afterID string) ([]*discordgo.UserGuild, error) {
				return nil, errors.New("owo, an error")
			}},
			want:    nil,
			wantErr: true,
		}, {
			name: "no guilds",
			guildLoader: testGuildLoader{func(amount int, beforeID, afterID string) ([]*discordgo.UserGuild, error) {
				return nil, nil
			}},
			want:    []*discordgo.UserGuild{},
			wantErr: false,
		}, {
			name: "100 guilds",
			guildLoader: testGuildLoader{func(amount int, beforeID, afterID string) ([]*discordgo.UserGuild, error) {
				//100 is the API limit
				if amount == 100 {
					if beforeID == "" {
						return generateGuilds(1, 100), nil
					} else if beforeID == "1" {
						return []*discordgo.UserGuild{}, nil
					}

					return nil, errors.New("unsupported case")
				}

				return nil, errors.New("test only supports usecase of 100 at once")
			}},
			want:    generateGuilds(1, 100),
			wantErr: false,
		}, {
			name: "150 guilds",
			guildLoader: testGuildLoader{func(amount int, beforeID, afterID string) ([]*discordgo.UserGuild, error) {
				//100 is the API limit
				if amount == 100 {
					if beforeID == "" {
						return generateGuilds(51, 100), nil
					} else if beforeID == "51" {
						return generateGuilds(1, 50), nil
					}

					return nil, errors.New("unsupported case")
				}

				return nil, errors.New("test only supports usecase of 100 at once")
			}},
			want:    generateGuilds(1, 150),
			wantErr: false,
		}, {
			name: "200 guilds",
			guildLoader: testGuildLoader{func(amount int, beforeID, afterID string) ([]*discordgo.UserGuild, error) {
				//100 is the API limit
				if amount == 100 {
					if beforeID == "" {
						return generateGuilds(101, 100), nil
					} else if beforeID == "101" {
						return generateGuilds(1, 100), nil
					} else if beforeID == "1" {
						return []*discordgo.UserGuild{}, nil
					}

					return nil, errors.New("unsupported case")
				}

				return nil, errors.New("test only supports usecase of 100 at once")
			}},
			want:    generateGuilds(1, 200),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadGuilds(tt.guildLoader)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGuilds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("length of LoadGuilds() = %v, want %v", len(got), len(tt.want))
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadGuilds() = %v, want %v", got, tt.want)
			}
		})
	}
}
