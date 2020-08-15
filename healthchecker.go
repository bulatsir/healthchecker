package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

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
			checkerror(err)
			urlForCheck.Domain = url

			if resp.StatusCode == 200 {
				urlForCheck.Status = "up"
			} else {
				urlForCheck.Status = "down"
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

	go periodiccheck() //run ticker
	http.HandleFunc("/", healthcheck)
	http.ListenAndServe(":8090", nil)

}
