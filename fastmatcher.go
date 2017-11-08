package hlconf2017

import (
	"regexp"
	"strings"
)

type MPattern struct {
	Raw    string
	Len    int
	Prefix string
	Parts  []string
	Strict bool
}

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	AllowedPatterns []string
	MYPatterns      []MPattern
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	for _, pattern := range allowedPatterns {
		mp := MPattern{}
		mp.Raw = pattern
		if pattern[len(pattern)-2:] == ".*" {
			// mp.Prefix = pattern[len(pattern)-2:]
			mp.Prefix = strings.Replace(pattern, ".*", "", -1)
		}
		mp.Parts = strings.Split(pattern, ".")
		mp.Len = len(mp.Parts)
		if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "{") {
			mp.Strict = true
		}
		p.MYPatterns = append(p.MYPatterns, mp)
	}
	p.AllowedPatterns = allowedPatterns
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	metricParts := strings.Split(metricName, ".")
NEXTPATTERN:
	for _, pattern := range p.MYPatterns {
		// на длинне бортуем сразу
		if pattern.Len != len(metricParts) {
			continue NEXTPATTERN
		}
		// есть точное совпадение
		if pattern.Raw == metricName {
			matchingPatterns = append(matchingPatterns, pattern.Raw)
			continue NEXTPATTERN
		}

		// если требовалось ТОЛЬКО точное совпадение
		if pattern.Strict {
			continue NEXTPATTERN
		}

		// если требовался префикс
		if pattern.Prefix != "" && strings.Contains(metricName, pattern.Prefix) {
			matchingPatterns = append(matchingPatterns, pattern.Raw)
			continue NEXTPATTERN
		} else {
			// если же требовали префикс а не нашли
			if pattern.Prefix != "" {
				continue NEXTPATTERN
			}
		}
		// теперь только регекспы
		for i, part := range pattern.Parts {
			regexPart := "^" + part + "$"
			regexPart = strings.Replace(regexPart, "*", ".*", -1)
			regexPart = strings.Replace(regexPart, "{", "(", -1)
			regexPart = strings.Replace(regexPart, "}", ")", -1)
			regexPart = strings.Replace(regexPart, ",", "|", -1)

			regex := regexp.MustCompile(regexPart)

			if !regex.MatchString(metricParts[i]) {
				continue NEXTPATTERN
			}
		}
		matchingPatterns = append(matchingPatterns, pattern.Raw)
	}

	return
}
