package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/oppewala/plex-local-dl/pkg/plex"
)

var requestQueue = make(chan DownloadRequest)

type DownloadRequest struct {
	Metadata plex.Metadata
	Part     plex.Part
}

type DownloadUpdate struct {
	MessageType     string
	Title           string
	BytesDownloaded uint64
	TotalBytes      uint64
}

type DownloadTracker struct {
	Title         string
	Key           string
	Total         uint64
	ExpectedTotal uint64
	NextUpdate    time.Time
	StartTime     time.Time
	Hub           *Hub
}

func (u *DownloadUpdate) ToBytes() []byte {
	j, _ := json.Marshal(u)

	return j
}

func (wc *DownloadTracker) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)

	if time.Now().After(wc.NextUpdate) {
		update := &DownloadUpdate{
			MessageType:     "download-update",
			Title:           wc.Title,
			BytesDownloaded: wc.Total,
			TotalBytes:      wc.ExpectedTotal,
		}
		wc.Hub.broadcast <- update

		wc.NextUpdate = time.Now().Add(time.Second)
	}

	return n, nil
}

func chanConsumer(hub *Hub) {
	for {
		r := <-requestQueue
		log.Printf("[Processor] Message Consumed: %v", r)

		err := downloadMedia(r, hub)
		if err != nil {
			log.Printf("[Processor] Failed to download %v: %v", r.Metadata.ConcatTitles(), err)
			continue
		}

		log.Printf("[Processor] Downloaded succesfully: %v", r.Metadata.ConcatTitles())
	}
}

func downloadMedia(r DownloadRequest, hub *Hub) error {
	path := fmt.Sprintf("%s%s", mediaPath, r.Part.File)
	log.Printf("[Processor] Downloading %s from %v to %v", r.Metadata.Title, r.Part.Key, path)

	log.Printf("[Processor] Creating directory: %v", filepath.Dir(path))
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	log.Printf("[Processor] Creating temp file: %v", path+".tmp")
	file, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer file.Close()

	log.Printf("[Processor] Creating request")
	res, err := http.Get(fmt.Sprintf("%v%v?X-Plex-Token=%v", plexUrl, r.Part.Key, plexToken))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	log.Printf("[Processor] Starting write")
	counter := &DownloadTracker{
		Title:         r.Metadata.ConcatTitles(),
		Key:           r.Part.Key,
		ExpectedTotal: r.Part.Size,
		StartTime:     time.Now(),
		Hub:           hub,
	}
	_, err = io.Copy(file, io.TeeReader(res.Body, counter))
	if err != nil {
		return err
	}

	log.Printf("[Processor] Closing file and renaming to final path: %v", path)
	file.Close()
	err = os.Rename(path+".tmp", path)

	hub.broadcast <- &DownloadUpdate{
		MessageType: "download-complete",
		Title:       r.Metadata.ConcatTitles(),
	}
	// TODO: Notify autoscan and/or plex

	return err
}
