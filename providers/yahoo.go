package providers

import (
	"fmt"
	"strings"

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
	if week == 0 {
		sb, _ = y.yf.CurrentScoreboard(y.leagueKey)
	}

	sb, _ = y.yf.Scoreboard(y.leagueKey, week)
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
				"%d: %s (%d-%d-%d)\n",
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
	standings, _ := y.yf.Standings(y.leagueKey)
	return formatYahooStandings(standings)
}

func formatYahooRoster(teamName string, roster *schema.Roster) string {
	var out strings.Builder

	out.WriteString("```\n")
	out.WriteString(teamName)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(teamName)))
	out.WriteString("\n")

	possiblePos := []string{"PG", "SG", "G", "SF", "PF", "F", "C", "UTIL", "BN", "IL", "IL+"}
	ros := make(map[string][]string)
	for _, pos := range possiblePos {
		ros[pos] = []string{}
	}

	for _, player := range roster.Players.Player {
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
	ros, _ := y.yf.TeamRoster(y.leagueKey, teamName)
	return formatYahooRoster(teamName, ros)
}

func formatPlayerStats(name string, pStats *schema.PlayerStats) string {
	var out strings.Builder

	header := name + " - Season Average"
	out.WriteString("```\n")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")

	for _, s := range pStats.Stats.Stat {
		out.WriteString(fmt.Sprintf("%-3s: %s", statIDToName[s.StatID], s.Value))
		out.WriteString("\n")
	}

	out.WriteString("```")

	return out.String()
}

// PlayerStats returns a formatted string containing the stats for a player.
func (y *Yahoo) PlayerStats(playerName string) string {
	stats, _ := y.yf.PlayerStats(y.leagueKey, playerName)
	return formatPlayerStats(playerName, stats)
}
