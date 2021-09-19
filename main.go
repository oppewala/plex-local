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
)

var (
	plexUrl      = os.Getenv("PLEX_URL")
	plexToken    = os.Getenv("PLEX_TOKEN")
	requestQueue = make(chan *Video)
)

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/library", getLibraries).Methods("GET")
	router.HandleFunc("/library/{key}/media", getLibraryContent).Methods("GET")
	router.HandleFunc("/media/{key}", getMediaMetadata).Methods("GET")
	router.HandleFunc("/media/{key}/download", postQueue).Methods("POST")
	router.HandleFunc("/search", getSearch).Queries("q", "{query}").Methods("GET")
	router.Use(loggingMiddleware)

	port := "8080"
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	log.Printf("Starting queue consumer")
	go chanConsumer()

	log.Printf("Starting server on :%v", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		populateTitles()
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
