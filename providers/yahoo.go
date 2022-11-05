package providers

import (
	"fmt"
	"strings"

	"github.com/famendola1/yauth"
	"github.com/famendola1/yfantasy"
)

// Yahoo is a provider for Yahoo Fantasy Sports.
type Yahoo struct {
	yf        *yfantasy.YFantasy
	gameKey   string
	leagueKey string
}

// NewYahooProvider returns a new Yahoo provider
func NewYahooProvider(auth *yauth.YAuth, gameKey, leagueKey string) *Yahoo {
	return &Yahoo{
		yf:        yfantasy.New(auth.Client()),
		gameKey:   gameKey,
		leagueKey: leagueKey,
	}
}

func formatYahooScoreboard(matchups *yfantasy.Matchups) string {
	var out strings.Builder

	header := fmt.Sprintf("Week %d Matchups", matchups.Matchup[0].Week)

	score := make(map[string]int)
	out.WriteString("```\n")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n")
	for _, m := range matchups.Matchup {
		for _, s := range m.StatWinners.StatWinner {
			score[s.WinnerTeamKey]++
		}

		out.WriteString(fmt.Sprintf("%-30s (%d)\n", m.Teams.Team[0].Name, score[m.Teams.Team[0].TeamKey]))
		out.WriteString(fmt.Sprintf("%-30s (%d)\n", m.Teams.Team[1].Name, score[m.Teams.Team[1].TeamKey]))
		out.WriteString("\n")
	}
	out.WriteString("```")

	return out.String()
}

// Scoreboard returns a formatted string of all the Yahoo matchups for the given
// week. If week is -1, then the current week is used.
func (y *Yahoo) Scoreboard(week int) string {
	gm := y.yf.Game(y.gameKey)
	lg, _ := gm.League(y.leagueKey)

	if week == -1 {
		week = lg.CurrentWeek
	}

	sb, _ := lg.GetScoreboard(week)
	return formatYahooScoreboard(sb)
}

func formatYahooStandings(standings *yfantasy.Standings) string {
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
	gm := y.yf.Game(y.gameKey)
	lg, _ := gm.League(y.leagueKey)
	standings, _ := lg.GetStandings()

	return formatYahooStandings(standings)
}
