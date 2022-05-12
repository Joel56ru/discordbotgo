package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
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

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	go func() {
		updates := bot.GetUpdatesChan(u)

		for update := range updates {
			if update.Message != nil { // If we got a message
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
			}
		}
	}()
	dg.AddHandler(messageCreate)
	dg.AddHandler(eventCreate)
	//dg.Identify.Intents = discordgo.IntentGuildScheduledEvents
	//dg.Identify.Intents = discordgo.IntentsGuildMessages
	go func() {
		for range time.Tick(time.Minute * 10) {
			sendNews(dg)
		}
	}()

	go func() {
		for range time.Tick(time.Hour) {
			if time.Now().Hour() == 8 {
				text, err := calend()
				if err != nil {
					dg.ChannelMessageSend("964688607015239771", err.Error())
				}
				dg.ChannelMessageSend(ChannelNews, text)
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

	defer dg.Close()
}

func eventCreate(s *discordgo.Session, e *discordgo.GuildScheduledEventCreate) {
	if e.Name == "Test" {
		return
	}
	s.ChannelMessageSend("963482521146916867", `https://discord.gg/EFyjYbqn7E?event=`+e.ID)
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
	if m.Content == "календарь" {
		//s.ChannelMessageSend(m.ChannelID, "https://www.calend.ru/img/export/informer.png")
		a, err := calend()
		if err != nil {
			fmt.Println(err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, a)
	}
	reg := `(?i)Сколько по времени (\d+) сер[А-Яа-я]{2}\?`
	ref, _ := regexp.MatchString(reg, m.Content)
	if ref {
		d, _ := regexp.Compile(reg)
		sch, _ := strconv.Atoi(d.FindStringSubmatch(m.Content)[1])
		allMin := sch * 24
		hour := allMin / 60
		minute := allMin - hour*60
		future := time.Now()
		r := future.Add(time.Minute * time.Duration(allMin))
		var dHour, dMinute string
		if r.Hour() < 10 {
			dHour = `0` + strconv.Itoa(r.Hour())
		} else {
			dHour = strconv.Itoa(r.Hour())
		}
		if r.Minute() < 10 {
			dMinute = `0` + strconv.Itoa(r.Minute())
		} else {
			dMinute = strconv.Itoa(r.Minute())
		}
		zone, _ := r.Zone()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(`%d %s %d %s (1 серия 24 минуты). Если начать сейчас, то закончим в %s:%s %s`, hour, oconHours(hour), minute, oconMinutes(minute), dHour, dMinute, zone))
	}
}

func oconHours(i int) string {
	switch i {
	case 1, 21:
		return "час"
	case 2, 3, 4, 22, 23:
		return "часа"
	}
	return "часов"
}

func oconMinutes(i int) string {
	switch i {
	case 1, 21:
		return "минуту"
	case 2, 3, 4, 22, 23, 24:
		return "минуты"
	}
	return "минут"
}

func calend() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://kakoysegodnyaprazdnik.ru/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")
	res, err := client.Do(req)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return res.Status, nil
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	var result string
	count := 0
	doc.Find("#main_frame .listing .listing_wr .main").Each(func(i int, s *goquery.Selection) {
		if count < 20 {
			text := s.Find("span").First().Text()
			text = strings.Replace(text, "США", ":flag_um:", 1)
			text = strings.Replace(text, "Япония", ":flag_jp:", 1)
			result += ":small_blue_diamond: " + text + "\n"
		}
		count++
	})
	result = "**Праздники сегодня**\n" + result
	return result, nil
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
