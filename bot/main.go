package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/famendola1/fantasy-discord-bot/bot/handlers"
	"github.com/famendola1/fantasy-discord-bot/providers"
	"github.com/famendola1/yauth"
)

var (
	cfg = flag.String("cfg", "", "Path to the config file containing")
)

type config struct {
	Auth         yauth.YAuth `json:"auth"`
	Game         string      `json:"game"`
	LeagueID     int         `json:"league_id"`
	DiscordToken string      `json:"discord_token"`
}

func main() {
	flag.Parse()

	if *cfg == "" {
		fmt.Println("no config file specified.")
		return
	}

	content, err := ioutil.ReadFile(*cfg)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var conf config
	err = json.Unmarshal(content, &conf)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	// // Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + conf.DiscordToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(handlers.CreateMessageCreateHandler(providers.NewYahooProvider(&conf.Auth, conf.Game, conf.LeagueID)))

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
