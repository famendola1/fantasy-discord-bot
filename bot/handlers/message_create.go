package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/fantasy-discord-bot/providers"
)

func usageError(comm string) string {
	return fmt.Sprintf("Error: invald !%s usage. See !help for usage.", comm)
}

// CreateMessageCreateHandler create a handler for the MessageCreate Discord event.
func CreateMessageCreateHandler(p providers.MessageCreateProvider) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		if strings.HasPrefix(m.Content, "!scoreboard") {
			args := strings.Fields(m.Content)
			week := 0
			var err error
			if len(args) > 1 {
				week, err = strconv.Atoi(args[1])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, usageError("scoreboard"))
					return
				}
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
				s.ChannelMessageSend(m.ChannelID, usageError("stats"))
				return
			}
			s.ChannelMessageSend(m.ChannelID, p.PlayerStats(splitArgs[0], strings.Join(splitArgs[1:], " ")))
		}

		if strings.HasPrefix(m.Content, "!compare ") {
			args := strings.TrimPrefix(m.Content, "!compare ")
			splitArgs := strings.Fields(args)
			if (len(splitArgs)) < 2 {
				s.ChannelMessageSend(m.ChannelID, usageError("compare"))
				return
			}

			playersJoined := strings.Join(splitArgs[1:], " ")
			players := strings.Split(playersJoined, "/")

			if (len(players)) != 2 {
				s.ChannelMessageSend(m.ChannelID, usageError("compare"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.Compare(splitArgs[0], players[0], players[1]))
		}

		if strings.HasPrefix(m.Content, "!analyze ") {
			args := strings.TrimPrefix(m.Content, "!analyze ")
			splitArgs := strings.Fields(args)

			if (len(splitArgs)) != 2 {
				s.ChannelMessageSend(m.ChannelID, usageError("analyze"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.AnalyzeFreeAgents(splitArgs[0], strings.Split(splitArgs[1], ",")))
		}

		if strings.HasPrefix(m.Content, "!vs ") {
			args := strings.TrimPrefix(m.Content, "!vs ")
			argsSplit := strings.Fields(args)
			if week, err := strconv.Atoi(argsSplit[0]); err == nil {
				tm := strings.Join(argsSplit[1:], " ")
				if tm == "" {
					s.ChannelMessageSend(m.ChannelID, usageError("vs"))
					return
				}
				s.ChannelMessageSend(m.ChannelID, p.VsLeague(tm, week))
				return
			}

			if args == "" {
				s.ChannelMessageSend(m.ChannelID, usageError("vs"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.VsLeague(args, 0))
		}

		if m.Content == "!help" {
			s.ChannelMessageSendEmbed(m.ChannelID, p.Help())
		}
	}
}
