package lib

import "encoding/xml"

type WikiData struct {
	XMLName xml.Name `xml:"mediawiki"`
	Pages   []Page   `xml:"page"`
}

type Page struct {
	XMLName   xml.Name   `xml:"page"`
	Title     string     `xml:"title"`
	Id        int        `xml:"id"`
	Revisions []Revision `xml:"revision"`
}

type Revision struct {
	Id      int    `xml:"id"`
	Comment string `xml:"comment"`
	Model   string `xml:"model"`
	Format  string `xml:"format"`
	Text    string `xml:"text"`
	Sha1    string `xml:"sha1"`
}

type Insert struct {
	Word      string
	Etymology int
	CatDefs   map[string][]string
}
