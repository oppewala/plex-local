package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/oppewala/plex-local-dl/pkg/plex"
	"github.com/oppewala/plex-local-dl/pkg/storage"
	"github.com/rs/cors"
)

var (
	plexUrl    string
	plexToken  string
	plexServer *plex.Server
	store      *storage.Storage
	hub        *Hub
	mediaPath  string
)

func main() {
	var port string
	var wait time.Duration
	var storageConnectionString string
	flag.StringVar(&plexUrl, "plexUrl", os.Getenv("PLEX_URL"), "the token for the source plex server - can be set through environment variable PLEX_URL")
	flag.StringVar(&plexToken, "plexToken", os.Getenv("PLEX_TOKEN"), "the url for the source plex server - can be set through environment variable PLEX_TOKEN")
	flag.StringVar(&storageConnectionString, "storageConnection", os.Getenv("AZURE_STORAGE"), "the connection string to the storage account - can be set through environment variable AZURE_STORAGE")
	flag.StringVar(&port, "port", "8080", "the port to run the UI on - e.g. 8080 (optional)")
	flag.StringVar(&mediaPath, "mediaPath", "/data/media", "the directory to download media to")
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m (optional)")
	flag.Parse()

	if plexUrl == "" || plexToken == "" {
		log.Fatal("Plex Token and URL must be provided as an environment variable or command line argument")
	}

	plexServer = plex.NewServer(plexUrl, plexToken)
	go func() {
		for {
			err := populateTitles()
			if err != nil {
				log.Printf("[Main] Failed to populate titles: %v", err)
			}

			// TODO: Refresh when page is loaded to reduce load when not in active use
			time.Sleep(time.Minute * 30)
		}
	}()

	store = storage.ConnectStorage(storageConnectionString)

	hub = newHub()
	go hub.run()
	go chanConsumer(hub)
	go func() {
		for {
			hub.broadcast <- Ping{}
			time.Sleep(time.Second)
		}
	}()

	newQueueConsumer(storageConnectionString)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/library", getLibraries).Methods(http.MethodGet)
	router.HandleFunc("/api/library/{key:[0-9]+}/media", getLibraryContent).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key:[0-9]+}", getMediaMetadata).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key:[0-9]+}/parts", getMediaParts).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key:[0-9]+}/download", postQueue).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/media/download/persist", getPersisted).Methods(http.MethodGet)
	router.HandleFunc("/api/media/{key:[0-9]+}/download/persist", deletePersist).Methods(http.MethodDelete)
	router.HandleFunc("/api/media/{key:[0-9]+}/download/persist", postPersist).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/media/download/persist/{partition}/{row}", deletePersistForce).Methods(http.MethodDelete)
	router.HandleFunc("/api/search", getSearch).Queries("q", "{query}").Methods(http.MethodGet)
	router.HandleFunc("/api/ws", func(writer http.ResponseWriter, request *http.Request) {
		ws(writer, request, hub)
	})
	router.Use(loggingMiddleware)
	_ = router.Walk(printRoutes)

	co := cors.AllowAll()

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      co.Handler(router),
	}

	log.Printf("[Main] Starting server on :%v", port)
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
	log.Println("[Main] Shutting down")
	os.Exit(0)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] --> %s %s", r.Method, r.URL.Path)

		if r.URL.Path == "/api/ws" {
			log.Printf("[API] Converting to WS connection")
			next.ServeHTTP(w, r)
			return
		}

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		log.Printf("[API] <-- %s %s [%d %s]", r.Method, r.URL.Path, lrw.statusCode, http.StatusText(lrw.statusCode))
	})
}

func printRoutes(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
	pathTemplate, err := route.GetPathTemplate()
	if err == nil {
		fmt.Println("ROUTE:", pathTemplate)
	}
	pathRegexp, err := route.GetPathRegexp()
	if err == nil {
		fmt.Println("Path regexp:", pathRegexp)
	}
	queriesTemplates, err := route.GetQueriesTemplates()
	if err == nil {
		fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
	}
	queriesRegexps, err := route.GetQueriesRegexp()
	if err == nil {
		fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
	}
	methods, err := route.GetMethods()
	if err == nil {
		fmt.Println("Methods:", strings.Join(methods, ","))
	}
	fmt.Println()
	return nil
}
