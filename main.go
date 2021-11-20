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

var quizRektio = map[string][]string{
	"Verbi + P": {
		"etsiä - искать",
		"häiritä - тревожить",
		"inhota - призерать",
		"juhlia - праздновать",
		"kokeilla - попробовать",
		"käyda - посещать",
	},

	"Verbi + sta/stä": {
		"nauttia - наслаждаться",
		"olla kiinnostunut - быть заинтересованным",
		"pitää",
	},

	"Verbi + lta/ltä": {
		"haista - вонять",
		"kuulosta - звучать",
		"maistua - иметь вкус",
		"näyttää - выглядеть",
		"tuntua - чувствуется",
		"tuoksua - пахнуть",
		"vaikuttaa - производить впечатление",
	},

	"Verbi + ILL": {
		"ihastua - влюбляться",
		"osallistua - принять участие в чем-то",
		"rakastua - влюбляться",
		"tutustua - познакомиться",
	},

	"Verbi + maan/mään": {
		"oppia - научиться",
		"ruveta - приняться за что-то",
	},
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
	if quizType := rand.Intn(3); quizType == 1 { //nolint:gosec
		sendWordTypeQuiz(bot, message)
	} else {
		sendRektioQuiz(bot, message)
	}
}

func sendRektioQuiz(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	answers := make([]string, 0, len(quizRektio))
	for k := range quizRektio {
		answers = append(answers, k)
	}
	rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })

	rightAnswerIndex := rand.Intn(len(answers)) //nolint:gosec
	rightAnswer := answers[rightAnswerIndex]
	guess := quizRektio[rightAnswer][rand.Intn(len(quizRektio[rightAnswer]))] //nolint:gosec
	poll := tgbotapi.SendPollConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
		},
		Question:              fmt.Sprintf("Как употребляют глагол '%s' ?", guess),
		Options:               answers,
		IsAnonymous:           true,
		Type:                  "quiz",
		AllowsMultipleAnswers: false,
		CorrectOptionID:       int64(rightAnswerIndex),
		Explanation:           fmt.Sprintf("correct - %s", answers[rightAnswerIndex]),
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

func sendWordTypeQuiz(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
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
