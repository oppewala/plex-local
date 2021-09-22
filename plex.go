package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type LibraryContainer struct {
	XMLName   xml.Name  `xml:"MediaContainer" json:"-"`
	Title     string    `xml:"title1,attr"`
	Libraries []Library `xml:"Directory"`
	Size      string    `xml:"size"`
}

type Library struct {
	XMLName xml.Name `xml:"Directory" json:"-"`
	Title   string   `xml:"title,attr"`
	Key     string   `xml:"key,attr"`
}

// VideoContainer returns from the library sections API when querying a movie or tvshow with /allLeaves
type VideoContainer struct {
	XMLName  xml.Name `xml:"MediaContainer" json:"-"`
	Title    string   `xml:"title1,attr"`
	SubTitle string   `xml:"title2,attr"`
	Videos   []Video  `xml:"Video"`
}

type Video struct {
	XMLName     xml.Name `xml:"Video" json:"-"`
	Title       string   `xml:"title,attr"`
	Key         string   `xml:"ratingKey,attr"`
	RelativeUrl string   `xml:"key,attr" json:"-"`
	Media       Media    `xml:"Media" json:"-"`
}

type Media struct {
	XMLName xml.Name `xml:"Media" json:"-"`
	Parts   []Part   `xml:"Part"`
}

type Part struct {
	XMLName xml.Name `xml:"Part" json:"-"`
	Key     string   `xml:"key,attr"`
	Path    string   `xml:"file,attr"`
	Size    uint64   `xml:"size,attr"`
}

func fetchVideo(key int) (*Video, error) {
	u := fmt.Sprintf("%v/library/metadata/%v?X-Plex-Token=%v", plexUrl, key, plexToken)
	req, _ := http.NewRequest("GET", u, nil)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to GET library data (%v - %v) \n %v", u, res.StatusCode, err)
		return nil, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	c := &VideoContainer{}
	err = xml.Unmarshal(body, c)
	if err != nil {
		err = fmt.Errorf("failed to convert body to xml \n %v", err)
		return nil, err
	}

	return &c.Videos[0], err
}
