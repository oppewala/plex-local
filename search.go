package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/alediaferia/prefixmap"
)

type Match struct {
	Value      string
	Similarity float64
}

type Results []*Match

func (r Results) Len() int           { return len(r) }
func (r Results) Less(i, j int) bool { return r[i].Similarity < r[j].Similarity }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

var similarity = 0.3
var titles = prefixmap.New()
var media = make(map[string]Video)

func populateTitles() error {
	log.Printf("[Search] Starting library indexing")

	libs, err := fetchLibraries()
	if err != nil {
		err = fmt.Errorf("failed to retrieve libraries\n %v", err)
		return err
	}

	log.Printf("[Search] Retrieving library contents")

	for _, lib := range libs.Libraries {
		log.Printf("[Search][%s (%v)] Retrieving library contents ", lib.Title, lib.Key)
		k, _ := strconv.Atoi(lib.Key)
		lc, err := fetchLibraryContent(k)
		if err != nil {
			err = fmt.Errorf("failed to retrieve library content ([%v] %s)\n%v", lib.Key, lib.Title, err)
		}

		log.Printf("[Search][%s (%v)] Inserting %v titles", lib.Title, lib.Key, len(lc.Videos))
		// TODO: Handle Anime/TV
		for _, v := range lc.Videos {
			// TODO: should map to media key
			media[v.Title] = v

			parts := strings.Split(strings.ToLower(v.Title), " ")
			for _, p := range parts {
				// TODO: should map to media key
				titles.Insert(p, v.Title)
			}

			// Alternative names (eg, BOFURI jap name)
			// parts := strings.Split(strings.ToLower(v.AltTitle), " ")
			// for _, p := range parts {
			//	    titles.Insert(p, v.Title)
			// }
		}
	}

	log.Printf("[Search] Library indexing complete")

	return nil
}

func search(input string) []*Video {
	log.Printf("[Search] Starting for %s", input)
	values := titles.GetByPrefix(strings.ToLower(input))
	log.Printf("[Search] Found %v raw results", len(values))

	results := make(Results, 0, len(values))
	for _, v := range values {
		value := v.(string)
		s := ComputeSimilarity(len(value), len(input), LevenshteinDistance(value, input))
		m := &Match{value, s}
		results = append(results, m)
	}

	log.Printf("[Search] Found %v results after Levenshtein filter", len(results))

	sort.Sort(results)

	videos := make([]*Video, 0, len(results))
	for _, r := range results {
		v := media[r.Value]
		videos = append(videos, &v)
	}

	return videos
}

func ComputeSimilarity(w1Len, w2Len, ld int) float64 {
	maxLen := math.Max(float64(w1Len), float64(w2Len))

	return 1.0 - float64(ld)/float64(maxLen)
}

func LevenshteinDistance(source, destination string) int {
	vec1 := make([]int, len(destination)+1)
	vec2 := make([]int, len(destination)+1)

	w1 := []rune(source)
	w2 := []rune(destination)

	// initializing vec1
	for i := 0; i < len(vec1); i++ {
		vec1[i] = i
	}

	// initializing the matrix
	for i := 0; i < len(w1); i++ {
		vec2[0] = i + 1

		for j := 0; j < len(w2); j++ {
			cost := 1
			if w1[i] == w2[j] {
				cost = 0
			}
			min := minimum(vec2[j]+1,
				vec1[j+1]+1,
				vec1[j]+cost)
			vec2[j+1] = min
		}

		for j := 0; j < len(vec1); j++ {
			vec1[j] = vec2[j]
		}
	}

	return vec2[len(w2)]
}

func minimum(value0 int, values ...int) int {
	min := value0
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}
