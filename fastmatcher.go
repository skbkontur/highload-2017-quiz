package hlconf2017

import (
	"regexp"
	"strings"
)

type Pattern struct {
	Full      string
	Len       int
	UseRegexp bool

	Prefix Part

	Parts []Part

	UberRgs *regexp.Regexp
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
		str := "^"
		for j, part := range strings.Split(pattern, ".") {
			regexPart := part
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			inner := regexPart

			regexPart = "^" + regexPart + "$"

			p.P[i].Parts = append(p.P[i].Parts, Part{
				Part: part,
				Rgs:  regexp.MustCompile(regexPart),
			})

			if j != 0 {
				str += `\.` + inner
			} else {
				str += inner
			}

		}

		str += "$"

		p.P[i].Len = len(p.P[i].Parts)
		p.P[i].Full = pattern

		p.P[i].Prefix = p.P[i].Parts[0]
		p.P[i].UberRgs = regexp.MustCompile(str)
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	metricParts := strings.Split(metricName, ".")
	matchingPatterns := make([]string, 0, len(p.P))
	//l := len(metricParts)

	for _, pt := range p.P {
		if pt.Len != len(metricParts) {
			continue
		}

		if !strings.HasPrefix(metricName, pt.Prefix.Part) {
			continue
		}

		//fmt.Println(pt.Parts)
		s := true
		for i, part := range pt.Parts {
			f := false
			if part.Part == "*" {
				f = true
				continue
			}

			if part.Part == metricParts[i] {
				f = true
			}

			f = part.Rgs.MatchString(metricParts[i])

			s = s && f
		}

		if !s {
			continue
		}

		matchingPatterns = append(matchingPatterns, pt.Full)
	}

	return matchingPatterns
}
