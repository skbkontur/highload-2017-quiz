package hlconf2017

import (
	"regexp"
	"strings"
)

type Pattern struct {
	Rgs   map[string]*regexp.Regexp
	Parts []string
	Full  string

	Parts2 []Part
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
		p.P[i].Parts = strings.Split(pattern, ".")
		rgs := map[string]*regexp.Regexp{}
		for _, part := range p.P[i].Parts {
			regexPart := "^" + part + "$"
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			rgs[part] = regexp.MustCompile(regexPart)

			p.P[i].Parts2 = append(p.P[i].Parts2, Part{
				Part: part,
				Rgs:  regexp.MustCompile(regexPart),
			})
		}

		p.P[i] = Pattern{
			Parts: strings.Split(pattern, "."),
			Rgs:   rgs,
			Full:  pattern,
		}

	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")

NEXTPATTERN:
	for _, pt := range p.P {

		if len(pt.Parts) != len(metricParts) {
			continue NEXTPATTERN
		}
		for i, part := range pt.Parts2 {
			if !part.Rgs.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pt.Full)
	}

	return
}
