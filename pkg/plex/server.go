package plex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct {
	URL   string
	Token string
}

func NewServer(url string, token string) *Server {
	s := &Server{
		URL:   url,
		Token: token,
	}

	return s
}

func (s *Server) executeGet(path string) ([]byte, error) {
	u := fmt.Sprintf("%v%v?X-Plex-Token=%v", s.URL, path, s.Token)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to execute GET on %v (%v - %v)\n%w", path, u, res.StatusCode, err)
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func (s *Server) GetLibraries() ([]Directory, error) {
	body, err := s.executeGet("/library/sections")
	if err != nil {
		return nil, err
	}

	l := &ResponseRoot{}
	err = json.Unmarshal(body, l)
	if err != nil {
		log.Printf("%v", string(body))
		err = fmt.Errorf("failed to convert body from json \n%w", err)
		return nil, err
	}

	return l.MediaContainer.Directory, nil
}

func (s *Server) GetLibraryContent(key string) ([]Metadata, error) {
	body, err := s.executeGet(fmt.Sprintf("/library/sections/%s/all", key))
	if err != nil {
		return nil, err
	}

	l := &ResponseRoot{}
	err = json.Unmarshal(body, l)
	if err != nil {
		log.Printf("%v", string(body))
		err = fmt.Errorf("failed to convert body from json \n%w", err)
		return nil, err
	}

	return l.MediaContainer.Metadata, nil
}

func (s *Server) GetMediaMetadata(key string) ([]Metadata, error) {
	rootRequest := fmt.Sprintf("/library/metadata/%s", key)
	body, err := s.executeGet(rootRequest)
	if err != nil {
		fmt.Errorf("failed to get root metadata")
		return nil, err
	}

	l := &ResponseRoot{}
	err = json.Unmarshal(body, l)
	if err != nil {
		log.Printf("%v", string(body))
		err = fmt.Errorf("failed to convert body from json \n%w", err)
		return nil, err
	}

	return l.MediaContainer.Metadata, nil
}

func (s *Server) GetMediaMetadataChildren(key string) ([]Metadata, error) {
	rootRequest := fmt.Sprintf("/library/metadata/%s/children", key)
	body, err := s.executeGet(rootRequest)
	if err != nil {
		fmt.Errorf("failed to get root metadata")
		return nil, err
	}

	l := &ResponseRoot{}
	err = json.Unmarshal(body, l)
	if err != nil {
		log.Printf("%v", string(body))
		err = fmt.Errorf("failed to convert body from json \n%w", err)
		return nil, err
	}

	return l.MediaContainer.Metadata, nil
}

func (s *Server) GetMediaParts(key string) ([]Part, error) {
	m, err := s.GetMediaMetadata(key)
	if err != nil {
		return nil, err
	}

	// Movie or episode requested
	if m[0].Type == "movie" || m[0].Type == "episode" {
		return m[0].Media[0].Part, nil
	}

	// key = 3298 - show
	if m[0].Type == "show" {
		// m = Metadata per season
		m, err = s.GetMediaMetadataChildren(key)
		if err != nil {
			err = fmt.Errorf("failed while retrieving child metadata for show with key %s: %w", key, err)
			return nil, err
		}
	}

	p := make([]Part, 0)
	for _, ms := range m {
		msc, err := s.GetMediaMetadataChildren(ms.RatingKey)
		if err != nil {
			err = fmt.Errorf("failed while retrieving child metadata for season with key %s: %w", ms.RatingKey, err)
			return nil, err
		}

		for _, e := range msc {
			p = append(p, e.Media[0].Part...)
		}
	}

	return p, err
}
