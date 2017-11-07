package hlconf2017

import (
	"regexp"
	"strings"
)

// FastPatternMatcher implements high-performance Graphite metric filtering
type (
	FastPatternMatcher struct {
		Patterns [][]Pattern
		Count    int
	}

	RegexpItem struct {
		isEmpty  bool
		isRegexp bool
		reg      *regexp.Regexp
		str      string
	}

	Pattern struct {
		count int
		r     []RegexpItem
		str   string
	}
)

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.Patterns = make([][]Pattern, 256)
	p.Count = len(allowedPatterns)

	for _, pattern := range allowedPatterns {
		patternParts := strings.Split(pattern, ".")

		count := len(patternParts)
		r := make([]RegexpItem, count)

		for j, patternPart := range patternParts {
			regexPart := strings.Replace(patternPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			if patternPart == "*" {
				r[j] = RegexpItem{
					isEmpty: true,
					str:     patternPart,
				}
			} else if regexPart == patternPart {
				r[j] = RegexpItem{
					isRegexp: false,
					str:      patternPart,
				}
			} else {
				regexPart := "^" + patternPart + "$"
				r[j] = RegexpItem{
					isRegexp: true,
					reg:      regexp.MustCompile(regexPart),
				}
			}

		}

		if p.Patterns[pattern[0]] == nil {
			p.Patterns[pattern[0]] = make([]Pattern, 0)
		}

		p.Patterns[pattern[0]] = append(p.Patterns[pattern[0]], Pattern{
			count: count,
			r:     r,
			str:   pattern,
		})
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	if p.Patterns[metricName[0]] == nil {
		return
	}

	metric := strings.Split(metricName, ".")

	q := 0
	matchingPatterns = make([]string, p.Count)
UP:
	for i := 0; i < len(p.Patterns[metricName[0]]); i++ {
		if p.Patterns[metricName[0]][i].count != len(metricName) {
			continue
		}

		for j := 0; j < p.Patterns[metricName[0]][i].count; i++ {
			if p.Patterns[metricName[0]][i].r[j].isEmpty {
				continue
			} else if p.Patterns[metricName[0]][i].r[j].isRegexp {
				if !p.Patterns[metricName[0]][i].r[j].reg.MatchString(metric[j]) {
					continue UP
				}
			} else {
				if p.Patterns[metricName[0]][i].r[j].str != metric[i] {
					continue UP
				}
			}
		}

		matchingPatterns[q] = p.Patterns[metricName[0]][i].str
		q++
	}

	if q > 0 {
		matchingPatterns = matchingPatterns[0 : q-1]
	}

	return
}
