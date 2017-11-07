package hlconf2017

import (
	"regexp"
	"strings"
)

type Pattern struct {
	Full string
	Len  int

	Part1 Part
	Part2 Part
	Part3 Part

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

		p.P[i].Part1 = p.P[i].Parts[0]
		p.P[i].Part2 = p.P[i].Parts[1]
		if len(p.P[i].Parts) == 3 {
			p.P[i].Part3 = p.P[i].Parts[2]
		}
	}
}

var (
	matchingPatterns = make([]string, 0, 50)
	l                int
	metricParts      []string
	pos              int
)

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	metricParts = strings.Split(metricName, ".")
	matchingPatterns = []string{}
	l = len(metricParts)

	pos = 0
	for _, pt := range p.P {
		if pt.Len != l {
			continue
		}

		if l == 3 {
			if !pt.Part3.Rgs.MatchString(metricParts[2]) {
				continue
			}
		}

		if !pt.Part2.Rgs.MatchString(metricParts[0]) {
			continue
		}

		if !pt.Part1.Rgs.MatchString(metricParts[1]) {
			continue
		}

		matchingPatterns[pos] = pt.Full
		pos++
	}

	return matchingPatterns
}
