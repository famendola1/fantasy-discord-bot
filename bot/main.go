package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/fantasy-discord-bot/bot/handlers"
	"github.com/famendola1/fantasy-discord-bot/providers"
	"github.com/famendola1/yauth"
)

var (
	authFile  = flag.String("auth_file", "", "Path to the file containing a serialized YAuth object.")
	game      = flag.String("game", "", "The Yahoo game key of the sport of the fantasy league.")
	leagueKey = flag.String("league", "", "The key of the Yahoo fantasy league.")
)

func main() {
	flag.Parse()

	if *authFile == "" {
		fmt.Println("No auth file provided.")
		return
	}

	if *game == "" {
		fmt.Println("No game provided.")
		return
	}

	if *leagueKey == "" {
		fmt.Println("No league provided.")
		return
	}

	// // Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	auth, _ := yauth.CreateYAuthFromJSON(*authFile)
	dg.AddHandler(handlers.CreateMessageCreateHandler(providers.NewYahooProvider(auth, *game, *leagueKey)))

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
