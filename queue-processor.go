package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
)

type DownloadTracker struct {
	Title         string
	Total         uint64
	ExpectedTotal uint64
	NextPrint     time.Time
	StartTime     time.Time
}

func (wc *DownloadTracker) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *DownloadTracker) PrintProgress() {
	if time.Now().After(wc.NextPrint) {
		diff := time.Now().Sub(wc.StartTime).Nanoseconds()
		estRemaining := uint64(diff) * ((wc.ExpectedTotal - wc.Total) / wc.Total)
		estComplete := time.Now().Add(time.Duration(estRemaining))

		log.Printf("Diff: %v | Remaining: %v | Complete: %v", diff, estRemaining, estComplete)

		log.Printf("Downloading... %s / %s (%v%%)", humanize.Bytes(wc.Total), humanize.Bytes(wc.ExpectedTotal), wc.Total/wc.ExpectedTotal)
		wc.NextPrint = time.Now().Add(time.Second * 5)
	}

	return
}

func chanConsumer() {
	for {
		v := <-requestQueue
		log.Printf("Message Consumed: %v", v)

		err := downloadMedia(*v)
		if err != nil {
			log.Printf("Failed to download %v: %v", v.Media.Parts[0].Key, err)
			continue
		}

		log.Printf("Downloaded succesfully: %v", v.Title)
	}
}

func downloadMedia(v Video) error {
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
		ExpectedTotal: part.Size,
		NextPrint:     time.Now(),
		StartTime:     time.Now(),
	}
	_, err = io.Copy(file, io.TeeReader(res.Body, counter))
	if err != nil {
		return err
	}

	log.Printf("Closing file and renaming to final path: %v", path)
	file.Close()
	err = os.Rename(path+".tmp", path)

	return err
}
