package providers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/yauth"
	"github.com/famendola1/yflib"
	"github.com/famendola1/yfquery"
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
	orderedStats9CAT = []int{5, 8, 10, 12, 15, 16, 17, 18, 19}
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
			} else if matchup.WinnerTeamKey == tm.TeamKey {
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

// Leaders returns the stat category leaders for a given day.
func (y *Yahoo) Leaders(date string) string {
	if date == "yesterday" {
		pst, _ := time.LoadLocation("America/Los_Angeles")
		date = time.Now().In(pst).AddDate(0, 0, -1).Format("2006-01-02")
	}

	leaders := make(map[int][]schema.Player)
	for stat := range statIDs9CAT {
		players, err := yflib.StatCategoryLeaders(y.client, date, y.gameKey, stat, 5)
		if err != nil {
			return formatError(err)
		}
		leaders[stat] = append(leaders[stat], players...)
	}

	var out strings.Builder
	out.WriteString("```\n")
	header := fmt.Sprintf("%s Stat Leaders\n", date)
	out.WriteString(header)
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")
	for _, stat := range orderedStats9CAT {
		players := leaders[stat]
		out.WriteString(yflib.StatIDToName[stat] + "\n")
		out.WriteString(strings.Repeat("-", 25) + "\n")
		for _, p := range players {
			out.WriteString(fmt.Sprintf("%s - %s", p.Name.Full, p.DisplayPosition))
			for _, s := range p.PlayerStats.Stats.Stat {
				if s.StatID == stat {
					out.WriteString(fmt.Sprintf(" (%s)\n", s.Value))
				}
			}
		}
		out.WriteString("\n")
	}
	out.WriteString("```")
	return out.String()
}

// HeadToHead displays the matchup results between the two given teams on the given week.
func (y *Yahoo) HeadToHead(week int, teamA, teamB string) string {
	allTeams, err := yfquery.League().Key(y.leagueKey).Teams().Stats().Week(week).Get(y.client)
	if err != nil {
		return formatError(err)
	}

	var teamAStats *schema.TeamStats
	var teamBStats *schema.TeamStats

	for _, tm := range allTeams.League.Teams.Team {
		if tm.Name == teamA {
			teamAStats = tm.TeamStats
			continue
		}

		if tm.Name == teamB {
			teamBStats = tm.TeamStats
			continue
		}
	}

	if teamAStats == nil {
		return formatError(fmt.Errorf("%q team not found", teamA))
	}
	if teamBStats == nil {
		return formatError(fmt.Errorf("%q team not found", teamB))
	}

	var out strings.Builder
	out.WriteString("```\n")
	header := fmt.Sprintf("H2H: %s vs %s\n", teamA, teamB)
	out.WriteString(header)
	out.WriteString(strings.Repeat("-", len(header)))
	out.WriteString("\n\n")

	var w, l, t int
	for i, stat := range teamAStats.Stats.Stat {
		teamAVal := stat.Value
		teamBVal := teamBStats.Stats.Stat[i].Value
		if statIDs9CAT[stat.StatID] {
			teamAStat, _ := strconv.ParseFloat(stat.Value, 32)
			teamBStat, _ := strconv.ParseFloat(teamBStats.Stats.Stat[i].Value, 32)
			if teamAStat == teamBStat {
				t++
			}

			if teamAStat > teamBStat {
				w++
			}

			if teamAStat < teamBStat {
				l++
			}

			if stat.StatID == 19 && teamAStat > teamBStat {
				l++
				w--
			}

			if stat.StatID == 19 && teamAStat < teamBStat {
				l--
				w++
			}
		}

		out.WriteString(fmt.Sprintf("%-3s: %8s | %s\n", yflib.StatIDToName[stat.StatID], teamAVal, teamBVal))
	}
	out.WriteString(fmt.Sprintf("\nTotal: %d-%d-%d", w, l, t))
	out.WriteString("```")

	return out.String()
}

func sortTeamsByStat(tms *schema.Teams, statID int) []schema.Team {
	teams := tms.Team
	sort.Slice(teams, func(i, j int) bool {
		for k, stat := range teams[i].TeamStats.Stats.Stat {
			if stat.StatID != statID {
				continue
			}

			valI, _ := strconv.ParseFloat(stat.Value, 32)
			valJ, _ := strconv.ParseFloat(teams[j].TeamStats.Stats.Stat[k].Value, 32)

			if valI > valJ {
				return true
			}
			return false
		}
		return false
	})
	return teams
}

// Ranks sorts all the teams by the given stat for the given week and returns
// the sorted list as a string. If no week is given, the current week is used.
func (y *Yahoo) Ranks(week int, stat string) string {
	if _, found := yflib.StatNameToID[strings.ToUpper(stat)]; !found {
		return formatError(fmt.Errorf("stat %q not found", stat))
	}
	allTeams, err := yfquery.League().Key(y.leagueKey).Teams().Stats().Week(week).Get(y.client)
	if err != nil {
		return formatError(err)
	}

	sortedTms := sortTeamsByStat(allTeams.League.Teams, yflib.StatNameToID[strings.ToUpper(stat)])

	var out strings.Builder
	out.WriteString("```\n")
	for i, tm := range sortedTms {
		val := ""
		for _, s := range tm.TeamStats.Stats.Stat {
			if s.StatID == yflib.StatNameToID[strings.ToUpper(stat)] {
				val = s.Value
			}
		}
		out.WriteString(fmt.Sprintf("%2d: %s - %s\n", i+1, tm.Name, val))
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
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!leaders <date>",
			Value: "Returns the stat category leaders for a given day. date is formatted as YYYY-MM-DD, if no date is provided then the current date in America/Los_Angeles is used. 'yesterday' can be used as a shortcut for the previous day's leaders.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!h2h [week] <team1>/<team2>",
			Value: "Returns the matchup result between the two given teams for the given week. If no week is provided, the current week is used.",
		})
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:  "!ranks [week] <stat>",
			Value: "Returns the team ranking for the given stat for the given week. If no week is provided, the current week is used.",
		})
	return embed
}
