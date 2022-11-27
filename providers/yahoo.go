package providers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/yauth"
	"github.com/famendola1/yflib"
	"github.com/famendola1/yfquery/schema"
)

// Yahoo is a provider for Yahoo Fantasy Sports.
type Yahoo struct {
	client    *http.Client
	gameKey   string
	leagueKey string
}

var (
	statIDs9CAT = map[int]bool{
		5:  true,
		8:  true,
		10: true,
		12: true,
		15: true,
		16: true,
		17: true,
		18: true,
		19: true,
	}
)

// NewYahooProvider returns a new Yahoo provider
func NewYahooProvider(auth *yauth.YAuth, gameKey string, leagueID int) *Yahoo {
	return &Yahoo{
		client:    auth.Client(),
		gameKey:   gameKey,
		leagueKey: yflib.MakeLeagueKey(gameKey, leagueID),
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
		sb, err = yflib.GetCurrentScoreboard(y.client, y.leagueKey)
	} else {
		sb, err = yflib.GetScoreboard(y.client, y.leagueKey, week)
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
	standings, err := yflib.GetLeagueStandings(y.client, y.leagueKey)
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
	tm, err := yflib.GetTeamRoster(y.client, y.leagueKey, teamName)
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
		out.WriteString(fmt.Sprintf("%-3s: %s", yflib.StatIDToName[s.StatID], s.Value))
		out.WriteString("\n")
	}

	out.WriteString("```")

	return out.String()
}

func convertStatsType(statsType string) (int, error) {
	var statsTypeNum int
	switch statsType {
	case "season":
		statsTypeNum = yflib.StatsTypeAverageSeason
		break
	case "week":
		statsTypeNum = yflib.StatsTypeAverageLastWeek
		break
	case "month":
		statsTypeNum = yflib.StatsTypeAverageLastMonth
		break
	default:
		return yflib.StatsTypeUnknown, fmt.Errorf("invald stats type (%q) requested", statsType)
	}

	return statsTypeNum, nil
}

// PlayerStats returns a formatted string containing the stats for a player.
func (y *Yahoo) PlayerStats(statsType, playerName string) string {
	statsTypeNum, err := convertStatsType(statsType)
	if err != nil {
		return formatError(err)
	}

	p, err := yflib.GetPlayerStats(y.client, y.leagueKey, playerName, statsTypeNum)
	if err != nil {
		return formatError(err)
	}
	return formatPlayerStats(p)
}

