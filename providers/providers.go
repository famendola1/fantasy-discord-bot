package providers

import "github.com/bwmarrin/discordgo"

// MessageCreateProvider is the interface for providers that are accessed on the
// MessageCreate Discord event.
type MessageCreateProvider interface {
	Scoreboard(week int) string
	Standings() string
	Roster(teamName string) string
	PlayerStats(statsType, playerName string) string
	Compare(statsType, playerA, playerB string) string
	AnalyzeFreeAgents(statsType string, stats []string) string
	VsLeague(teamName string, week int) string
	Schedule(teamName string) string
	Owner(playerName []string) string
	Leaders(date string) string
	HeadToHead(week int, teamA, teamB string) string
	Ranks(week int, stat string) string
	Help() *discordgo.MessageEmbed
}
