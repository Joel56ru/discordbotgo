package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Topic struct {
	Id         int    `json:"id"`
	TopicTitle string `json:"topic_title"`
	Body       string `json:"body"`
	Forum      Forum  `json:"forum"`
	CreatedAt  string `json:"created_at"`
	Linked     Linked `json:"linked"`
}
type Forum struct {
	Id        int    `json:"id"`
	Position  int    `json:"position"`
	Name      string `json:"name"`
	Permalink string `json:"permalink"`
	Url       string `json:"url"`
}
type Linked struct {
	Id    int   `json:"id"`
	Image Image `json:"image"`
}
type Image struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
}

var (
	Token       string
	ChannelNews string
	MainConfig  AppConfig
)

type AppConfig struct {
	LastID int `json:"last_id"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}
	ChannelNews = os.Getenv("CHANNELNEWS")
	Token = os.Getenv("DGU_TOKEN")

	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		MainConfig.LastID, _ = strconv.Atoi(os.Getenv("LASTID"))
		content, err := json.Marshal(MainConfig)
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile("config.json", content, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	content, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &MainConfig)
	if err != nil {
		log.Fatal(err)
	}
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	ticker := time.NewTicker(time.Minute * 10)

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				sendNews(dg)
				//fmt.Println("Tick at", t)
			}
		}
	}()

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func sendNews(s *discordgo.Session) {
	var res []Topic
	err := ShikiGetTopics(&res)
	if err != nil {
		log.Println(err)
	}
	if MainConfig.LastID == res[0].Id {
		return
	}
	MainConfig.LastID = res[0].Id
	content, err := json.Marshal(MainConfig)
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile("config.json", content, 0644)
	if err != nil {
		log.Fatal(err)
	}
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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "жив?" {
		s.ChannelTyping(m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "жив!")
	}
	if m.Content == "новости тут будут" {
		ChannelNews = m.ChannelID
		os.Setenv("CHANNELNEWS", m.ChannelID)
	}
}

func ShikiGetTopics(target interface{}) error {
	spaceClient := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, `https://shikimori.one/api/topics?limit=1&forum=news&page=1`, nil)
	if err != nil {
		return err
	}
	res, err := spaceClient.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}
