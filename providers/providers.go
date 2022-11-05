package providers

// MessageCreateProvider is the interface for providers that are accessed on the
// MessageCreate Discord event.
type MessageCreateProvider interface {
	Scoreboard(week int) string
	Standings() string
}
