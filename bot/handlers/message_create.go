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

func parseArgs(prefix, command string, ind int, sep string) []string {
	args := strings.TrimPrefix(command, prefix)
	if ind == -1 {
		return strings.Fields(args)
	}

	parsedArgs := []string{}
	argsSplit := strings.Fields(args)

	for i, arg := range argsSplit {
		if i == ind {
			break
		}
		parsedArgs = append(parsedArgs, arg)
	}

	argsWithSep := strings.Join(argsSplit[ind:], " ")
	if sep == "" {
		return append(parsedArgs, argsWithSep)
	}
	return append(parsedArgs, strings.Split(argsWithSep, sep)...)
}

// CreateMessageCreateHandler create a handler for the MessageCreate Discord event.
func CreateMessageCreateHandler(p providers.MessageCreateProvider) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		if strings.HasPrefix(m.Content, "!scoreboard") {
			args := parseArgs("!scoreboard", m.Content, -1, "")
			week := 0
			var err error
			if len(args) > 0 {
				week, err = strconv.Atoi(args[0])
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
			args := parseArgs("!roster", m.Content, 0, "")
			if len(args) != 1 {
				s.ChannelMessageSend(m.ChannelID, usageError("roster"))
				return
			}
			s.ChannelMessageSend(m.ChannelID, p.Roster(args[0]))
		}

		if strings.HasPrefix(m.Content, "!stats ") {
			args := parseArgs("!stats", m.Content, 1, "")
			if len(args) < 2 {
				s.ChannelMessageSend(m.ChannelID, usageError("stats"))
				return
			}
			s.ChannelMessageSend(m.ChannelID, p.PlayerStats(args[0], args[1]))
		}

		if strings.HasPrefix(m.Content, "!compare ") {
			args := parseArgs("!compare", m.Content, 1, "/")
			if len(args) != 3 {
				s.ChannelMessageSend(m.ChannelID, usageError("compare"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.Compare(args[0], args[1], args[2]))
		}

		if strings.HasPrefix(m.Content, "!analyze ") {
			args := parseArgs("!analyze", m.Content, 1, ",")

			if len(args) < 2 {
				s.ChannelMessageSend(m.ChannelID, usageError("analyze"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.AnalyzeFreeAgents(args[0], args[1:]))
		}

		if strings.HasPrefix(m.Content, "!vs ") {
			args := parseArgs("!vs", m.Content, -1, "")
			if week, err := strconv.Atoi(args[0]); err == nil {
				tm := strings.Join(args[1:], " ")
				if tm == "" {
					s.ChannelMessageSend(m.ChannelID, usageError("vs"))
					return
				}
				s.ChannelMessageSend(m.ChannelID, p.VsLeague(tm, week))
				return
			}

			if len(args) == 0 {
				s.ChannelMessageSend(m.ChannelID, usageError("vs"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.VsLeague(strings.Join(args, " "), 0))
		}

		if strings.HasPrefix(m.Content, "!schedule ") {
			args := parseArgs("!schedule", m.Content, 0, "")
			if len(args) == 0 {
				s.ChannelMessageSend(m.ChannelID, usageError("schedule"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.Schedule(args[0]))
		}

		if strings.HasPrefix(m.Content, "!owner ") {
			args := parseArgs("!owner", m.Content, 0, ",")
			if len(args) == 0 {
				s.ChannelMessageSend(m.ChannelID, usageError("owner"))
				return
			}

			s.ChannelMessageSend(m.ChannelID, p.Owner(args))
		}

		if m.Content == "!help" {
			s.ChannelMessageSendEmbed(m.ChannelID, p.Help())
		}
	}
}
