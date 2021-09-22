package plex

import "fmt"

func (m Metadata) ConcatTitles() string {
	t := ""
	if m.GrandparentTitle != "" {
		t += fmt.Sprintf("%s - ", m.GrandparentTitle)
	}
	if m.ParentTitle != "" {
		t += fmt.Sprintf("%s - ", m.ParentTitle)
	}
	t += m.Title
	return t
}
