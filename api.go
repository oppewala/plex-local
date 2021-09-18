package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"github.com/wagslane/go-rabbitmq"
)

func getSections(w http.ResponseWriter, _ *http.Request) {
	u := fmt.Sprintf("%v/library/sections?X-Plex-Token=%v", plexUrl, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get library data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	l := LibraryContainer{}
	_ = xml.Unmarshal(body, &l)

	j, _ := json.Marshal(l)

	_, _ = w.Write(j)
}

func getSectionMedia(w http.ResponseWriter, r *http.Request) {
	section := mux.Vars(r)["section"]

	u := fmt.Sprintf("%v/library/sections/%v/all?X-Plex-Token=%v", plexUrl, section, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get library data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	c := VideoContainer{}
	_ = xml.Unmarshal(body, &c)

	j, _ := json.Marshal(c)

	_, _ = w.Write(j)
}

func getMediaMetadata(w http.ResponseWriter, r *http.Request) {
	media := mux.Vars(r)["media"]

	u := fmt.Sprintf("%v/library/metadata/%v?X-Plex-Token=%v", plexUrl, media, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get library data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	c := VideoContainer{}
	_ = xml.Unmarshal(body, &c)

	j, _ := json.Marshal(c.Videos[0])

	_, _ = w.Write(j)
}

func queue(w http.ResponseWriter, r *http.Request) {
	media := mux.Vars(r)["media"]

	u := fmt.Sprintf("%v/library/metadata/%v?X-Plex-Token=%v", plexUrl, media, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get media data (%v - %v) \n %v", u, res.StatusCode, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	vc := VideoContainer{}
	_ = xml.Unmarshal(body, &vc)

	v := vc.Videos[0]

	log.Printf("Triggering download of media (%v - %v - %v)", media, v.Title, v.Media.Parts[0].Path)

	c := amqp.Config{}
	publisher, returns, err := rabbitmq.NewPublisher(rUrl, c)
	if err != nil {
		log.Fatal(err)
	}

	j, _ := json.Marshal(v)
	err = publisher.Publish(j, []string{requestQueue})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for r := range returns {
			log.Printf("message returned from server: %s", string(r.Body))
		}
	}()

	_, _ = w.Write(j)
}
