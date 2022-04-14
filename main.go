package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	Token       string
	TokenShiki  string
	ChannelNews string
	LastID      int
)

func main() {
	ChannelNews = os.Getenv("CHANNELNEWS")
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}
	Token = os.Getenv("DGU_TOKEN")
	TokenShiki = os.Getenv("SHIKI_ACCESS_TOKEN")

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	ticker := time.NewTicker(time.Minute * 10)

	go func() {
		for {
			select {

			case t := <-ticker.C:
				sendNews(dg)
				fmt.Println("Tick at", t)
			}
		}
	}()
	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
func messageDELETE(s *discordgo.Session, m *discordgo.MessageDelete) {

	s.ChannelMessageSend(m.ChannelID, "Удалил сообщение Ник: **fdf**")
	log.Println(m.BeforeDelete.Author.Username)

}

func sendNews(s *discordgo.Session) {
	var res []Topic
	err := ShikiGetTopics(TokenShiki, &res)
	if err != nil {
		logrus.Fatal(err)
	}
	if LastID == res[0].Id {
		return
	}
	LastID = res[0].Id
	embed := discordgo.MessageEmbed{
		URL:         `https://shikimori.one` + res[0].Forum.Url + "/" + strconv.Itoa(res[0].Id),
		Type:        "rich",
		Title:       res[0].TopicTitle,
		Description: res[0].Body,
		Timestamp:   res[0].CreatedAt,
		Color:       123222,
		Footer: &discordgo.MessageEmbedFooter{
			Text: res[0].Forum.Name,
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "https://kawai.shikimori.one" + res[0].Linked.Image.Original,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://kawai.shikimori.one" + res[0].Linked.Image.Preview,
		},
		Video:    nil,
		Provider: nil,
		Author:   nil,
		Fields:   nil,
	}
	s.ChannelMessageSendEmbed(ChannelNews, &embed)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelTyping(m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == "новости тут будут" {
		ChannelNews = m.ChannelID
	}
}
