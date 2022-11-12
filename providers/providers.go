package providers

import "github.com/bwmarrin/discordgo"

// MessageCreateProvider is the interface for providers that are accessed on the
// MessageCreate Discord event.
type MessageCreateProvider interface {
	Scoreboard(week int) string
	Standings() string
	Roster(teamName string) string
	PlayerStats(statsType, playerName string) string
	Help() *discordgo.MessageEmbed
}
