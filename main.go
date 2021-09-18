package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"github.com/wagslane/go-rabbitmq"
)

var plexUrl string
var plexToken string
var rUrl string
var requestQueue = "video-request"

func main() {
	plexUrl = os.Getenv("PLEX_URL")
	plexToken = os.Getenv("PLEX_TOKEN")
	rUrl = os.Getenv("RABBITMQ_CONNECTION")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/library", getSections).Methods("GET")
	router.HandleFunc("/library/{section}/media", getSectionMedia).Methods("GET")
	router.HandleFunc("/media/{media}", getMediaMetadata).Methods("GET")
	router.HandleFunc("/media/{media}/download", queue).Methods("POST")
	router.Use(loggingMiddleware)

	port := "8080"
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	log.Printf("Starting queue consumer: %s", requestQueue)
	rConf := amqp.Config{}
	consumer, err := rabbitmq.NewConsumer(rUrl, rConf)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	err = consumer.StartConsuming(
		consumeDownload,
		requestQueue,
		[]string{requestQueue},
		rabbitmq.WithConsumeOptionsConcurrency(1),
		rabbitmq.WithConsumeOptionsQueueDurable)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Printf("Starting server on :%v", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: %s", r.RequestURI)

		next.ServeHTTP(w, r)

		log.Printf("Request complete: %s", r.RequestURI)
	})
}

func consumeDownload(d rabbitmq.Delivery) bool {
	log.Printf("Message Consumed: %v", string(d.Body))

	v := &Video{}
	err := json.Unmarshal(d.Body, v)
	if err != nil {
		log.Printf("Failed to unmarshal json to video")
	}

	rmt := fmt.Sprintf("/data/rmt/%s", v.Media.Parts[0].Path)
	local := fmt.Sprintf("/data/local/%s", v.Media.Parts[0].Path)
	log.Printf("Downloading from %v to %v", rmt, local)

	// Todo: Copy the files
	time.Sleep(time.Second * 10)
	return true
}

func getSections(w http.ResponseWriter, _ *http.Request) {
	u := fmt.Sprintf("%v/library/sections?X-Plex-Token=%v", plexUrl, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get library data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	l := LibraryContainer{}
	_ = xml.Unmarshal(body, &l)

	j, _ := json.Marshal(l)

	_, _ = w.Write(j)
}

func getSectionMedia(w http.ResponseWriter, r *http.Request) {
	section := mux.Vars(r)["section"]

	u := fmt.Sprintf("%v/library/sections/%v/all?X-Plex-Token=%v", plexUrl, section, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get library data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	c := VideoContainer{}
	_ = xml.Unmarshal(body, &c)

	j, _ := json.Marshal(c)

	_, _ = w.Write(j)
}

func getMediaMetadata(w http.ResponseWriter, r *http.Request) {
	media := mux.Vars(r)["media"]

	u := fmt.Sprintf("%v/library/metadata/%v?X-Plex-Token=%v", plexUrl, media, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get library data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	c := VideoContainer{}
	_ = xml.Unmarshal(body, &c)

	j, _ := json.Marshal(c.Videos[0])

	_, _ = w.Write(j)
}

func queue(w http.ResponseWriter, r *http.Request) {
	media := mux.Vars(r)["media"]

	u := fmt.Sprintf("%v/library/metadata/%v?X-Plex-Token=%v", plexUrl, media, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get media data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	vc := VideoContainer{}
	_ = xml.Unmarshal(body, &vc)

	v := vc.Videos[0]

	log.Printf("Triggering download of media (%v - %v - %v)", media, v.Title, v.Media.Parts[0].Path)

	c := amqp.Config{}
	publisher, returns, err := rabbitmq.NewPublisher(rUrl, c)
	if err != nil {
		log.Fatal(err)
	}

	j, _ := json.Marshal(v)
	err = publisher.Publish(j, []string{requestQueue})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for r := range returns {
			log.Printf("message returned from server: %s", string(r.Body))
		}
	}()

	_, _ = w.Write(j)
}
