package hlconf2017

import (
	"regexp"
	"strings"
)

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	AllowedPatterns map[string][]*regexp.Regexp
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	
	p.AllowedPatterns = make(map[string][]*regexp.Regexp)

	for _, pattern := range allowedPatterns {
		parts := strings.Split(pattern, ".")
		p.AllowedPatterns[pattern] = make([]*regexp.Regexp, len(parts))
		
		for i, part := range parts {
			initialRegexPart := "^" + part + "$"
			regexPart := strings.Replace(initialRegexPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)
			
			if regexPart == initialRegexPart {
				p.AllowedPatterns[pattern][i] = nil
			} else {
				p.AllowedPatterns[pattern][i] = regexp.MustCompile(regexPart)
			}
		}
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")

NEXTPATTERN:
	for pattern, patternParts := range p.AllowedPatterns {
		patternStringParts := strings.Split(pattern, ".")
		if len(patternParts) != len(metricParts) {
			continue
		}
		for i, part := range patternParts {
			if part == nil {
				if patternStringParts[i] == metricParts[i] {
					continue
				} else {
					continue NEXTPATTERN
				}
			if !part.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pattern)
	}

	return
}
