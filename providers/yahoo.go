package providers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/yauth"
	"github.com/famendola1/yfantasy"
	"github.com/famendola1/yfantasy/schema"
)

var (
	statIDToName = map[int]string{
		9004003: "FG",
		5:       "FG%",
		9007006: "FT",
		8:       "FT%",
		10:      "3PM",
		12:      "PTS",
		15:      "REB",
		16:      "AST",
		17:      "STL",
		18:      "BLK",
		19:      "TOV",
	}
)

// Yahoo is a provider for Yahoo Fantasy Sports.
type Yahoo struct {
	yf        *yfantasy.YFantasy
	gameKey   string
	leagueKey string
}

// NewYahooProvider returns a new Yahoo provider
func NewYahooProvider(auth *yauth.YAuth, gameKey string, leagueID int) *Yahoo {
	return &Yahoo{
		yf:        yfantasy.New(auth.Client()),
		gameKey:   gameKey,
		leagueKey: yfantasy.MakeLeagueKey(gameKey, leagueID),
	}
}

func formatError(err error) string {
	var out strings.Builder
	out.WriteString("```\n")
	out.WriteString(fmt.Sprintf("Error: %s", err))
	out.WriteString("```")
	return out.String()
}

func formatYahooScoreboard(sb *schema.Scoreboard) string {
	var out strings.Builder

	header := fmt.Sprintf("Week %d Matchups", sb.Matchups.Matchup[0].Week)

	score := make(map[string]int)
	out.WriteString("```\n")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n")
	for _, m := range sb.Matchups.Matchup {
		for _, s := range m.StatWinners.StatWinner {
			score[s.WinnerTeamKey]++
		}

		out.WriteString(fmt.Sprintf("%s (%d)\n", m.Teams.Team[0].Name, score[m.Teams.Team[0].TeamKey]))
		out.WriteString(fmt.Sprintf("%s (%d)\n", m.Teams.Team[1].Name, score[m.Teams.Team[1].TeamKey]))
		out.WriteString("\n")
	}
	out.WriteString("```")

	return out.String()
}

// Scoreboard returns a formatted string of all the Yahoo matchups for the given
// week. If week is -1, then the current week is used.
func (y *Yahoo) Scoreboard(week int) string {
	var sb *schema.Scoreboard
	var err error

	if week == 0 {
		sb, err = y.yf.CurrentScoreboard(y.leagueKey)
	} else {
		sb, err = y.yf.Scoreboard(y.leagueKey, week)
	}

	if err != nil {
		return formatError(err)
	}
	return formatYahooScoreboard(sb)
}

func formatYahooStandings(standings *schema.Standings) string {
	var out strings.Builder

	header := fmt.Sprintln("Standings")
	out.WriteString("```\n")
	out.WriteString(header)
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n")

	for _, tm := range standings.Teams.Team {
		out.WriteString(
			fmt.Sprintf(
				"%2d: %s (%d-%d-%d)\n",
				tm.TeamStandings.Rank,
				tm.Name,
				tm.TeamStandings.OutcomeTotals.Wins,
				tm.TeamStandings.OutcomeTotals.Losses,
				tm.TeamStandings.OutcomeTotals.Ties))
	}
	out.WriteString("```")

	return out.String()
}

// Standings returns a formatted string containing the Yahoo league's standings.
func (y *Yahoo) Standings() string {
	standings, err := y.yf.Standings(y.leagueKey)
	if err != nil {
		return formatError(err)
	}
	return formatYahooStandings(standings)
}

func formatYahooRoster(team *schema.Team) string {
	var out strings.Builder

	out.WriteString("```\n")
	out.WriteString(team.Name)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(team.Name)))
	out.WriteString("\n")

	possiblePos := []string{"PG", "SG", "G", "SF", "PF", "F", "C", "UTIL", "BN", "IL", "IL+"}
	ros := make(map[string][]string)
	for _, pos := range possiblePos {
		ros[pos] = []string{}
	}

	for _, player := range team.Roster.Players.Player {
		ros[player.SelectedPosition.Position] = append(ros[player.SelectedPosition.Position], player.Name.Full)
	}

	for _, pos := range possiblePos {
		for _, name := range ros[pos] {
			out.WriteString(fmt.Sprintf("%s: %s\n", pos, name))
		}
	}

	out.WriteString("```")

	return out.String()
}

// Roster returns a formatted string containg the roster of a team.
func (y *Yahoo) Roster(teamName string) string {
	tm, err := y.yf.TeamRoster(y.leagueKey, teamName)
	if err != nil {
		return formatError(err)
	}
	return formatYahooRoster(tm)
}

func formatPlayerStats(player *schema.Player) string {
	var out strings.Builder

	header := player.Name.Full + " - " + strings.Title(strings.Replace(player.PlayerStats.CoverageType, "_", " ", 1))
	out.WriteString("```\n")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")

	for _, s := range player.PlayerStats.Stats.Stat {
		out.WriteString(fmt.Sprintf("%-3s: %s", statIDToName[s.StatID], s.Value))
		out.WriteString("\n")
	}

	out.WriteString("```")

	return out.String()
}

// PlayerStats returns a formatted string containing the stats for a player.
func (y *Yahoo) PlayerStats(statsType, playerName string) string {
	var statsTypeNum int
	switch statsType {
	case "season":
		statsTypeNum = yfantasy.StatsTypeAverageSeason
		break
	case "week":
		statsTypeNum = yfantasy.StatsTypeLastWeekAverage
		break
	case "month":
		statsTypeNum = yfantasy.StatsTypeLastMonthAverage
		break
	default:
		return formatError(fmt.Errorf("invald stats type (%q) requested", statsType))
	}

	p, err := y.yf.PlayerStats(y.leagueKey, playerName, statsTypeNum)
	if err != nil {
		return formatError(err)
	}
	return formatPlayerStats(p)
}

// Help returns the help docs.
func (y *Yahoo) Help() *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "Yahoo Fantasy Sports Bot",
		Description: "Discord Bot for Yahoo Fantasy Sports",
	}

	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{Name: "!help", Value: "Returns this message."})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!scoreboard <week>",
			Value: "Returns the scoreboard of the given week. If no week is provided, returns the current scoreboard.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{Name: "!roster <team>", Value: "Returns the roster of the given team."})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!stats <type> <player>",
			Value: "Returns the stats of the requested player. The provided player's name must be at least 3 letters long. <type> must be one of season|week|month.",
		})
	return embed
}
