package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/Azure/azure-storage-queue-go/azqueue"
)

type RadarrWebhook struct {
	Movie struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Year        int    `json:"year"`
		ReleaseDate string `json:"releaseDate"`
		FolderPath  string `json:"folderPath"`
		TmdbID      int    `json:"tmdbId"`
		ImdbID      string `json:"imdbId"`
	} `json:"movie"`
	RemoteMovie struct {
		TmdbID int    `json:"tmdbId"`
		ImdbID string `json:"imdbId"`
		Title  string `json:"title"`
		Year   int    `json:"year"`
	} `json:"remoteMovie"`
	MovieFile struct {
		ID             int    `json:"id"`
		RelativePath   string `json:"relativePath"`
		Path           string `json:"path"`
		Quality        string `json:"quality"`
		QualityVersion int    `json:"qualityVersion"`
		ReleaseGroup   string `json:"releaseGroup"`
		SceneName      string `json:"sceneName"`
		IndexerFlags   string `json:"indexerFlags"`
		Size           int64  `json:"size"`
	} `json:"movieFile"`
	IsUpgrade    bool   `json:"isUpgrade"`
	DownloadID   string `json:"downloadId"`
	DeletedFiles []struct {
		ID             int    `json:"id"`
		RelativePath   string `json:"relativePath"`
		Path           string `json:"path"`
		Quality        string `json:"quality"`
		QualityVersion int    `json:"qualityVersion"`
		ReleaseGroup   string `json:"releaseGroup"`
		SceneName      string `json:"sceneName"`
		IndexerFlags   string `json:"indexerFlags"`
		Size           int64  `json:"size"`
	} `json:"deletedFiles"`
	EventType string `json:"eventType"`
}

type SonarrWebhook struct {
	Series struct {
		ID       int    `json:"id"`
		Title    string `json:"title"`
		Path     string `json:"path"`
		TvdbID   int    `json:"tvdbId"`
		TvMazeID int    `json:"tvMazeId"`
		ImdbID   string `json:"imdbId"`
		Type     string `json:"type"`
	} `json:"series"`
	Episodes []struct {
		ID            int       `json:"id"`
		EpisodeNumber int       `json:"episodeNumber"`
		SeasonNumber  int       `json:"seasonNumber"`
		Title         string    `json:"title"`
		AirDate       string    `json:"airDate"`
		AirDateUtc    time.Time `json:"airDateUtc"`
	} `json:"episodes"`
	//EpisodeFile struct {
	//	ID             int    `json:"id"`
	//	RelativePath   string `json:"relativePath"`
	//	Path           string `json:"path"`
	//	//Quality        string `json:"quality"`
	//	//QualityVersion int    `json:"qualityVersion"`
	//	ReleaseGroup   string `json:"releaseGroup"`
	//	SceneName      string `json:"sceneName"`
	//	Size           int    `json:"size"`
	//} `json:"episodeFile"`
	IsUpgrade      bool   `json:"isUpgrade"`
	DownloadClient string `json:"downloadClient"`
	DownloadID     string `json:"downloadId"`
	EventType      string `json:"eventType"`
}

func newQueueConsumer(connString string) {
	r := regexp.MustCompile("DefaultEndpointsProtocol=(?P<protocol>[^;]+);AccountName=(?P<accountName>[^;]+);AccountKey=(?P<accountKey>[^;]+);EndpointSuffix=(?P<endpoingSuffix>[a-z0-9.]+)")
	m := r.FindStringSubmatch(connString)
	saProtocol, saAccount, saKey, saSuffix := m[1], m[2], m[3], m[4]

	c, err := azqueue.NewSharedKeyCredential(saAccount, saKey)
	if err != nil {
		log.Fatalf("[Arr] Invalid azure queue credentials: %v", err)
	}
	po := azqueue.PipelineOptions{}
	p := azqueue.NewPipeline(c, po)

	u := fmt.Sprintf("%s://%s.queue.%s/webhook-requests/messages", saProtocol, saAccount, saSuffix)
	qu, err := url.Parse(u)
	if err != nil {
		log.Fatalf("[Arr] Failed to create valid azure queue URL: %s", u)
	}
	q := azqueue.NewMessagesURL(*qu, p)
	//_, err = q.Peek(context.Background(), 1)
	//if err != nil {
	//	log.Fatalf("[Arr] Failed to connect to azure queue URL %s: %v", u, err)
	//}

	log.Printf("[Arr] Connected to azure queue: %s", u)

	go func(q *azqueue.MessagesURL) {
		fails := 0
		for {
			r, err := q.Dequeue(context.TODO(), 5, time.Second*30)
			if err != nil {
				d := time.Minute * 2
				if fails > 3 {
					d = time.Minute * 30
				}

				fails++
				log.Printf("[Arr] Failed to dequeue message. Pausing for %v. (Fail %v): %v", d, fails, err)
				time.Sleep(d)

				continue
			}

			for i := int32(0); i < r.NumMessages(); i++ {
				m := r.Message(i)

				err = handleMessage(m.Text)
				if err != nil {
					log.Printf("[Arr] Failed to handle message: %v\n%v", err, m.Text)

					// TODO: How to check retry status? Always delete on error for now to avoid poison pill
					//continue
				}

				murl := q.NewMessageIDURL(m.ID)
				_, _ = murl.Delete(context.TODO(), m.PopReceipt)
			}
		}
	}(&q)
}

