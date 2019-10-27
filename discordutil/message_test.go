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

type messageSupplier struct {
	requestAmount int
}

func (l *messageSupplier) ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
	l.requestAmount++
	return nil, nil
}

func Test_LoadMessages_CacheAccess(t *testing.T) {
	t.Run("Test emtpy channel", func(t *testing.T) {
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
