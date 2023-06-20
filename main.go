package main

import (
	"database/sql"
	"flag"
	"go-wikitionary-parse/lib"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/macdub/go-colorlog"
	_ "github.com/mattn/go-sqlite3"
	_ "go-wikitionary-parse/lib"
)

var (
	// other stuff
	language        string             = ""
	logger          *colorlog.ColorLog = &colorlog.ColorLog{}
	lexicalCategory []string           = []string{"Proper noun", "Noun", "Adjective", "Adverb",
		"Verb", "Article", "Particle", "Conjunction",
		"Pronoun", "Determiner", "Interjection", "Morpheme",
		"Numeral", "Preposition", "Postposition"}
)

func main() {
	iFile := flag.String("file", "", "XML file to parse")
	db := flag.String("database", "database.db", "Database file to use")
	lang := flag.String("lang", "English", "Language to target for parsing")
	cacheFile := flag.String("cache_file", "xmlCache.gob", "Use this as the cache file")
	logFile := flag.String("log_file", "", "Log to this file")
	threads := flag.Int("threads", 5, "Number of threads to use for parsing")
	useCache := flag.Bool("use_cache", false, "Use a 'gob' of the parsed XML file")
	makeCache := flag.Bool("make_cache", false, "Make a cache file of the parsed XML")
	purge := flag.Bool("purge", false, "Purge the selected database")
	verbose := flag.Bool("verbose", false, "Use verbose logging")
	flag.Parse()

	if *logFile != "" {
		logger = colorlog.NewFileLog(colorlog.Linfo, *logFile)
	} else {
		logger = colorlog.New(colorlog.Linfo)
	}

	if *verbose {
		logger.SetLogLevel(colorlog.Ldebug)
	}

	language = *lang

	startTime := time.Now()
	logger.Info("+--------------------------------------------------\n")
	logger.Info("| Start Time    :    %v\n", startTime)
	logger.Info("| Parse File    :    %s\n", *iFile)
	logger.Info("| Database      :    %s\n", *db)
	logger.Info("| Language      :    %s\n", language)
	logger.Info("| Cache File    :    %s\n", *cacheFile)
	logger.Info("| Use Cache     :    %t\n", *useCache)
	logger.Info("| Make Cache    :    %t\n", *makeCache)
	logger.Info("| Verbose       :    %t\n", *verbose)
	logger.Info("| Purge         :    %t\n", *purge)
	logger.Info("+--------------------------------------------------\n")

	logger.Debug("NOTE: input language should be provided as a proper noun. (e.g. English, French, West Frisian, etc.)\n")

	data := &lib.WikiData{}
	if *useCache {
		d, err := lib.DecodeCache(*cacheFile)
		data = d
		lib.Check(err)
	} else if *iFile == "" {
		logger.Error("Input file is empty. Exiting\n")
		os.Exit(1)
	} else {
		logger.Info("Parsing XML file\n")
		d := lib.ParseXML(*makeCache, *iFile, *cacheFile)
		data = d
	}

	if *purge {
		err := os.Remove(*db)
		lib.Check(err)
	}

	logger.Debug("Number of Pages: %d\n", len(data.Pages))
	dbh, err := lib.Init(*db)
	lib.Check(err)

	lib.FilterPages(data, language)
	logger.Info("Post filter page count: %d\n", len(data.Pages))

	// split the work into 5 chunks
	var chunks [][]lib.Page
	size := len(data.Pages) / *threads
	logger.Debug("Chunk size: %d\n", size)
	logger.Debug(" >> %d\n", len(data.Pages)/size)
	for i := 0; i < *threads; i++ {
		end := size + size*i
		if end > len(data.Pages) || i+1 == *threads {
			end = len(data.Pages)
		}
		logger.Debug("Splitting chunk %d :: [%d, %d]\n", i, size*i, end)
		chunks = append(chunks, data.Pages[size*i:end])
	}

	logger.Debug("Have %d chunks\n", len(chunks))
	logger.Debug("Chunk Page Last: %s Page Last: %s\n", chunks[len(chunks)-1][len(chunks[len(chunks)-1])-1].Title, data.Pages[len(data.Pages)-1].Title)

	var wg sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go pageWorker(i, &wg, chunks[i], dbh)
	}

	wg.Wait()

	endTime := time.Now()
	logger.Info("Completed in %s\n", endTime.Sub(startTime))
}

