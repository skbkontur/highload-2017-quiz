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

		p.P[i].Len = len(p.P[i].Parts)
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	metricParts := strings.Split(metricName, ".")

	matchingPatterns := make([]string, 0, 100)

	for _, pt := range p.P {

		if pt.Len != len(metricParts) {
			continue
		}

		if !pt.Parts[0].Rgs.MatchString(metricParts[0]) {
			continue
		}

		if !pt.Parts[1].Rgs.MatchString(metricParts[1]) {
			continue
		}

		if len(pt.Parts) == 3 {
			if !pt.Parts[2].Rgs.MatchString(metricParts[2]) {
				continue
			}
		}

		matchingPatterns = append(matchingPatterns, pt.Full)
	}

	return matchingPatterns
}
