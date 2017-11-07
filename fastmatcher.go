package hlconf2017

import (
	"regexp"
	"strings"
	)

type (
	patternRegexps struct {
		name    string
		regexps []*regexp.Regexp
	}
	// FastPatternMatcher implements high-performance Graphite metric filtering
	FastPatternMatcher struct {
		AllowedPatterns []patternRegexps
		index map[int][]int
	}
)

var (
	r = strings.NewReplacer(
		"*", ".*",
	)
	r2 = strings.NewReplacer(
		"{", "(",
		"}", ")",
		",", "|",
	)
)

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.AllowedPatterns = make([]patternRegexps, len(allowedPatterns))
	p.index = make(map[int][]int)
	for i, pattern := range allowedPatterns {
		patternParts := strings.Split(pattern, ".")
		partsCount := len(patternParts)
		p.AllowedPatterns[i] = patternRegexps{name: pattern}
		p.index[partsCount] = append(p.index[partsCount], i)
		p.AllowedPatterns[i].regexps = make([]*regexp.Regexp, partsCount)
		for n, part := range patternParts {
			part = r.Replace(part)
			match, _ := regexp.MatchString("{", part)
			if match {
				part = r2.Replace(part)
			}
			part = strings.Join([]string{"^", part, "$"}, "")
			p.AllowedPatterns[i].regexps[n] = regexp.MustCompile(part)
		}
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")
	metricPartsLen := len(metricParts)

NEXTPATTERN:
	for _, n := range p.index[metricPartsLen] {
		patternParts := p.AllowedPatterns[n]
		for i, regex := range patternParts.regexps {
			if !regex.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, patternParts.name)
	}

	return
}