func formatStatsDiff(diff *yflib.StatsDiff) string {
	var out strings.Builder

	header := diff.PlayerA + " / " + diff.PlayerB
	out.WriteString("```")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")

	out.WriteString(fmt.Sprintf("%-3s: %.3f\n", "FG%", diff.Diffs[yflib.StatNameToID["FG%"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.3f\n", "FT%", diff.Diffs[yflib.StatNameToID["FT%"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "3PM", diff.Diffs[yflib.StatNameToID["3PM"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "PTS", diff.Diffs[yflib.StatNameToID["PTS"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "REB", diff.Diffs[yflib.StatNameToID["REB"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "AST", diff.Diffs[yflib.StatNameToID["AST"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "STL", diff.Diffs[yflib.StatNameToID["STL"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "BLK", diff.Diffs[yflib.StatNameToID["BLK"]]))
	out.WriteString(fmt.Sprintf("%-3s: %.1f\n", "TOV", diff.Diffs[yflib.StatNameToID["TOV"]]))

	out.WriteString("\n")
	out.WriteString("```")

	return out.String()
}

// Compare computes the difference in stats between the two provided players.
func (y *Yahoo) Compare(statsType, playerA, playerB string) string {
	statsTypeNum, err := convertStatsType(statsType)
	if err != nil {
		return formatError(err)
	}

	diff, err := yflib.ComparePlayers(y.client, y.leagueKey, playerA, playerB, statsTypeNum, yflib.NBA9CATIDs)
	if err != nil {
		return formatError(err)
	}

	return formatStatsDiff(diff)
}

func formatFreeAgents(freeAgents map[string][]*schema.Player) string {
	var out strings.Builder

	out.WriteString("```")
	out.WriteString("\n")
	for stat, players := range freeAgents {
		out.WriteString(stat)
		out.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("-", 20)))
		for _, player := range players {
			out.WriteString(player.Name.Full)
			for _, s := range player.PlayerStats.Stats.Stat {
				if yflib.StatIDToName[s.StatID] == stat {
					out.WriteString(fmt.Sprintf(" (%s)", s.Value))
				}
			}
			out.WriteString("\n")
		}
		out.WriteString("\n\n")
	}
	out.WriteString("\n```")

	return out.String()
}

// AnalyzeFreeAgents prints the top 5 players for the given stats with th given type.
func (y *Yahoo) AnalyzeFreeAgents(statsType string, stats []string) string {
	statsTypeNum, err := convertStatsType(statsType)
	if err != nil {
		return formatError(err)
	}

	freeAgents := make(map[string][]*schema.Player)
	for _, stat := range stats {
		players, err := yflib.SortFreeAgentsByStat(y.client, y.leagueKey, yflib.StatNameToID[strings.ToUpper(stat)], 5, statsTypeNum)
		if err != nil {
			return formatError(err)
		}

		freeAgents[strings.ToUpper(stat)] = players
	}
	return formatFreeAgents(freeAgents)
}

type matchupResult struct {
	win, tie int
}

func formatCategoryMatchupResults(results []yflib.CategoryMatchupResult) string {
	var out strings.Builder

	header := results[0].HomeTeam + " vs. The League"
	out.WriteString("```")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")

	win := 0
	loss := 0
	tie := 0
	for _, res := range results {
		out.WriteString(fmt.Sprintf("%s (%d)\n", res.HomeTeam, len(res.CategoriesWon)))
		out.WriteString(fmt.Sprintf("%s (%d)\n\n", res.AwayTeam, len(res.CategoriesLost)))

		if len(res.CategoriesWon) > len(res.CategoriesLost) {
			win++
		} else if len(res.CategoriesWon) < len(res.CategoriesLost) {
			loss++
		} else {
			tie++
		}
	}

	out.WriteString(fmt.Sprintf("Total: %d-%d-%d", win, loss, tie))
	out.WriteString("```")

	return out.String()
}

// VsLeague computes the given teams matchup outcome against every other team in the league.
func (y *Yahoo) VsLeague(teamName string, week int) string {
	results, err := yflib.CalculateCategoryMathchupResultsVsLeague(y.client, y.leagueKey, teamName, yflib.NBA9CATIDs, week)
	if err != nil {
		return formatError(err)
	}
	return formatCategoryMatchupResults(results)
}

// Schedule returns the season schedule for the given team.
func (y *Yahoo) Schedule(teamName string) string {
	tm, err := yflib.GetTeamMatchups(y.client, y.leagueKey, teamName)
	if err != nil {
		return formatError(err)
	}

	var out strings.Builder
	header := tm.Name + " Schedule"
	out.WriteString("```")
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")

	win := 0
	loss := 0
	tie := 0
	for _, matchup := range tm.Matchups.Matchup {
		switch matchup.Status {
		case "postevent":
			var result string
			if matchup.IsTied {
				result = "T"
				tie++
			}
			if matchup.WinnerTeamKey == tm.TeamKey {
				result = "W"
				win++
			} else {
				result = "L"
				loss++
			}
			out.WriteString(fmt.Sprintf("%2d: %s (%s)\n", matchup.Week, matchup.Teams.Team[1].Name, result))
		case "midevent":
			out.WriteString(fmt.Sprintf("%2d: *%s*\n", matchup.Week, matchup.Teams.Team[1].Name))
		case "preevent":
			out.WriteString(fmt.Sprintf("%2d: %s\n", matchup.Week, matchup.Teams.Team[1].Name))
		}
	}
	out.WriteString(fmt.Sprintf("\nTotal: %d-%d-%d", win, loss, tie))
	out.WriteString("```")

	return out.String()
}

// Owner returns the owner for all the provided players.
func (y *Yahoo) Owner(playerNames []string) string {
	players := []*schema.Player{}
	for _, name := range playerNames {
		player, err := yflib.GetPlayerOwnership(y.client, y.leagueKey, name)
		if err != nil {
			return formatError(err)
		}
		players = append(players, player)
	}

	var out strings.Builder
	out.WriteString("```")

	for _, player := range players {
		out.WriteString(fmt.Sprintf("%s: ", player.Name.Full))
		switch player.Ownership.OwnershipType {
		case "freeagents":
			out.WriteString("Free Agent")
		case "waivers":
			t, _ := time.Parse("2006-01-02", player.Ownership.WaiverDate)
			out.WriteString(fmt.Sprintf("Waivers (%s)", t.Format("Mon 01/02")))
		case "team":
			out.WriteString(player.Ownership.OwnerTeamName)
		}
		out.WriteString("\n\n")
	}
	out.WriteString("```")
	return out.String()
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
			Name:  "!scoreboard [week]",
			Value: "Returns the scoreboard of the given week. If no week is provided, returns the current scoreboard.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!standings",
			Value: "Returns the current league standings.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!roster <team>",
			Value: "Returns the roster of the given team.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!stats <type> <player>",
			Value: "Returns the stats of the requested player. The provided player's name must be at least 3 letters long. <type> must be one of season|week|month.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!compare <type> <player1>/<player2>",
			Value: "Returns the difference in stats between player1 and player2. The provided players' names must be at least 3 letters long. <type> must be one of season|week|month.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!analyze <type> <stat1>,<stat2>,...",
			Value: "Returns the top 5 free agents for each stat. <type> must be one of season|week|month.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!vs [week] <team>",
			Value: "Returns the matchups results of the provided team against all other teams in the league. If week is not provided, the current week is used.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!schedule <team>",
			Value: "Returns season schedule of the provided team.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!owner <player1>,<player2>,...",
			Value: "Returns the current owner of the provided players.",
		})
	return embed
}
