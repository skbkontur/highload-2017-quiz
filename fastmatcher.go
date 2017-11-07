package hlconf2017

import (
	"regexp"
	"strings"
)

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	AllowedPatterns []string
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.AllowedPatterns = allowedPatterns
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")
	lenMetricParts := len(metricParts)

	bracesRegex := regexp.MustCompile(`{(.*)}`)

NEXTPATTERN:
	for _, pattern := range p.AllowedPatterns {
		patternParts := strings.Split(pattern, ".")
		if len(patternParts) != lenMetricParts {
			continue NEXTPATTERN
		}
		for i, part := range patternParts {
			regexPart := "^" + part + "$"
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regexPart = bracesRegex.ReplaceAllString(regexPart, "($1)")
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			regex := regexp.MustCompile(regexPart)

			if !regex.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pattern)
	}

	return
}
