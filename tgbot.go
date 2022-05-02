package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wcharczuk/go-chart"
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
		tgbotapi.NewKeyboardButton("🌡️Sensor data"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📈Charts"),
	),
)
var graphKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("5"),
		tgbotapi.NewKeyboardButton("10"),
		tgbotapi.NewKeyboardButton("15"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🔙Back"),
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
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		var timeRange string //use for display time at charts
		var filePath string  //usr as path to render chart file

		switch update.Message.Text {
		case "/start":
			msg.Text = helloMessage()
			msg.ReplyMarkup = mainKeyboard
		case "/help":
			msg.Text = helpMessage()
		case "/open":
			msg.ReplyMarkup = mainKeyboard
		case "/close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		case "🌡️Sensor data":
			msg.Text = s.sensorData()
		case "📈Charts":
			msg.Text = "Choose number of iterations"
			msg.ReplyMarkup = graphKeyboard
		case "🔙Back":
			msg.Text = "Choose method"
			msg.ReplyMarkup = mainKeyboard
		case "5":
			timeRange, filePath, err = s.genChart(5)
			msg.Text = timeRange
			photoBytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			photoFileBytes := tgbotapi.FileBytes{Name: "picture", Bytes: photoBytes}
			chatID := update.Message.Chat.ID
			_, err = bot.Send(tgbotapi.NewPhoto(int64(chatID), photoFileBytes))
		case "10":
			timeRange, filePath, err = s.genChart(10)
			msg.Text = timeRange
			photoBytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			photoFileBytes := tgbotapi.FileBytes{Name: "picture", Bytes: photoBytes}
			chatID := update.Message.Chat.ID
			_, err = bot.Send(tgbotapi.NewPhoto(int64(chatID), photoFileBytes))
		case "15":
			timeRange, filePath, err = s.genChart(15)
			msg.Text = timeRange
			photoBytes, err := ioutil.ReadFile(filePath)
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

//return bot START message as a string
func helloMessage() string {
	message := "Hello!👋\nI'm a isesnsor telegram bot.🤖 I can show current information from sensors or build charts from lots of data.\nPrint /help for more information."
	return message
}

//return bot HELP message as a string
func helpMessage() string {
	message := "This bot was writen in go using open source libraries:\n🤖github.com/go-telegram-bot-api/telegram-bot-api\n📈github.com/wcharczuk/go-chart\n\nIsensor project: github.com/i-sensor"
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

	//adapt time
	local := (*s)[0].Date
	location, err := time.LoadLocation("Europe/Budapest") //similar to Kyiv (almost^_^)
	if err == nil {
		local = local.In(location)
	}

	res = "🌡️" + strconv.Itoa((*s)[0].Temperature) + " °C\n" +
		"💧" + strconv.Itoa((*s)[0].Humidity) + " %\n" +
		"🌎" + strconv.Itoa((*s)[0].Pressure) + " Pa\n" +
		"☀️" + strconv.Itoa((*s)[0].Uv) + " W/m²\n" +
		"(last update: " + local.Format("02 Jan 06 at 15:04") + ")"

	return res
}
func (s *Sensor) genChart(iterations int) (string, string, error) {
	err := s.sensorResponse(iterations)
	if err != nil {
		fmt.Errorf("Request error")
	}

	tX_Values, tY_Values := s.temperatureData(iterations)
	hX_Values, hY_Values := s.humidityData(iterations)
	pX_Values, pY_Values := s.pressureData(iterations)
	uX_Values, uY_Values := s.uvData(iterations)

	graph := chart.Chart{
		Series: []chart.Series{

			chart.ContinuousSeries{
				XValues: tX_Values,
				YValues: tY_Values,
			},
			chart.ContinuousSeries{
				XValues: pX_Values,
				YValues: pY_Values,
			},
			chart.ContinuousSeries{
				XValues: hX_Values,
				YValues: hY_Values,
			},
			chart.ContinuousSeries{
				XValues: uX_Values,
				YValues: uY_Values,
			},
		},
	}

	filename := "chart.png"
	f, err := os.Create(filename)
	if err != nil {
		fmt.Errorf("Failed to create file: %v: %v", filename, err)
		return "", "", err
	}

	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		fmt.Errorf("Unable to render graph: %v", err)
		return "", "", err
	}
	message := s.timeRange(iterations)
	return message, filename, nil
}

func (s *Sensor) timeRange(iterations int) (res string) {
	//adapt time
	from := (*s)[0].Date
	to := (*s)[iterations-1].Date
	location, err := time.LoadLocation("Europe/Budapest") //similar to Kyiv (almost^_^)
	if err == nil {
		from = from.In(location)
		to = to.In(location)
	}
	res = "Last " + strconv.Itoa(iterations) + " updats\n" +
		"From " + from.Format("02 Jan 06 15:04") +
		" to " + to.Format("02 Jan 06 15:04")

	return
}

func (s *Sensor) temperatureData(iterations int) (x, y []float64) {
	x = make([]float64, iterations)
	for i := range x {
		x[i] = float64(i)
	}
	y = make([]float64, iterations)
	for i := range y {
		y[i] = float64((*s)[i].Temperature)
	}
	return
}

func (s *Sensor) humidityData(iterations int) (x, y []float64) {
	x = make([]float64, iterations)
	for i := range x {
		x[i] = float64(i)
	}
	y = make([]float64, iterations)
	for i := range y {
		y[i] = float64((*s)[i].Humidity)
	}
	return
}

func (s *Sensor) pressureData(iterations int) (x, y []float64) {
	x = make([]float64, iterations)
	for i := range x {
		x[i] = float64(i)
	}
	y = make([]float64, iterations)
	for i := range y {
		y[i] = float64((*s)[i].Pressure / 10)
	}
	return
}

func (s *Sensor) uvData(iterations int) (x, y []float64) {
	x = make([]float64, iterations)
	for i := range x {
		x[i] = float64(i)
	}
	y = make([]float64, iterations)
	for i := range y {
		y[i] = float64((*s)[i].Uv * 10)
	}
	return
}
