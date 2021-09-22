package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getLibraries(w http.ResponseWriter, _ *http.Request) {
	l, err := s.GetLibraries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(l)

	_, _ = w.Write(j)
}

func getLibraryContent(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	c, err := s.GetLibraryContent(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(c)
	_, _ = w.Write(j)
}

func getMediaMetadata(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	v, err := s.GetMediaMetadata(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(v)
	_, _ = w.Write(j)
}

func postQueue(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	m, err := s.GetMediaMetadata(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parts, err := s.GetMediaParts(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l := fmt.Sprintf("Queuing download of media (%s - %s)", k, m.ConcatTitles())
	log.Printf(l)

	for _, part := range parts {
		go func() {
			requestQueue <- DownloadRequest{
				Metadata: m,
				Part:     part,
			}
		}()
	}

	j, _ := json.Marshal(struct {
		log string
	}{
		log: l,
	})
	_, _ = w.Write(j)
}

func getSearch(w http.ResponseWriter, r *http.Request) {
	q := mux.Vars(r)["query"]
	results := search(q)

	j, _ := json.Marshal(results)
	_, _ = w.Write(j)
}

func getMediaParts(w http.ResponseWriter, r *http.Request) {
	sk := mux.Vars(r)["key"]

	p, err := s.GetMediaParts(sk)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(p)
	_, _ = w.Write(j)

}
