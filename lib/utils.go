package lib

import (
	"encoding/xml"
	"fmt"
	"github.com/macdub/go-colorlog"
	"io/ioutil"
	"regexp"
	"time"
)

// filter out the pages that are not words in the desired language
func FilterPages(wikidata *WikiData, language string) {
	engCheck := regexp.MustCompile(fmt.Sprintf(`==%s==`, language))
	spaceCheck := regexp.MustCompile(`[:0-9]`)
	skipCount := 0
	i := 0
	for i < len(wikidata.Pages) {
		if !engCheck.MatchString(wikidata.Pages[i].Revisions[0].Text) || spaceCheck.MatchString(wikidata.Pages[i].Title) {
			// remove the entry from the array
			wikidata.Pages[i] = wikidata.Pages[len(wikidata.Pages)-1]
			wikidata.Pages = wikidata.Pages[:len(wikidata.Pages)-1]
			skipCount++
			continue
		}
		i++
	}

	logger.Debug("Skipped %d pages\n", skipCount)
}

// parse the input XML file into a struct and create a cache file optionally
func ParseXML(makeCache bool, parseFile string, cacheFile string) *WikiData {
	logger.Info("Opening xml file\n")
	file, err := ioutil.ReadFile(parseFile)
	Check(err)

	wikidata := &WikiData{}

	start := time.Now()
	logger.Info("Unmarshalling xml ... ")
	err = xml.Unmarshal(file, wikidata)
	end := time.Now()
	logger.Printc(colorlog.Linfo, colorlog.Grey, "elapsed %s\n", end.Sub(start))
	Check(err)

	logger.Info("Parsed %d pages\n", len(wikidata.Pages))

	if makeCache {
		err = EncodeCache(wikidata, cacheFile)
		Check(err)
	}

	return wikidata
}

// Helper functions
func Check(err error) {
	if err != nil {
		logger.Fatal("%s\n", err.Error())
		panic(err)
	}
}

func GetSection(start int, end int, text []byte) []byte {
	if end < 0 {
		return text[start:]
	}

	return text[start:end]
}

func StringInSlice(str string, list []string) bool {
	for _, lStr := range list {
		if str == lStr {
			return true
		}
	}
	return false
}

// adjust the index offset to account for leading whitespace character
func AdjustIndexLW(index int, text []byte) int {
	if text[index : index+1][0] == byte('\n') {
		index++
	}
	return index
}
