package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	ytdl "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/music"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/music/player"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/search"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
	"github.com/HalvaPovidlo/discordBotGo/pkg/context"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

// @title           HalvaBot for Discord
// @version         1.0
// @description     A music discord bot.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api
func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	logger := zap.NewLogger()
	ctx := context.WithLogger(context.Background(), logger)

	// Initialize discord session
	session, err := discord.OpenSession(cfg.Discord.Token, logger)
	if err != nil {
		panic(err)
	}
	defer func(session *discordgo.Session) {
		_ = session.Close()
		logger.Infow("Bot session closed")
	}(session)

	// YouTube services
	ytService, err := youtube.NewService(ctx, option.WithCredentialsFile("halvabot-google.json"))
	if err != nil {
		panic(err)
	}

	ytClient := search.NewYouTubeClient(&ytdl.Client{
		Debug:      true,
		HTTPClient: http.DefaultClient,
	}, ytService)

	// Music stage
	voiceClient := voice.NewVoice(session, cfg.Discord.Voice)

	musicPlayer := player.NewPlayer(ytClient, voiceClient, cfg.Discord, logger)

	musicCog := music.NewCog(musicPlayer, cfg.Discord.Prefix, logger)
	musicCog.RegisterCommands(session)

	// Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Infow("Graceful shutdown")
	defer logger.Sync()
}
