package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
)

var Config Conf

type Conf struct {
	Key    string `yaml:"key"`
	ChatId string `yaml:"chat_id"`
}

func sendMessage(s string, reply string) {
	bot, err := tgbotapi.NewBotAPI(Config.Key)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	log.Printf("send message to %s", s)
	chatId, _ := strconv.ParseInt(s, 10, 64)
	msg := tgbotapi.NewMessage(chatId, reply)

	bot.Send(msg)
}

//error handler

func checkerror(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

//run check resource periodically
func periodiccheck() {

	logfile, err := os.OpenFile("check.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	checkerror(err)

	tick := time.NewTicker(time.Second * 15)
	for _ = range tick.C {
		file, err := os.Open("uri.conf")
		checkerror(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		urlForCheck := struct {
			Domain string
			Status string
		}{}

		for scanner.Scan() {
			url := scanner.Text()
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error:", err)
			}
			urlForCheck.Domain = url

			if resp.StatusCode == 200 {
				urlForCheck.Status = "up"

			} else {
				urlForCheck.Status = "down"
				alertMessage := urlForCheck.Domain + " is down"
				sendMessage(Config.ChatId, alertMessage)

			}

			log.SetOutput(logfile)
			log.Println(urlForCheck.Domain, urlForCheck.Status)
		}
	}

}

//check resource from uri.conf and do http response

func healthcheck(w http.ResponseWriter, req *http.Request) {

	file, err := os.Open("uri.conf")
	checkerror(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	urlForCheck := struct {
		Domain string
		Status string
	}{}

	t, err := template.ParseFiles("status.html")
	checkerror(err)

	for scanner.Scan() {
		url := scanner.Text()
		resp, err := http.Get(url)
		checkerror(err)
		urlForCheck.Domain = url

		if resp.StatusCode == 200 {
			urlForCheck.Status = "up"
		} else {
			urlForCheck.Status = "down"
		}

		t.Execute(w, urlForCheck)

	}

}

func main() {

	//read configuration

	yamlFile, err := ioutil.ReadFile("key.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	go periodiccheck() //run ticker
	http.HandleFunc("/", healthcheck)
	http.ListenAndServe(":8090", nil)

}
