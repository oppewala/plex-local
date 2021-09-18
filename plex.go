package main

import "encoding/xml"

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
	Media       Media    `xml:"Media"`
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
