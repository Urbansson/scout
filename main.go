package main

import (
	"flag"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var ipProviders = []string{
	"http://ifconfig.me",
	"http://ip.me",
}

var (
	BotToken = flag.String("token", "", "Bot access token")
	logger   *slog.Logger
)

var (
	defaultMemberPermissions int64 = discordgo.PermissionSendMessages
	command                        = discordgo.ApplicationCommand{
		Name:                     "get-ip",
		Description:              "Get server public IP",
		DefaultMemberPermissions: &defaultMemberPermissions,
	}
)

func main() {
	flag.Parse()

	// Initialize structured logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Scout Discord bot")

	if BotToken == nil || *BotToken == "" {
		logger.Error("missing flag --token")
		os.Exit(1)
	}
	logger.Info("Bot token provided, creating Discord session")

	dg, err := discordgo.New("Bot " + *BotToken)
	if err != nil {
		logger.Error("Error creating Discord session", "error", err)
		os.Exit(1)
	}
	defer dg.Close()
	logger.Info("Discord session created successfully")

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Info("Bot logged in",
			"username", s.State.User.Username,
			"discriminator", s.State.User.Discriminator)
	})

	dg.AddHandler(func(s *discordgo.Session, gc *discordgo.GuildCreate) {
		logger.Info("Joined guild",
			"guild_name", gc.Guild.Name,
			"guild_id", gc.Guild.ID)
		logger.Info("Registering command",
			"command", command.Name,
			"guild", gc.Guild.Name)
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, gc.ID, &command)
		if err != nil {
			logger.Error("Error registering command",
				"guild", gc.Guild.Name,
				"error", err)
		} else {
			logger.Info("Successfully registered command",
				"command", command.Name,
				"guild", gc.Guild.Name)
		}
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commandName := i.ApplicationCommandData().Name
		logger.Info("Received interaction",
			"command", commandName,
			"user", i.Member.User.Username,
			"discriminator", i.Member.User.Discriminator)

		if commandName == "get-ip" {
			logger.Info("Fetching current IP address")
			ip := getCurrentIP()

			if ip == "" {
				logger.Warn("Failed to retrieve IP address")
				ip = "Failed to retrieve IP address. Check logs for details."
			} else {
				logger.Info("Successfully retrieved IP", "ip", ip)
			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ip,
				},
			})
			if err != nil {
				logger.Error("Error responding to interaction", "error", err)
			} else {
				logger.Info("Successfully sent response to user")
			}
		} else {
			logger.Warn("Unknown command received", "command", commandName)
		}
	})

	logger.Info("Opening Discord connection")
	err = dg.Open()
	if err != nil {
		logger.Error("Error opening Discord connection", "error", err)
		os.Exit(1)
	}
	logger.Info("Discord connection opened successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logger.Info("Bot is now running. Press Ctrl+C to exit")
	<-stop
	logger.Info("Received shutdown signal, closing bot")
}

func getCurrentIP() string {
	logger.Info("Attempting to retrieve IP", "provider_count", len(ipProviders))

	for i, provider := range ipProviders {
		logger.Info("Trying provider",
			"index", i+1,
			"total", len(ipProviders),
			"provider", provider)

		resp, err := http.Get(provider)
		if err != nil {
			logger.Error("Error getting IP from provider",
				"provider", provider,
				"error", err)
			continue
		}
		defer resp.Body.Close()

		logger.Info("Response received",
			"provider", provider,
			"status_code", resp.StatusCode,
			"status", http.StatusText(resp.StatusCode))

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Error reading response body",
					"provider", provider,
					"error", err)
				continue
			}
			ip := string(body)
			logger.Info("Successfully retrieved IP",
				"provider", provider,
				"ip", ip)
			return ip
		} else {
			logger.Warn("Non-OK status code",
				"provider", provider,
				"status_code", resp.StatusCode)
		}
	}

	logger.Error("All IP providers failed")
	return ""
}