func pageWorker(id int, wg *sync.WaitGroup, pages []lib.Page, dbh *sql.DB) {
	defer wg.Done()
	var inserts []*lib.Insert // etymology : lexical category : [definitions...]
	for _, page := range pages {
		word := page.Title
		logger.Debug("Processing page: %s\n", word)

		// convert the text to a byte string
		text := []byte(page.Revisions[0].Text)
		logger.Debug("Raw size: %d\n", len(text))

		text = lib.WikiModifier.ReplaceAll(text, []byte("'$1'"))
		logger.Debug("Modifier size: %d\n", len(text))

		//text = wikiLabel.ReplaceAll(text, []byte("(${2})"))
		//logger.Debug("Label size: %d\n", len(text))

		text = lib.WikiExample.ReplaceAll(text, []byte(""))
		logger.Debug("Example size: %d\n", len(text))

		text = lib.WikiWordAlt.ReplaceAll(text, []byte("$1"))
		logger.Debug("WordAlt size: %d\n", len(text))

		text = lib.WikiBracket.ReplaceAll(text, []byte(""))
		logger.Debug("Bracket size: %d\n", len(text))

		text = lib.HtmlBreak.ReplaceAll(text, []byte(" "))
		logger.Debug("Html Break size: %d\n", len(text))

		textSize := len(text)
		logger.Debug("Starting Size of corpus: %d bytes\n", textSize)

		// get language section of the page
		text = lib.GetLanguageSection(text, language)
		logger.Debug("Reduced corpus by %d bytes to %d\n", textSize-len(text), len(text))

		// get all indices of the etymology headings
		etymologyIdx := lib.WikiEtymologyM.FindAllIndex(text, -1)
		if len(etymologyIdx) == 0 {
			logger.Debug("Did not find multi-style etymology. Checking for singular ...\n")
			etymologyIdx = lib.WikiEtymologyS.FindAllIndex(text, -1)
		}
		/*
		   When there is only a single or no etymology, then lexical categories are of the form ===[\w\s]+===
		   Otherwise, then lexical categories are of the form ====[\w\s]+====
		*/
		logger.Debug("Found %d etymologies\n", len(etymologyIdx))
		if len(etymologyIdx) <= 1 {
			// need to get the lexical category via regexp
			logger.Debug("Parsing by lexical category\n")
			lexcatIdx := lib.WikiLexS.FindAllIndex(text, -1)
			inserts = append(inserts, parseByLexicalCategory(word, lexcatIdx, text)...)
		} else {
			logger.Debug("Parsing by etymologies\n")
			inserts = append(inserts, parseByEtymologies(word, etymologyIdx, text)...)
		}
	}

	// perform inserts
	inserted := lib.PerformInserts(dbh, inserts)
	logger.Info("[Worker %2d] Inserted %6d records for %6d pages\n", id, inserted, len(pages))
}

func parseByEtymologies(word string, etList [][]int, text []byte) []*lib.Insert {
	var inserts []*lib.Insert
	etSize := len(etList)
	for i := 0; i < etSize; i++ {
		ins := &lib.Insert{Word: word, Etymology: i, CatDefs: make(map[string][]string)}
		var section []byte
		if i+1 >= etSize {
			section = lib.GetSection(etList[i][1], -1, text)
		} else {
			section = lib.GetSection(etList[i][1], etList[i+1][0], text)
		}

		logger.Debug("parseByEtymologies> Section is %d bytes\n", len(section))

		lexcatIdx := lib.WikiLexM.FindAllIndex(section, -1)
		lexcatIdxSize := len(lexcatIdx)

		var definitions []string
		for j := 0; j < lexcatIdxSize; j++ {
			jthIdx := lib.AdjustIndexLW(lexcatIdx[j][0], section)
			lexcat := string(section[jthIdx+4 : lexcatIdx[j][1]-4])
			logger.Debug("parseByEtymologies> [%2d] lexcat: %s\n", j, lexcat)

			if !lib.StringInSlice(lexcat, lexicalCategory) {
				logger.Debug("parseByLemmas> Lexical category '%s' not in list. Skipping...\n", lexcat)
				continue
			}

			nHeading := lib.WikiGenHeading.FindIndex(section[lexcatIdx[j][1]:])
			if len(nHeading) > 0 {
				nHeading[0] = nHeading[0] + lexcatIdx[j][1]
				nHeading[1] = nHeading[1] + lexcatIdx[j][1]
				logger.Debug("parseByLemmas> LEM_LIST %d: %+v NHEADING: %+v\n", j, lexcatIdx[j], nHeading)
				definitions = lib.GetDefinitions(lexcatIdx[j][1], nHeading[0], section)
			} else if j+1 >= lexcatIdxSize {
				definitions = lib.GetDefinitions(lexcatIdx[j][1], -1, section)
			} else {
				jth1Idx := lib.AdjustIndexLW(lexcatIdx[j+1][0], section)
				definitions = lib.GetDefinitions(lexcatIdx[j][1], jth1Idx, section)
			}
			logger.Debug("parseByEtymologies> Definitions: " + strings.Join(definitions, ", ") + "\n")
			ins.CatDefs[lexcat] = definitions
		}
		inserts = append(inserts, ins)
	}

	return inserts
}

// parseByLemmas
func parseByLexicalCategory(word string, lexList [][]int, text []byte) []*lib.Insert {
	var inserts []*lib.Insert
	lexSize := len(lexList)
	logger.Debug("parseByLexicalCategory> Found %d lexcats\n", lexSize)

	for i := 0; i < lexSize; i++ {
		ins := &lib.Insert{Word: word, Etymology: 0, CatDefs: make(map[string][]string)}
		ithIdx := lib.AdjustIndexLW(lexList[i][0], text)
		lexcat := string(text[ithIdx+3 : lexList[i][1]-3])

		logger.Debug("parseByLexicalCategory> [%2d] working on lexcat '%s'\n", i, lexcat)

		if !lib.StringInSlice(lexcat, lexicalCategory) {
			logger.Debug("parseByLexicalCategory> Lemma '%s' not in list. Skipping...\n", lexcat)
			continue
		}

		var definitions []string
		if i+1 >= lexSize {
			definitions = lib.GetDefinitions(lexList[i][1], -1, text)
		} else {
			ith1Idx := lib.AdjustIndexLW(lexList[i+1][0], text)
			logger.Debug("parseByLexicalCategory> LEMMA: %s\n", string(text[lexList[i][1]:ith1Idx]))
			definitions = lib.GetDefinitions(lexList[i][1], ith1Idx, text)
		}

		logger.Debug("parseByLexicalCategory> Found %d definitions\n", len(definitions))
		ins.CatDefs[lexcat] = definitions

		inserts = append(inserts, ins)
	}

	return inserts
}
