package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	isesnsorPath = "https://isensor.herokuapp.com/data?limit="
)

type Sensor []struct {
	ID          int       `json:"id"`
	Temperature int       `json:"temperature"`
	Humidity    int       `json:"humidity"`
	Pressure    int       `json:"pressure"`
	Uv          int       `json:"uv"`
	Date        time.Time `json:"date"`
}

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Sensor data🌡️"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Gen Chart"),
	),
)
var graphKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("5"),
		tgbotapi.NewKeyboardButton("10"),
		tgbotapi.NewKeyboardButton("15"),
	),
	tgbotapi.NewKeyboardButtonRow(
		//TODO: add "back" emoji
		tgbotapi.NewKeyboardButton("Back"),
	),
)

func main() {
	//file := tgbotapi.FilePath("chart.png")
	s := Sensor{}
	bot, err := tgbotapi.NewBotAPI(mustToken())
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		switch update.Message.Text {
		case "/start":
			msg.Text = helloMessage()
			msg.ReplyMarkup = mainKeyboard
		case "/open":
			msg.ReplyMarkup = mainKeyboard
		case "/close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		case "Sensor data🌡️":
			msg.Text = s.sensorData()
		case "Gen Chart":
			msg.Text = "Choose number of iterations"
			msg.ReplyMarkup = graphKeyboard
		case "Back":
			msg.ReplyMarkup = mainKeyboard
		case "5":
			m, _ := s.genChart(5)
			photoBytes, err := ioutil.ReadFile(m)
			if err != nil {
				panic(err)
			}
			photoFileBytes := tgbotapi.FileBytes{Name: "picture", Bytes: photoBytes}
			chatID := update.Message.Chat.ID
			_, err = bot.Send(tgbotapi.NewPhoto(int64(chatID), photoFileBytes))
		case "10":
			m, _ := s.genChart(10)
			photoBytes, err := ioutil.ReadFile(m)
			if err != nil {
				panic(err)
			}
			photoFileBytes := tgbotapi.FileBytes{Name: "picture", Bytes: photoBytes}
			chatID := update.Message.Chat.ID
			_, err = bot.Send(tgbotapi.NewPhoto(int64(chatID), photoFileBytes))
		case "15":
			m, _ := s.genChart(15)
			photoBytes, err := ioutil.ReadFile(m)
			if err != nil {
				panic(err)
			}
			photoFileBytes := tgbotapi.FileBytes{Name: "picture", Bytes: photoBytes}
			chatID := update.Message.Chat.ID
			_, err = bot.Send(tgbotapi.NewPhoto(int64(chatID), photoFileBytes))
		}
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

//parses the flags and return the token
func mustToken() string {
	token := flag.String("token", "", "token for access to Telegram bot")

	flag.Parse()
	if *token == "" {
		log.Fatal("token is not specified")
	}
	return *token
}

//return bot start message as a string
func helloMessage() string {
	message := "Hello!\nI'm a isesnsor Telegram bot. Press /open to open the keyboard."
	return message
}

//request to isensor API and parse to struct
func (s *Sensor) sensorResponse(limit int) error {
	link := isesnsorPath + strconv.Itoa(limit)
	resp, err := http.Get(link)
	if err != nil {
		return fmt.Errorf("can't do request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't read request")
	}

	err = json.Unmarshal(body, &s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON")
	}
	return nil
}

func (s *Sensor) sensorData() string {
	var res string
	err := s.sensorResponse(1)
	if err != nil {
		res = "Can't do request"
	}
	res = "🌡️" + strconv.Itoa((*s)[0].Temperature) + " °C\n" +
		"💧" + strconv.Itoa((*s)[0].Humidity) + " %\n" +
		"🌎" + strconv.Itoa((*s)[0].Pressure) + " Pa\n" +
		"☀️" + strconv.Itoa((*s)[0].Uv) + " W/m²"
	return res
}
