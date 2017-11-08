package hlconf2017

import (
	"strings"
)

type Pattern struct {
	Full   string
	Len    int
	Prefix Part
	PrefixString string
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
	Patterns map[string][]Pattern
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	patterns := make([]Pattern, len(allowedPatterns))

	p.Patterns = map[string][]Pattern{}
	//p.Patterns = make([]Pattern, len(allowedPatterns))

	for i, pattern := range allowedPatterns {
		for _, part := range strings.Split(pattern, ".") {
			pp := Part{
				Part: part,
			}

			if strings.Contains(pp.Part, "{") {
				raw := strings.Replace(pp.Part, "{", ",", -1)
				raw = strings.Replace(raw, "}", ",", -1)

				pparts := strings.Split(raw, ",")
				ln := len(pparts)

				if strings.Index(pp.Part, "{") != 0 {
					pp.Prefix = strings.Replace(pparts[0], "*", "", -1)
					pparts = pparts[1:]
				}

				ln = len(pparts)

				if strings.Index(pp.Part, "}") != ln-1 {
					pp.Sufix = strings.Replace(pparts[ln-1], "*", "", -1)
					pparts = pparts[:ln-1]
				}

				if pparts[0] == "" {
					pparts = pparts[1:]
				}

				pp.Or = pparts
			}

			pp.HasStart = strings.Contains(pp.Part, "*")
			pp.ClearPart = strings.Replace(pp.Part, "*", "", -1)

			patterns[i].Parts = append(patterns[i].Parts, pp)
		}

		patterns[i].Len = len(patterns[i].Parts)
		patterns[i].Full = pattern
		patterns[i].Prefix = patterns[i].Parts[0]
		patterns[i].PrefixString = patterns[i].Parts[0].Part

		//fmt.Println(p.Patterns)
	}

	for _, pp := range patterns {
		_, ok := p.Patterns[pp.PrefixString]
		if ok {
			p.Patterns[pp.PrefixString] = append(p.Patterns[pp.PrefixString], pp)
			continue
		}

		p.Patterns[pp.PrefixString] = []Pattern{pp}
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	//metricParts := strings.Split(metricName, ".")
	matchingPatterns := make([]string, 0, len(p.Patterns))
	metricParts := split(metricName, ".", -1)

	patterns, ok := p.Patterns[metricParts[0]]
	if !ok {
		return []string{}
	}

	for _, pt := range patterns {
		if pt.Len != len(metricParts) {
			continue
		}

		if metricParts[0] != pt.Prefix.Part {
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

func split(s, sep string, n int) []string {
	if n < 0 {
		n = strings.Count(s, sep) + 1
	}

	a := make([]string, n)
	n--
	i := 0
	for i < n {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}
		a[i] = s[:m]
		s = s[m+len(sep):]
		i++
	}
	a[i] = s
	return a[:i+1]
}
