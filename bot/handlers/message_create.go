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
			args := strings.TrimPrefix(m.Content, "!stats ")
			splitArgs := strings.Fields(args)
			if (len(splitArgs)) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Error: invald !stats usage. See !help for usage.")
			}
			s.ChannelMessageSend(m.ChannelID, p.PlayerStats(splitArgs[0], strings.Join(splitArgs[1:], " ")))
		}

		if strings.HasPrefix(m.Content, "!compare ") {
			args := strings.TrimPrefix(m.Content, "!compare ")
			splitArgs := strings.Fields(args)
			if (len(splitArgs)) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Error: invald !compare usage. See !help for usage.")
			}

			playersJoined := strings.Join(splitArgs[1:], " ")
			players := strings.Split(playersJoined, "/")

			if (len(players)) != 2 {
				s.ChannelMessageSend(m.ChannelID, "Error: invald !compare usage. See !help for usage.")
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.Compare(splitArgs[0], players[0], players[1]))
		}

		if strings.HasPrefix(m.Content, "!analyze ") {
			args := strings.TrimPrefix(m.Content, "!analyze ")
			splitArgs := strings.Fields(args)

			if (len(splitArgs)) != 2 {
				s.ChannelMessageSend(m.ChannelID, "Error: invald !analyze usage. See !help for usage.")
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.AnalyzeFreeAgents(splitArgs[0], strings.Split(splitArgs[1], ",")))
		}

		if m.Content == "!help" {
			s.ChannelMessageSendEmbed(m.ChannelID, p.Help())
		}
	}
}
