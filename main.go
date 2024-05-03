package main

import (
	"flag"
	"io"
	"log"
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
	if BotToken == nil || *BotToken == "" {
		log.Fatal("missing flag --token")
	}

	dg, err := discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}
	defer dg.Close()

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

	})

	dg.AddHandler(func(s *discordgo.Session, gc *discordgo.GuildCreate) {
		dg.ApplicationCommandCreate(dg.State.User.ID, gc.ID, &command)
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name == "get-ip" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: getCurrentIP(),
				},
			})
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening Discord connection:", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}

func getCurrentIP() string {
	for _, provider := range ipProviders {
		resp, err := http.Get(provider)
		if err != nil {
			log.Printf("Error getting IP from %s: %v\n", provider, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body from %s: %v\n", provider, err)
				continue
			}
			return string(body)
		}
	}
	return ""
}
