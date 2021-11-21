package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/oppewala/plex-local-dl/pkg/plex"
	"github.com/oppewala/plex-local-dl/pkg/storage"
)

type apiPostResponse struct {
	Message string
}

func getLibraries(w http.ResponseWriter, _ *http.Request) {
	l, err := plexServer.GetLibraries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(l)
	_, _ = w.Write(j)
}

func getLibraryContent(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	c, err := plexServer.GetLibraryContent(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(c)
	_, _ = w.Write(j)
}

func getMediaMetadata(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	v, err := plexServer.GetMediaMetadata(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(v)
	_, _ = w.Write(j)
}

func postQueue(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	meta, err := plexServer.GetMetadataWithParts(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, m := range meta {
		log.Printf("[API] Queuing download of media (%s - %s)", k, m.ConcatTitles())

		queueDownload(m)
	}

	j, _ := json.Marshal(apiPostResponse{Message: "Download queued"})
	_, _ = w.Write(j)
}

func queueDownload(m plex.Metadata) {
	hub.broadcast <- &DownloadUpdate{
		MessageType:     "download-start",
		Title:           m.ConcatTitles(),
		BytesDownloaded: 0,
		TotalBytes:      0,
	}

	go func(m plex.Metadata) {
		requestQueue <- DownloadRequest{
			Metadata: m,
			Part:     m.Media[0].Part[0],
		}
	}(m)
}

func getSearch(w http.ResponseWriter, r *http.Request) {
	q := mux.Vars(r)["query"]
	results := search(q)

	j, _ := json.Marshal(results)
	_, _ = w.Write(j)
}

func getMediaParts(w http.ResponseWriter, r *http.Request) {
	sk := mux.Vars(r)["key"]

	p, err := plexServer.GetMetadataWithParts(sk)
	if err != nil {
		log.Printf("[API] Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(p)
	_, _ = w.Write(j)
}

func postPersist(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	m, err := plexServer.GetMediaMetadata(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tvdbid, err := plexServer.GetDbId(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plexId, _ := strconv.ParseUint(k, 10, 64)
	err = store.Add(storage.Entry{
		Category: m.Type,
		Title:    m.Title,
		DBId:     tvdbid,
		PlexKey:  uint(plexId),
	})
	var dupErr *storage.DuplicateEntryError
	if err != nil && errors.As(err, &dupErr) {
		log.Printf("[API] Entry already being tracked: %v", dupErr)
		j, _ := json.Marshal(apiPostResponse{Message: "Entry already being tracked"})
		_, _ = w.Write(j)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[API] Future releases of %v (%v) will automatically be downloaded", m.Title, m.Type)
	j, _ := json.Marshal(apiPostResponse{Message: fmt.Sprintf("Future releases of %v (%v) will automatically be downloaded", m.Title, m.Type)})
	_, _ = w.Write(j)
}

func deletePersist(w http.ResponseWriter, r *http.Request) {
	k := mux.Vars(r)["key"]

	m, err := plexServer.GetMediaMetadata(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tvdbid, err := plexServer.GetDbId(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = store.Remove(m.Type, tvdbid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getPersisted(w http.ResponseWriter, _ *http.Request) {
	e, err := store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(e)
	_, _ = w.Write(j)
}

func deletePersistForce(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)["partition"]
	row := mux.Vars(r)["row"]

	err := store.ForceRemove(p, row)
	if err != nil {
		log.Printf("[API] Failed to force delete entry at '%v' '%v'", p, row)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
