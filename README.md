# Fantasy Sports Discord Bot
A Discord bot for fantasy sports

## Disclaimer
Currently, only Yahoo fantasy basketball is supported.

## Before You Start
Before you start you will need to register an app as a developer for Discord and Yahoo Fantasy.

* [Discord](https://www.upwork.com/resources/how-to-make-discord-bot)
  * You'll need to save the Token for your bot
* [Yahoo Fantasy](https://developer.yahoo.com/apps/create/)
  * You'll need to save the consumer key and secret
  
## Configuration
The bot is configured using a JSON file, whose path is passed to the bot via a command line flag. The configuratin schema is as follows:

```json
{
	"auth": {
		"client_id": "",
		"client_secret": "",
		"token": {
			"access_token": "",
			"token_type": "",
			"refresh_token": "",
			"expiry": ""
		}
	},

	"game": "",
	"provider": "",
	"league_id": ,
	"discord_token": ""
}
```
* `auth` is modeled after the YAuth object from https://pkg.go.dev/github.com/famendola1/yauth. You can use the `yauth` package to generate this auth object.
* `game` is the sport of the fantasy league.
* `provider` is the fantasy sports provider. Currently only "yahoo" is supported.
* `league_id` is the ID if your Yahoo fantasy league. This can be found in the URL of your league's homepage.
* `discord_token` is the token of your Discord bot.

## Running the bot locally
```bash
go run bot/main.go --cfg=conf.json
```

## Deploying the bot
A Dockerfile is provided to package the bot into an image. Before, packaging the bot you must add a `conf.json` file to the project's root directory.

```bash
docker build --tag bot
```

You can deploy the Docker image using your preferred method.
