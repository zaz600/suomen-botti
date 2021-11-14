package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var botToken = os.Getenv("SUOMEN_BOTTI_TG_TOKEN")

type SearchItem struct {
	Name        string
	CSSSelector string
}

type SearchResult struct {
	Name  string
	Value string
}

var cfg = []SearchItem{
	{Name: "G.yks", CSSSelector: "y.gen"},
	{Name: "P.yks", CSSSelector: "y.part"},
	{Name: "ILL.yks", CSSSelector: "y.ill"},

	{Name: "N.mon", CSSSelector: "mon.akk"},
	{Name: "G.mon", CSSSelector: "mon.gen"},
	{Name: "P.mon", CSSSelector: "mon.part"},
	{Name: "ILL.mon", CSSSelector: "mon.ill"},
}

var quizWords = []string{
	"banaani",
	"kerros",
	"kauppa",
	"pallo",
	"kala",
	"maa",
	"luu",
	"perhe",
	"vene",
	"lentokone",
	"koe",
	"lääke",
	"punainen",
	"ihminen",
	"pankki",
	"koti",
	"tori",
	"lasi",
	"lasi",
	"kieli",
	"meri",
	"vuori",
	"lehti",
	"järvi",
	"joki",
	"vuosi",
	"kuukausi",
	"vesi",
	"susi",
	"puhelin",
	"soitin",
	"avain",
	"morsian",
	"lounas",
	"hidas",
	"asiakas",
	"asiakas",
	"ananas",
	"vastaus",
	"harjoitus",
	"kerros",
	"ostos",
	"rikos",
	"vihannes",
	"veljes",
	"mies",
	"juures",
	"ystävyys",
	"rakkaus",
	"rakkaus",
	"korkeus",
	"talous",
	"kauneus",
	"olut",
	"kuollut",
	"manner",
	"tytär",
	"tanssijatar",
	"työtön",
	"koditon",
	"opiskelija",
	"kynttilä",
	"astia",
	"makkara",
	"ravintola",
	"omena",
}

func isEmptyCommand(message *tgbotapi.Message) bool {
	return message.IsCommand() && len(strings.Fields(message.Text)) == 1
}

func processSearchCmd(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	if isEmptyCommand(message) {
		answer := "Поискать формы слова. Пример:\n/search kerros "
		msg := tgbotapi.NewMessage(message.Chat.ID, answer)
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
		return
	}

	word := strings.Fields(strings.ToLower(message.Text))[1]

	answer := strings.Builder{}
	answer.WriteString(fmt.Sprintf("🔍 %s\n\n", word))

	items, err := getTaivutus(word)
	if err != nil {
		answer.WriteString("search error")
	} else {
		for _, result := range items {
			answer.WriteString(fmt.Sprintf("🔻 %s: %s\n", result.Name, result.Value))
		}
	}
	answer.WriteString(fmt.Sprintf("\n📖 https://fi.wiktionary.org/wiki/%s", word))
	msg := tgbotapi.NewMessage(message.Chat.ID, answer.String())
	msg.ReplyToMessageID = message.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func getTaivutus(word string) ([]SearchResult, error) {
	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		Get("https://fi.wiktionary.org/wiki/" + word)

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	var result []SearchResult

	for _, searchItem := range cfg {
		queryResult := doc.Find(fmt.Sprintf(`span[data-kuvaus*="%s"] a`, searchItem.CSSSelector))

		if len(queryResult.Nodes) > 0 {
			result = append(result, SearchResult{
				Name:  searchItem.Name,
				Value: queryResult.First().Text(),
			})
		} else {
			result = append(result, SearchResult{
				Name:  searchItem.Name,
				Value: "???",
			})
		}
	}
	return result, nil
}

func processQuizCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	word := quizWords[rand.Intn(len(quizWords))] //nolint:gosec
	items, err := getTaivutus(word)
	if err != nil {
		answer := "Не смогли подготовить данные для квиза 😭"
		msg := tgbotapi.NewMessage(message.Chat.ID, answer)
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
		return
	}

	rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })

	rightAnswer := rand.Intn(len(items)) //nolint:gosec
	answers := make([]string, 0, len(items))
	for _, item := range items {
		answers = append(answers, item.Value)
	}

	poll := tgbotapi.SendPollConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
		},
		Question:              fmt.Sprintf("Выберите %s для слова %s", items[rightAnswer].Name, word),
		Options:               answers,
		IsAnonymous:           true,
		Type:                  "quiz",
		AllowsMultipleAnswers: false,
		CorrectOptionID:       int64(rightAnswer),
		Explanation:           fmt.Sprintf("correct [answer](https://fi.wiktionary.org/wiki/%s): %s", word, items[rightAnswer].Value),
		ExplanationParseMode:  "markdown",
		OpenPeriod:            0,
		CloseDate:             0,
		IsClosed:              false,
	}
	poll.ReplyToMessageID = message.MessageID

	if _, err := bot.Send(poll); err != nil {
		log.Println(err)
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			log.Println(update)
			continue
		}

		log.Printf("user:%d [%s], chatId: %d, %s", update.Message.From.ID, update.Message.From.UserName, update.Message.Chat.ID, update.Message.Text)

		if update.Message.Command() == "search" {
			processSearchCmd(bot, update.Message)
		}

		if update.Message.Command() == "quiz" {
			processQuizCommand(bot, update.Message)
		}
	}
}
