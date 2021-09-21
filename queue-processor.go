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
)

var requestQueue = make(chan *Video)

type DownloadUpdate struct {
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

func (wc *DownloadTracker) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)

	if time.Now().After(wc.NextUpdate) {
		update := &DownloadUpdate{
			Title:           wc.Title,
			BytesDownloaded: wc.Total,
			TotalBytes:      wc.ExpectedTotal,
		}
		j, _ := json.Marshal(update)
		wc.Hub.broadcast <- j

		wc.NextUpdate = time.Now().Add(time.Second)
	}

	return n, nil
}

func chanConsumer(hub *Hub) {
	for {
		v := <-requestQueue
		log.Printf("Message Consumed: %v", v)

		err := downloadMedia(*v, hub)
		if err != nil {
			log.Printf("Failed to download %v: %v", v.Media.Parts[0].Key, err)
			continue
		}

		log.Printf("Downloaded succesfully: %v", v.Title)
	}
}

func downloadMedia(v Video, hub *Hub) error {
	part := v.Media.Parts[0]

	path := fmt.Sprintf("/data/local%s", part.Path)
	log.Printf("Downloading from %v to %v", part.Key, path)

	log.Printf("Creating directory: %v", filepath.Dir(path))
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	log.Printf("Creating temp file: %v", path+".tmp")
	file, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer file.Close()

	log.Printf("Creating request")
	res, err := http.Get(fmt.Sprintf("%v%v?X-Plex-Token=%v", plexUrl, part.Key, plexToken))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	log.Printf("Starting write")
	counter := &DownloadTracker{
		Title:         v.Title,
		Key:           v.Key,
		ExpectedTotal: part.Size,
		StartTime:     time.Now(),
		Hub:           hub,
	}
	_, err = io.Copy(file, io.TeeReader(res.Body, counter))
	if err != nil {
		return err
	}

	log.Printf("Closing file and renaming to final path: %v", path)
	file.Close()
	err = os.Rename(path+".tmp", path)

	// TODO: Notify autoscan and/or plex

	return err
}
