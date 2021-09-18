package main

import (
	"context"
	"flag"
	"fmt"
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
	var err error
	for i := 1; i < 5; i++ {
		err = startConsumer()
		if err != nil && i != 4 {
			log.Printf("Failed to start, retrying: %v", err)

			time.Sleep(time.Second * 5)
		}
	}
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
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

func startConsumer() error {
	rConf := amqp.Config{}
	consumer, err := rabbitmq.NewConsumer(rUrl, rConf)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %v", err)
	}

	err = consumer.StartConsuming(
		consumeDownload,
		requestQueue,
		[]string{requestQueue},
		rabbitmq.WithConsumeOptionsConcurrency(1),
		rabbitmq.WithConsumeOptionsQueueDurable)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %v", err)
	}

	return nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: %s", r.RequestURI)

		next.ServeHTTP(w, r)

		log.Printf("Request complete: %s", r.RequestURI)
	})
}
