package hlconf2017

import (
	"regexp"
	"strings"
)

type MPattern struct {
	Raw    string
	Len    int
	Prefix string
	Parts  []string
	Strict bool
}

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	AllowedPatterns []string
	MYPatterns      []MPattern
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	for _, pattern := range allowedPatterns {
		// разобьем на несколько паттернов
		if strings.Contains(pattern, "{") {
			p.ExpandPattern(pattern)
			continue
		}
		mp := MPattern{}
		mp.Raw = pattern
		if pattern[len(pattern)-2:] == ".*" {
			mp.Prefix = strings.Replace(pattern, ".*", "", -1)
		}
		mp.Parts = strings.Split(pattern, ".")
		mp.Len = len(mp.Parts)
		if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "{") {
			mp.Strict = true
		}
		p.MYPatterns = append(p.MYPatterns, mp)
	}
	p.AllowedPatterns = allowedPatterns
}

// ExpandPattern expands metric.{a,b}
func (p *FastPatternMatcher) ExpandPattern(pattern string) {
	startIndex := strings.Index(pattern, "{")
	endIndex := strings.Index(pattern, "}")
	var newPatterns []string
	for _, variant := range strings.Split(pattern[startIndex+1:endIndex], ",") {
		newPatterns = append(newPatterns, pattern[0:startIndex]+variant+pattern[endIndex+1:])
	}
	p.InitPatterns(newPatterns)
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")
NEXTPATTERN:
	for _, pattern := range p.MYPatterns {
		// на длинне бортуем сразу
		if pattern.Len != len(metricParts) {
			continue NEXTPATTERN
		}
		// есть точное совпадение
		if pattern.Raw == metricName {
			matchingPatterns = append(matchingPatterns, pattern.Raw)
			continue NEXTPATTERN
		}

		// если требовалось ТОЛЬКО точное совпадение
		if pattern.Strict {
			continue NEXTPATTERN
		}

		// если требовался префикс
		if pattern.Prefix != "" {
			if strings.Contains(metricName, pattern.Prefix) {
				matchingPatterns = append(matchingPatterns, pattern.Raw)
				continue NEXTPATTERN
			} else {
				continue NEXTPATTERN
			}
		}
		var lazyRegexps map[string]string
		lazyRegexps = make(map[string]string)
		for i, part := range pattern.Parts {
			if part == "*" {
				continue
			}
			if part == metricParts[i] {
				continue
			}
			// отложенная регулярка
			if strings.Contains(part, "{") || strings.Contains(part, "*") {
				lazyRegexps[metricParts[i]] = part
				continue
			}
			continue NEXTPATTERN
		}
		for a, reg := range lazyRegexps {
			regexPart := "^" + reg + "$"
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regex := regexp.MustCompile(regexPart)
			if !regex.MatchString(a) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pattern.Raw)
		continue NEXTPATTERN
	}

	return
}
