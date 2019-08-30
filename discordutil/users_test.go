package discordutil

import (
	"reflect"
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

func TestSortUserRoles(t *testing.T) {
	type args struct {
		roles      []string
		guildRoles []*discordgo.Role
	}
	tests := []struct {
		name     string
		args     args
		expected []string
	}{
		{
			name: "empty",
			args: args{
				roles:      []string{},
				guildRoles: []*discordgo.Role{},
			},
			expected: []string{},
		}, {
			name: "one role without discord roles in the state",
			args: args{
				roles:      []string{"lol"},
				guildRoles: []*discordgo.Role{},
			},
			expected: []string{"lol"},
		}, {
			name: "two roles without discord roles in the state",
			args: args{
				roles:      []string{"a", "b"},
				guildRoles: []*discordgo.Role{},
			},
			expected: []string{"a", "b"},
		}, {
			name: "one role that's in the state",
			args: args{
				roles: []string{"a"},
				guildRoles: []*discordgo.Role{
					&discordgo.Role{ID: "a"},
				},
			},
			expected: []string{"a"},
		}, {
			name: "two roles that's in the state",
			args: args{
				roles: []string{"a", "b"},
				guildRoles: []*discordgo.Role{
					&discordgo.Role{
						ID:       "a",
						Position: 0,
					},
					&discordgo.Role{
						ID:       "b",
						Position: 1,
					},
				},
			},
			expected: []string{"b", "a"},
		}, {
			name: "three roles that's in the state",
			args: args{
				roles: []string{"a", "b", "c"},
				guildRoles: []*discordgo.Role{
					&discordgo.Role{
						ID:       "a",
						Position: 0,
					},
					&discordgo.Role{
						ID:       "b",
						Position: 1,
					},
					&discordgo.Role{
						ID:       "c",
						Position: 2,
					},
				},
			},
			expected: []string{"c", "b", "a"},
		}, {
			name: "three roles that's in the state; already correct",
			args: args{
				roles: []string{"a", "b", "c"},
				guildRoles: []*discordgo.Role{
					&discordgo.Role{
						ID:       "a",
						Position: 4,
					},
					&discordgo.Role{
						ID:       "b",
						Position: 2,
					},
					&discordgo.Role{
						ID:       "c",
						Position: 0,
					},
				},
			},
			expected: []string{"a", "b", "c"},
		}, {
			name: "roles with same positions should be preserved in order",
			args: args{
				roles: []string{"a", "b", "c", "d"},
				guildRoles: []*discordgo.Role{
					&discordgo.Role{
						ID:       "a",
						Position: 0,
					},
					&discordgo.Role{
						ID:       "b",
						Position: 2,
					},
					&discordgo.Role{
						ID:       "c",
						Position: 2,
					},
					&discordgo.Role{
						ID:       "d",
						Position: 4,
					},
				},
			},
			expected: []string{"d", "b", "c", "a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortUserRoles(tt.args.roles, tt.args.guildRoles)
			if !reflect.DeepEqual(tt.args.roles, tt.expected) {
				t.Errorf("Expected %v, but was %v", tt.expected, tt.args.roles)
			}
		})
	}
}

func TestIsBlocked(t *testing.T) {
	type args struct {
		state *discordgo.State
		user  *discordgo.User
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no relationships",
			args: args{
				state: &discordgo.State{},
				user:  &discordgo.User{ID: "a"},
			},
			want: false,
		}, {
			name: "only a friendship with the user in question",
			args: args{
				state: &discordgo.State{
					Ready: discordgo.Ready{
						Relationships: []*discordgo.Relationship{
							&discordgo.Relationship{
								User: &discordgo.User{ID: "a"},
								Type: discordgo.RelationTypeFriend,
							},
						},
					},
				},
				user: &discordgo.User{ID: "a"},
			},
			want: false,
		}, {
			name: "only one blocked other user",
			args: args{
				state: &discordgo.State{
					Ready: discordgo.Ready{
						Relationships: []*discordgo.Relationship{
							&discordgo.Relationship{
								User: &discordgo.User{ID: "a"},
								Type: discordgo.RelationTypeBlocked,
							},
						},
					},
				},
				user: &discordgo.User{ID: "a"},
			},
			want: true,
		}, {
			name: "only one blocked other user",
			args: args{
				state: &discordgo.State{
					Ready: discordgo.Ready{
						Relationships: []*discordgo.Relationship{
							&discordgo.Relationship{
								User: &discordgo.User{ID: "b"},
								Type: discordgo.RelationTypeBlocked,
							},
						},
					},
				},
				user: &discordgo.User{ID: "a"},
			},
			want: false,
		}, {
			name: "multiple blocked users, including the one in question",
			args: args{
				state: &discordgo.State{
					Ready: discordgo.Ready{
						Relationships: []*discordgo.Relationship{
							&discordgo.Relationship{
								User: &discordgo.User{ID: "c"},
								Type: discordgo.RelationTypeBlocked,
							},
							&discordgo.Relationship{
								User: &discordgo.User{ID: "b"},
								Type: discordgo.RelationTypeBlocked,
							},
							&discordgo.Relationship{
								User: &discordgo.User{ID: "a"},
								Type: discordgo.RelationTypeBlocked,
							},
						},
					},
				},
				user: &discordgo.User{ID: "a"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlocked(tt.args.state, tt.args.user); got != tt.want {
				t.Errorf("IsBlocked() = %v, want %v", got, tt.want)
			}
		})
	}
}
