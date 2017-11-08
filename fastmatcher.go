package hlconf2017

import (
	"regexp"
	"strings"
)

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	AllowedPatterns map[string]*regexp.Regexp
	StaticPatterns []string
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.AllowedPatterns = make(map[string]*regexp.Regexp)

	for _, pattern := range allowedPatterns {
		if isStaticPattern(pattern) {
			p.StaticPatterns = append(p.StaticPatterns, pattern)
			continue
		}

		regexPattern := "^" + pattern + "$"
		regexPattern = strings.Replace(regexPattern, ".", "\\.", -1)
		regexPattern = strings.Replace(regexPattern, "*", ".*", -1)
		regexPattern = strings.Replace(regexPattern, "{", "(", -1)
		regexPattern = strings.Replace(regexPattern, "}", ")", -1)
		regexPattern = strings.Replace(regexPattern, ",", "|", -1)

		p.AllowedPatterns[pattern] = regexp.MustCompile(regexPattern)
	}
}

func isStaticPattern(pattern string) bool {
	return !strings.Contains(pattern, "*") && !strings.Contains(pattern, "{")
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	for _, staticPattern := range p.StaticPatterns {
		if staticPattern == metricName {
			matchingPatterns = append(matchingPatterns, staticPattern)
		}
	}

	for pattern, regex := range p.AllowedPatterns {
		if !regex.MatchString(metricName) {
			continue
		}

		matchingPatterns = append(matchingPatterns, pattern)
	}

	return
}
