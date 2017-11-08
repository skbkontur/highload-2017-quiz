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
	OrFull        []string
	ClearPart string
	HasStart  bool
}

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	Patterns [5][256][256][]Pattern
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	patterns := make([]Pattern, len(allowedPatterns))

	p.Patterns = [5][256][256][]Pattern{}
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

				for _, item := range pparts {
					pp.OrFull = append(pp.OrFull, pp.Prefix + item + pp.Sufix)
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

	for i := 0; i < 5; i++ {
		p.Patterns[i] = [256][256][]Pattern{}
	}

	for _, pp := range patterns {
		p.Patterns[pp.Len][int(pp.PrefixString[0])][int(pp.PrefixString[1])] = append(p.Patterns[pp.Len][int(pp.PrefixString[0])][int(pp.PrefixString[1])], pp)
	}
}

var matchingPatterns = [20]string{}
var metricParts = [4]string{}
var cn = 0

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	//matchingPatterns := make([]string, 0, len(p.Patterns))

	cn = strings.Count(metricName, ".") + 1

	split(&metricParts, metricName, ".", cn)

	count := 0
	for _, pt := range p.Patterns[cn][int(metricParts[0][0])][int(metricParts[0][1])] {
		if pt.Len != cn {
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
				for _, item := range part.OrFull {
					//patt := part.Prefix + item + part.Sufix

					if strings.Contains(metricParts[i], item) {
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


		matchingPatterns[count] = pt.Full

		count++
	}

	return matchingPatterns[:count]
}

func split(a *[4]string, s, sep string, n int) (*[4]string) {
	n = n+1
	n--
	i := 0
	for i < n {
		m := index(s)
		if m < 0 {
			break
		}
		a[i] = s[:m]
		s = s[m+len(sep):]
		i++
	}
	a[i] = s
	return a
}

var dot int32 = 46

func index(str string) int {
	for pos ,i := range str {
		if i == dot {
			return pos
		}
	}

	return -1
}
