package hlconf2017

import (
	"regexp"
	"strings"
)

type Pattern struct {
	Full  string
	Len   int
	Parts []Part
}

type Part struct {
	Part string
	Rgs  *regexp.Regexp
}

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	P []Pattern
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.P = make([]Pattern, len(allowedPatterns))

	for i, pattern := range allowedPatterns {
		for _, part := range strings.Split(pattern, ".") {
			regexPart := "^" + part + "$"
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			p.P[i].Parts = append(p.P[i].Parts, Part{
				Part: part,
				Rgs:  regexp.MustCompile(regexPart),
			})
		}
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	metricParts := strings.Split(metricName, ".")

	matchingPatterns := make([]string, 0, 100)

NEXTPATTERN:
	for _, pt := range p.P {

		if pt.Len != len(metricParts) {
			continue NEXTPATTERN
		}
		for i, part := range pt.Parts {
			if !part.Rgs.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pt.Full)
	}

	return matchingPatterns
}
