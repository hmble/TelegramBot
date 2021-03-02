package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/joho/godotenv/autoload"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	// WordTemplate is a const path word template
	WordTemplate = "./word.tmpl"
)
var t *template.Template
var startTemplate *template.Template
func init() {
	 t = template.Must(template.ParseFiles(WordTemplate))
}
// MyGroup is personal group struct to implement tb.Recipient
type MyGroup struct {
	Name string
	ChatID string
}

// Recipient returns personal group chatID
func (mg MyGroup) Recipient() string{
	return mg.ChatID
}
func main() {

	startBot()
	// fmt.Println(getDefinition("whim"))
}

func startBot() {
	mg := &MyGroup{Name: "PersonalBot", ChatID: os.Getenv("GROUP_CHAT_ID")}
	b, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
	
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	
	if err != nil {
		log.Fatal(err)
		return
	}
	
	b.Handle(tb.OnAddedToGroup, func(m *tb.Message) {
		fmt.Println("Chat ID of Group is: ", m.Chat.ID)
		b.Send(mg, "I got invited to ", m.Chat.Username)
	})
	b.Handle("/hello", func(m *tb.Message) {
		if strconv.Itoa(int(m.Chat.ID)) != mg.ChatID {
			b.Send(m.Chat, "You are not allowed")
			return
		}
		b.Send(m.Chat, fmt.Sprintf("Hello %s", m.Sender.FirstName))
	})
	
	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Chat, fmt.Sprintf(
			`Hello %s
			I am SamPersonalBot
			Use following commands to get help from me
			/help		To get help from me
			/w			To get word definition
			`, m.Sender.FirstName))
	})
	b.Handle("/help", func(m *tb.Message) {
		b.Send(m.Chat, fmt.Sprintf(
			`Hello %s
			I am SamPersonalBot
			Use following commands to get help from me
			/help		To get help from me ( This Message )
			/w			To get word definition
			`, m.Sender.FirstName))
	})
	b.Handle("/w", func(m *tb.Message) {
		if strconv.Itoa(int(m.Chat.ID)) != mg.ChatID {
			b.Send(m.Chat, "You are not allowed")
			return
		}
		b.Send(m.Chat, getDefinition(m.Payload))
	})
	
	b.Start()

}
type meaning struct {
	MeaningType string
	Meaning string
}
type wordDetails struct {
	Word string
	Meaning []string
}
func getDefinition(word string) string{
  // Request the HTML page.
  res, err := http.Get("https://vocabulary.com/dictionary/" + word)
  if err != nil {
    log.Fatal(err)
  }
  defer res.Body.Close()
  if res.StatusCode != 200 {
    log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
  }

  // Load the HTML document
  doc, err := goquery.NewDocumentFromReader(res.Body)
  if err != nil {
    log.Fatal(err)
  }

  // Find the review items
	meanArray := make([]string, 0)

	doc.Find(".short").Each(func(i int, s *goquery.Selection) {
		meanArray = append(meanArray, strings.TrimSpace(s.Text()))
	})
  doc.Find(".word-definitions > ol > li .definition:first-child").Each(func(i int, s *goquery.Selection) {
		means := &meaning{}
		s.Children().Each( func(i int, s *goquery.Selection) {
			means.MeaningType = strings.TrimSpace(s.Text())
		})
		s.Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			means.Meaning = strings.TrimSpace(s.Text())
		}

	})
	meanArray = append(meanArray, fmt.Sprintf("%s : %s", means.MeaningType, means.Meaning))
})
return process(wordDetails{Word: word, Meaning: meanArray})
}
// process applies the data structure 'vars' onto an already
// parsed template 't', and returns the resulting string.
func process(vars interface{}) string {
    var tmplBytes bytes.Buffer

    err := t.Execute(&tmplBytes, vars)
    if err != nil {
        panic(err)
    }

    return tmplBytes.String()
}