package hlconf2017

import (
	"strings"
)

type Pattern struct {
	Full   string
	Len    int
	Prefix Part
	Parts  []Part
}

type Part struct {
	Part      string
	Prefix    string
	Sufix     string
	Or        []string
	ClearPart string
	HasStart  bool
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
			pp := Part{
				Part: part,
			}

			if strings.Contains(pp.Part, "{") {
				raw := strings.Replace(pp.Part, "{", ",", -1)
				raw = strings.Replace(raw, "}", ",", -1)

				pparts := strings.Split(raw, ",")
				ll := len(pparts)

				if strings.Index(pp.Part, "{") != 0 {
					pp.Prefix = strings.Replace(pparts[0], "*", "", -1)
					pparts = pparts[1:]
				}

				ll = len(pparts)

				if strings.Index(pp.Part, "}") != ll-1 {
					pp.Sufix = strings.Replace(pparts[ll-1], "*", "", -1)
					pparts = pparts[:ll-1]
				}

				if pparts[0] == "" {
					pparts = pparts[1:]
				}

				pp.Or = pparts
			}

			pp.HasStart = strings.Contains(pp.Part, "*")
			pp.ClearPart = strings.Replace(pp.Part, "*", "", -1)

			p.P[i].Parts = append(p.P[i].Parts, pp)

			if j != 0 {
				str += `\.` + part
			} else {
				str += part
			}

		}

		str += "$"

		p.P[i].Len = len(p.P[i].Parts)
		p.P[i].Full = pattern
		p.P[i].Prefix = p.P[i].Parts[0]
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	metricParts := strings.Split(metricName, ".")
	matchingPatterns := make([]string, 0, len(p.P))

	for _, pt := range p.P {
		if pt.Len != len(metricParts) {
			continue
		}

		if !strings.HasPrefix(metricName, pt.Prefix.Part) {
			continue
		}

		s := true
		for i, part := range pt.Parts {
			f := false
			if part.Part == "*" {
				f = true
				continue
			}

			if part.Part == metricParts[i] {
				f = true
				continue
			}

			if len(part.Or) > 0 {
				for _, item := range part.Or {
					patt := part.Prefix + item + part.Sufix

					if strings.Contains(metricParts[i], patt) {
						f = true
						break
					}
				}

				if f {
					continue
				}
			}

			if part.HasStart {
				if strings.Contains(metricParts[i], part.ClearPart) {
					f = true
					break
				}

				if f {
					continue
				}
			}

			s = s && f
		}

		if !s {
			continue
		}

		matchingPatterns = append(matchingPatterns, pt.Full)
	}

	return matchingPatterns
}
