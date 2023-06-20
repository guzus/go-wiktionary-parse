package lib

import "regexp"

var (
	// regex pointers
	WikiLang        *regexp.Regexp = regexp.MustCompile(`(\s==|^==)[\w\s]+==`)          // most languages are a single word; there are some that are multiple words
	WikiLexM        *regexp.Regexp = regexp.MustCompile(`(\s====|^====)[\w\s]+====`)    // lexical category could be multi-word (e.g. "Proper Noun") match for multi-etymology
	WikiLexS        *regexp.Regexp = regexp.MustCompile(`(\s===|^===)[\w\s]+===`)       // lexical category match for single etymology
	WikiEtymologyS  *regexp.Regexp = regexp.MustCompile(`(\s===|^===)Etymology===`)     // check for singular etymology
	WikiEtymologyM  *regexp.Regexp = regexp.MustCompile(`(\s===|^===)Etymology \d+===`) // these heading may or may not have a number designation
	WikiTranslation *regexp.Regexp = regexp.MustCompile(`(\s====|^====)Translations====`)
	WikiNumListAny  *regexp.Regexp = regexp.MustCompile(`\s##?[\*:]*? `)     // used to find all num list indices
	WikiNumList     *regexp.Regexp = regexp.MustCompile(`\s#[^:*] `)         // used to find the num list entries that are of concern
	WikiGenHeading  *regexp.Regexp = regexp.MustCompile(`(\s=+|^=+)[\w\s]+`) // generic heading search
	WikiNewLine     *regexp.Regexp = regexp.MustCompile(`\n`)
	WikiBracket     *regexp.Regexp = regexp.MustCompile(`[\[\]]+`)
	WikiWordAlt     *regexp.Regexp = regexp.MustCompile(`\[\[([\w\s]+)\|[\w\s]+\]\]`)
	WikiModifier    *regexp.Regexp = regexp.MustCompile(`\{\{m\|\w+\|([\w\s]+)\}\}`)
	WikiLabel       *regexp.Regexp = regexp.MustCompile(`\{\{(la?b?e?l?)\|\w+\|([\w\s\|'",;\(\)_\[\]-]+)\}\}`)
	WikiTplt        *regexp.Regexp = regexp.MustCompile(`\{\{|\}\}`) // open close template bounds "{{ ... }}"
	WikiExample     *regexp.Regexp = regexp.MustCompile(`\{\{examples(.+)\}\}`)
	//wikiRefs       *regexp.Regexp = regexp.MustCompile(`\<ref\>(.*?)\</ref\>`)
	HtmlBreak *regexp.Regexp = regexp.MustCompile(`\<br\>`)
)
