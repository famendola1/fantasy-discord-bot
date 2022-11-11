package handlers

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/fantasy-discord-bot/providers"
)

// CreateMessageCreateHandler create a handler for the MessageCreate Discord event.
func CreateMessageCreateHandler(p providers.MessageCreateProvider) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		if strings.HasPrefix(m.Content, "!scoreboard") {
			pieces := strings.Fields(m.Content)
			week := 0
			if len(pieces) > 1 {
				week, _ = strconv.Atoi(pieces[1])
			}
			s.ChannelMessageSend(m.ChannelID, p.Scoreboard(week))
		}

		if m.Content == "!standings" {
			s.ChannelMessageSend(m.ChannelID, p.Standings())
		}

		if strings.HasPrefix(m.Content, "!roster ") {
			teamName := strings.TrimPrefix(m.Content, "!roster ")
			s.ChannelMessageSend(m.ChannelID, p.Roster(teamName))
		}

		if strings.HasPrefix(m.Content, "!stats ") {
			name := strings.TrimPrefix(m.Content, "!stats ")
			s.ChannelMessageSend(m.ChannelID, p.PlayerStats(name))
		}

		if m.Content == "!help" {
			s.ChannelMessageSendEmbed(m.ChannelID, p.Help())
		}
	}
}
