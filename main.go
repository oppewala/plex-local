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
	"github.com/oppewala/plex-local-dl/pkg/plex"
	"github.com/rs/cors"
)

var (
	plexUrl   string
	plexToken string
	s         *plex.Server
	hub       *Hub
	mediaPath string
)

func main() {
	var port string
	var wait time.Duration
	flag.StringVar(&plexUrl, "plexUrl", os.Getenv("PLEX_URL"), "the token for the source plex server - can be set through environment variable PLEX_URL")
	flag.StringVar(&plexToken, "plexToken", os.Getenv("PLEX_TOKEN"), "the url for the source plex server - can be set through environment variable PLEX_TOKEN")
	flag.StringVar(&port, "port", "8080", "the port to run the UI on - e.g. 8080 (optional)")
	flag.StringVar(&mediaPath, "mediaPath", "/data/media", "the directory to download media to")
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m (optional)")
	flag.Parse()

	if plexUrl == "" || plexToken == "" {
		log.Fatal("Plex Token and URL must be provided as an environment variable or command line argument")
	}

	go func() {
		err := populateTitles()
		if err != nil {
			log.Printf("failed to populate titles: %v", err)
		}
	}()

	s = plex.NewServer(plexUrl, plexToken)

	hub = newHub()
	go hub.run()
	go chanConsumer(hub)
	go func() {
		for {
			hub.broadcast <- Ping{}
			time.Sleep(time.Second)
		}
	}()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/library", getLibraries).Methods(http.MethodGet)
	router.HandleFunc("/api/library/{key}/media", getLibraryContent).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key}", getMediaMetadata).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key}/parts", getMediaParts).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key}/download", postQueue).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/search", getSearch).Queries("q", "{query}").Methods(http.MethodGet)
	router.HandleFunc("/api/ws", func(writer http.ResponseWriter, request *http.Request) {
		ws(writer, request, hub)
	})
	router.Use(loggingMiddleware)

	co := cors.AllowAll()

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      co.Handler(router),
	}

	log.Printf("Starting server on :%v", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
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
		log.Printf("[API] Request received: %s", r.RequestURI)

		next.ServeHTTP(w, r)

		log.Printf("[API] Request complete: %s")
	})
}
