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

func (s *Server) GetMediaMetadata(key string) (Metadata, error) {
	rootRequest := fmt.Sprintf("/library/metadata/%s", key)
	body, err := s.executeGet(rootRequest)
	if err != nil {
		err = fmt.Errorf("failed to get root metadata: %w", err)
		return Metadata{}, err
	}

	l := &ResponseRoot{}
	err = json.Unmarshal(body, l)
	if err != nil {
		log.Printf("%v", string(body))
		err = fmt.Errorf("failed to convert body from json \n%w", err)
		return Metadata{}, err
	}

	return l.MediaContainer.Metadata[0], nil
}

func (s *Server) GetMediaMetadataChildren(key string) ([]Metadata, error) {
	rootRequest := fmt.Sprintf("/library/metadata/%s/children", key)
	body, err := s.executeGet(rootRequest)
	if err != nil {
		err = fmt.Errorf("failed to get root metadata: %w", err)
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

func (s *Server) GetMetadataWithParts(key string) ([]Metadata, error) {
	m, err := s.GetMediaMetadata(key)
	if err != nil {
		return nil, err
	}

	meta := make([]Metadata, 0)
	var seasons []Metadata
	switch m.Type {
	case "movie":
		meta = append(meta, m)
		return meta, nil
	case "episode":
		meta = append(meta, m)
		return meta, nil
	case "show":
		seasons, err = s.GetMediaMetadataChildren(key)
		if err != nil {
			err = fmt.Errorf("failed while retrieving child metadata for show with key %s: %w", key, err)
			return nil, err
		}
	case "season":
		seasons = make([]Metadata, 1)
		seasons[0] = m
	default:
		err = fmt.Errorf("unhandled metadata type: %s", m.Type)
		return nil, err
	}

	for _, season := range seasons {
		episodes, err := s.GetMediaMetadataChildren(season.RatingKey)
		if err != nil {
			err = fmt.Errorf("failed while retrieving child metadata for season with key %s: %w", season.RatingKey, err)
			return nil, err
		}

		meta = append(meta, episodes...)
	}

	return meta, err
}
