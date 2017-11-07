package hlconf2017

import (
	"regexp"
	"strings"
)

type (
	patternRegexps struct {
		name    string
		regexps []*regexp.Regexp
		count   int
	}
	// FastPatternMatcher implements high-performance Graphite metric filtering
	FastPatternMatcher struct {
		AllowedPatterns []patternRegexps
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
	for i, pattern := range allowedPatterns {
		patternParts := strings.Split(pattern, ".")
		p.AllowedPatterns[i] = patternRegexps{name: pattern, count: len(patternParts)}
		p.AllowedPatterns[i].regexps = make([]*regexp.Regexp, p.AllowedPatterns[i].count)
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
	for _, patternParts := range p.AllowedPatterns {
		if patternParts.count != metricPartsLen {
			continue NEXTPATTERN
		}
		for i, regex := range patternParts.regexps {
			if !regex.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, patternParts.name)
	}

	return
}
