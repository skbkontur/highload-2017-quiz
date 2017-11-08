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
		patternParts := strings.Split(pattern, ".")

		compiledParts := make([]*regexp.Regexp, len(patternParts), len(patternParts))
		for i, part := range patternParts {
			regexPart := strings.Join([]string{"^", part, "$"}, "")
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			compiledParts[i] = regexp.MustCompile(regexPart)
		}
		p.AllowedPatterns[pattern] = compiledParts
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")

NEXTPATTERN:
	for pattern, compiledParts := range p.AllowedPatterns {
		if len(compiledParts) != len(metricParts) {
			continue NEXTPATTERN
		}
		for i, compiledPart := range compiledParts {
			if !compiledPart.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pattern)
	}

	return
}
