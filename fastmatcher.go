package hlconf2017

import (
	"regexp"
	"strings"
)

type (
	Pattern struct {
		Full   string
		RegExs []*regexp.Regexp
	}

	// FastPatternMatcher implements high-performance Graphite metric filtering
	FastPatternMatcher struct {
		AllowedPatterns []Pattern
	}
)

func (p *FastPatternMatcher) compileRegexPart(part string) *regexp.Regexp {
	regexPart := "^" + part + "$"
	regexPart = strings.Replace(regexPart, "*", ".*", -1)
	regexPart = strings.Replace(regexPart, "{", "(", -1)
	regexPart = strings.Replace(regexPart, "}", ")", -1)
	regexPart = strings.Replace(regexPart, ",", "|", -1)

	return regexp.MustCompile(regexPart)
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.AllowedPatterns = make([]Pattern, len(allowedPatterns))
	for i, pat := range allowedPatterns {
		parts := strings.Split(pat, `.`)

		p.AllowedPatterns[i].Full = pat
		p.AllowedPatterns[i].RegExs = make([]*regexp.Regexp, len(parts))

		for j := range parts {
			p.AllowedPatterns[i].RegExs[j] = p.compileRegexPart(parts[j])
		}
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")

NEXTPATTERN:
	for _, patternParts := range p.AllowedPatterns {
		if len(patternParts.RegExs) != len(metricParts) {
			continue NEXTPATTERN
		}
		for i, regex := range patternParts.RegExs {
			if !regex.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, patternParts.Full)
	}

	return
}
