package lib

import (
	"bytes"
	"fmt"
	"go-wikitionary-parse/lib/wikitemplates"
)

func getTranslations(start int, end int, text []byte) []string {
	return []string{}
}

func GetDefinitions(start int, end int, text []byte) []string {
	var category []byte
	var defs []string

	if end < 0 {
		category = text[start:]
	} else {
		category = text[start:end]
	}

	logger.Debug("getDefinitions> TEXT: %s\n", string(text))
	nHeading := WikiGenHeading.FindIndex(text[start:])
	logger.Debug("getDefinitions> START: %d END: %d NHEADING: %+v\n", start, end, nHeading)
	if len(nHeading) > 0 && nHeading[1]+start < end {
		nHeading[0], nHeading[1] = nHeading[0]+start, nHeading[1]+start
		category = text[start:nHeading[0]]
	}

	nlIndices := WikiNumListAny.FindAllIndex(category, -1)
	logger.Debug("getDefinitions> Found %d NumList entries\n", len(nlIndices))
	nlIndicesSize := len(nlIndices)
	for i := 0; i < nlIndicesSize; i++ {
		ithIdx := AdjustIndexLW(nlIndices[i][0], category)
		if string(category[ithIdx:nlIndices[i][1]]) != "# " {
			logger.Debug("getDefinitions> Got quotation or annotation bullet. Skipping...\n")
			continue
		}

		if i+1 >= nlIndicesSize && string(category[ithIdx:nlIndices[i][1]]) == "# " {
			def := parseDefinition(nlIndices[i][1], len(category), category)
			logger.Debug("getDefinitions> [%0d] Appending %s to the definition list\n", i, string(def))
			defs = append(defs, string(def))
		}

		if i+1 < nlIndicesSize && string(category[ithIdx:nlIndices[i][1]]) == "# " {
			ith1Idx := AdjustIndexLW(nlIndices[i+1][0], category)
			def := parseDefinition(nlIndices[i][1], ith1Idx, category)
			logger.Debug("getDefinitions> [%0d] Appending %s to the definition list\n", i, string(def))
			defs = append(defs, string(def))
		}
	}

	logger.Debug("getDefinitions> Got %d definitions\n", len(defs))
	return defs
}

func parseDefinition(start int, end int, text []byte) []byte {
	def := text[start:end]
	//def = WikiNewLine.ReplaceAll(def, []byte(" "))

	// need to parse the templates in the definition
	sDef, err := wikitemplates.ParseRecursive(def)
	Check(err)

	def = []byte(sDef)
	newline := WikiNewLine.FindIndex(def)

	if len(newline) > 0 {
		def = def[:newline[0]]
	}

	def = bytes.TrimSpace(def)

	return def
}

func GetLanguageSection(text []byte, language string) []byte {
	// this is going to pull out the "section" of the text bounded by the
	// desired language heading and the following heading or the end of
	// the data.

	indices := WikiLang.FindAllIndex(text, -1)
	indicesSize := len(indices)

	logger.Debug("CORPUS: %s\n", string(text))
	logger.Debug("CORPUS SIZE: %d INDICES_SIZE: %d INDICES: %+v\n", len(text), indicesSize, indices)

	if indicesSize == 0 {
		return text
	}

	// when the match has a leading \s, remove it
	if text[indices[0][0] : indices[0][0]+1][0] == byte('\n') {
		indices[0][0]++
	}

	if indicesSize == 1 {
		// it is assumed at this point that the pages have been filterd by the
		// desired language already, which means that the only heading present
		// is the one that is wanted.
		logger.Debug("Found only 1 heading. Returning corpus for heading '%s'\n", string(text[indices[0][0]:indices[0][1]]))
		return text[indices[0][1]:]
	}

	logger.Debug("Found %d indices\n", indicesSize)
	logger.Debug("Indices: %v\n", indices)
	corpus := text
	for i := 0; i < indicesSize; i++ {
		heading := string(text[indices[i][0]:indices[i][1]])
		logger.Debug("Checking heading: %s\n", heading)

		if heading != fmt.Sprintf("==%s==", language) {
			logger.Debug("'%s' != '==%s=='\n", heading, language)
			continue
		}

		if i == indicesSize-1 {
			logger.Debug("Found last heading\n")
			return text[indices[i][1]:]
		}

		corpus = text[indices[i][1]:indices[i+1][0]]
		break
	}

	return corpus
}
