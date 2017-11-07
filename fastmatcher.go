package hlconf2017

import (
	"regexp"
	"strings"
)

// FastPatternMatcher implements high-performance Graphite metric filtering
type (
	FastPatternMatcher struct {
		Checkers [][]Checker
		Patterns [][]Pattern
		Count    int
	}

	RegexpItem struct {
		isEmpty  bool
		isRegexp bool
		reg      func(string) bool
		str      string
		pat      string
		pattern  string
	}

	Pattern struct {
		count int
		r     []RegexpItem
		str   string
		parts []string
	}

	Checker struct {
		level    int
		rItem    *RegexpItem
		children []Checker
	}
)

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.Patterns = make([][]Pattern, 256)
	p.Checkers = make([][]Checker, 256)
	re := regexp.MustCompile("^([^{]+)({[^}]+})(.*)$")

	patternsMap := map[string]string{}

	for {
		replaced := false
		tmpAllowedPatterns := make([]string, 0)
		for _, pattern := range allowedPatterns {
			matches := re.FindAllStringSubmatch(pattern, -1)
			if len(matches) == 1 {
				matches[0][2] = strings.Replace(matches[0][2], "{", "", -1)
				matches[0][2] = strings.Replace(matches[0][2], "}", "", -1)

				tokens := strings.Split(matches[0][2], ",")

				for _, token := range tokens {
					tmpAllowedPatterns = append(tmpAllowedPatterns, matches[0][1]+token+matches[0][3])

					if _, ok := patternsMap[matches[0][1]+token+matches[0][3]]; !ok {
						patternsMap[matches[0][1]+token+matches[0][3]] = pattern
					}
				}

				replaced = true
			} else {
				tmpAllowedPatterns = append(tmpAllowedPatterns, pattern)
				if _, ok := patternsMap[pattern]; !ok {
					patternsMap[pattern] = pattern
				}
			}
		}

		allowedPatterns = tmpAllowedPatterns

		if !replaced {
			break
		}
	}

	p.Count = len(allowedPatterns)

	for _, pattern := range allowedPatterns {
		patternParts := strings.Split(pattern, ".")

		count := len(patternParts)
		r := make([]RegexpItem, count)

		for j, patternPart := range patternParts {
			regexPart := strings.Replace(patternPart, "*", ".*", -1)

			lastPattern := ""
			if j == len(patternParts)-1 {
				lastPattern = patternsMap[pattern]
			}

			pat := patternPart
			if patternPart == "*" {
				r[j] = RegexpItem{
					isEmpty: true,
					str:     patternPart,
					pat:     pat,
					pattern: lastPattern,
				}
			} else if regexPart == patternPart {
				r[j] = RegexpItem{
					isRegexp: false,
					str:      patternPart,
					pat:      pat,
					pattern:  lastPattern,
				}
			} else {
				regexPart = "^" + regexPart + "$"

				parts := strings.Split(pat, "*")
				leadingGlob := strings.HasPrefix(pat, "*")
				trailingGlob := strings.HasSuffix(pat, "*")
				end := len(parts) - 1

				r[j] = RegexpItem{
					isRegexp: true,
					pat:      pat,
					pattern:  lastPattern,
					reg: func(subj string) bool {
						if len(parts) == 1 {
							return subj == pat
						}

						for i := 0; i < end; i++ {
							idx := strings.Index(subj, parts[i])

							switch i {
							case 0:
								if !leadingGlob && idx != 0 {
									return false
								}
							default:
								if idx < 0 {
									return false
								}
							}

							subj = subj[idx+len(parts[i]):]
						}

						return trailingGlob || strings.HasSuffix(subj, parts[end])
					},
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
			parts: patternParts,
		})
	}

	for i := 0; i < len(p.Patterns); i++ {
		if p.Patterns[i] == nil {
			continue
		}

		for j := 0; j < len(p.Patterns[i]); j++ {
			if p.Checkers[p.Patterns[i][j].count] == nil {
				p.Checkers[p.Patterns[i][j].count] = make([]Checker, 0)
			}

			list := &p.Checkers[p.Patterns[i][j].count]

			for q := 0; q < len(p.Patterns[i][j].r); q++ {
				list = addChecker(list, q, &p.Patterns[i][j].r[q])
			}
		}
	}
}

func addChecker(arr *[]Checker, level int, rItem *RegexpItem) *[]Checker {
	for i := 0; i < len(*arr); i++ {
		if (*arr)[i].rItem.pat == rItem.pat && level == (*arr)[i].level {
			return &(*arr)[i].children
		}
	}

	*arr = append(*arr, Checker{
		level:    level,
		rItem:    rItem,
		children: make([]Checker, 0),
	})

	return &(*arr)[len(*arr)-1].children
}

func getMatches(arr *[]Checker, metric *[]string, matches *[]string, index *int) {
	for i := 0; i < len(*arr); i++ {
		if (*arr)[i].rItem.isEmpty {
			if (*arr)[i].rItem.pattern != "" {
				(*matches)[*index] = (*arr)[i].rItem.pattern
				*index++
			} else {
				if len((*arr)[i].children) > 0 {
					getMatches(&(*arr)[i].children, metric, matches, index)
				}
			}
		} else if (*arr)[i].rItem.isRegexp {
			if (*arr)[i].rItem.reg((*metric)[(*arr)[i].level]) {
				if (*arr)[i].rItem.pattern != "" {
					(*matches)[*index] = (*arr)[i].rItem.pattern
					*index++
				} else {
					if len((*arr)[i].children) > 0 {
						getMatches(&(*arr)[i].children, metric, matches, index)
					}
				}
			}
		} else {
			if (*arr)[i].rItem.str == (*metric)[(*arr)[i].level] {
				if (*arr)[i].rItem.pattern != "" {
					(*matches)[*index] = (*arr)[i].rItem.pattern
					*index++
				} else {
					if len((*arr)[i].children) > 0 {
						getMatches(&(*arr)[i].children, metric, matches, index)
					}
				}
			}
		}
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) []string {
	matches := make([]string, p.Count, p.Count)

	if p.Patterns[metricName[0]] == nil {
		return matches
	}

	metric := strings.Split(metricName, ".")
	ind := 0

	getMatches(&p.Checkers[len(metric)], &metric, &matches, &ind)

	return matches[0:ind]
}