func handleMessage(msg string) error {
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &dat); err != nil {
		err = fmt.Errorf("could not unmarshal dequeued message: %w", err)
		return err
	}

	if dat["movie"] != nil {
		var wh RadarrWebhook
		if err := json.Unmarshal([]byte(msg), &wh); err != nil {
			err = fmt.Errorf("failed to parse 'movie' as object: %w", err)
			return err
		}
		if err := handleMovie(wh); err != nil {
			err = fmt.Errorf("failed to handle movie message %w", err)
			return err
		}
		return nil
	}

	if dat["series"] != nil {
		var wh SonarrWebhook
		if err := json.Unmarshal([]byte(msg), &wh); err != nil {
			err = fmt.Errorf("failed to parse 'series' as object %w", err)
			return err
		}
		if err := handleSeries(wh); err != nil {
			err = fmt.Errorf("failed to handle series message: %w", err)
			return err
		}
		return nil
	}

	return errors.New(fmt.Sprintf("could not handle message: %v", msg))
}

func handleMovie(wh RadarrWebhook) error {
	log.Printf("[Arr] Movie webhook message received (%s - %s)", wh.Movie.ImdbID, wh.Movie.Title)

	exists, err := store.Exists("movie", wh.Movie.ImdbID)
	if err != nil {
		err = fmt.Errorf("failed to check store for movie (%s - %s): %w", wh.Movie.ImdbID, wh.Movie.Title, err)
		return err
	}
	if !exists {
		log.Printf("[Arr] Movie not found in store (%s - %s)", wh.Movie.ImdbID, wh.Movie.Title)
		return nil
	}

	e, err := store.Get("movie", wh.Movie.ImdbID)
	k := strconv.FormatUint(uint64(e.PlexKey), 10)

	parts, err := plexServer.GetMetadataWithParts(k)
	if err != nil {
		err = fmt.Errorf("failed to get metadata (%s - %s - %s): %w", k, wh.Movie.ImdbID, wh.Movie.Title, err)
		return err
	}

	for _, p := range parts {
		queueDownload(p)
	}

	return nil
}

func handleSeries(wh SonarrWebhook) error {
	log.Printf("[Arr] Series webhook message received (%v - %s)", wh.Series.TvdbID, wh.Series.Title)

	id := strconv.FormatInt(int64(wh.Series.TvdbID), 10)
	exists, err := store.Exists("series", id)
	if err != nil {
		err = fmt.Errorf("failed to check store for series (%s - %s): %w", id, wh.Series.Title, err)
		return err
	}
	if !exists {
		log.Printf("[Arr] Series not found in store (%s - %s)", id, wh.Series.Title)
		return nil
	}

	e, err := store.Get("series", id)
	k := strconv.FormatUint(uint64(e.PlexKey), 10)

	// TODO: This currently downloads all episodes again, instead of only queuing the new episodes
	meta, err := plexServer.GetMetadataWithParts(k)
	if err != nil {
		err = fmt.Errorf("failed to get metadata (%s - %s - %s): %w", k, wh.Series.ImdbID, wh.Series.Title, err)
		return err
	}

	for _, p := range meta {
		queueDownload(p)
	}

	return nil
}
