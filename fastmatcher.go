package hlconf2017

import (
	"regexp"
	"strconv"
	"strings"
)

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	PatternTrieRoot *TrieVertex
}

type TrieVertex struct {
	token      string
	pattern    string         // Only for terminal vertices
	regexp     *regexp.Regexp // Only for regexp vertices
	isWildcard bool
	isRegexp   bool
	isTerminal bool
	children   []*TrieVertex
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	patterns := expandPatterns(allowedPatterns)

	p.PatternTrieRoot = &TrieVertex{
		isTerminal: false,
	}

	createTrie(p.PatternTrieRoot, patterns)
	//printTrie(p.PatternTrieRoot, 0)
}

func expandPatterns(allowedPatterns []string) (patterns []string) {
	for _, pattern := range allowedPatterns {
		startIndex := strings.Index(pattern, "{")

		if startIndex == -1 {
			patterns = append(patterns, pattern)
		} else {
			endIndex := strings.Index(pattern, "}")
			offset := startIndex

			for _, alternative := range strings.Split(pattern[startIndex + 1:endIndex], ",") {
				patterns = append(patterns, pattern[0:offset] + alternative + pattern[endIndex + 1:])
			}
		}
	}

	return patterns
}

func createTrie(root *TrieVertex, patterns []string) {
	for _, pattern := range patterns {
		vertex := root

		for _, part := range strings.Split(pattern, ".") {
			hasSuitableChild := false

			for _, child := range vertex.children {
				if child.token == part {
					vertex = child
					hasSuitableChild = true
					break
				}
			}

			if !hasSuitableChild {
				newChild := &TrieVertex{
					token:      part,
					isWildcard: part == "*",
				}

				if !newChild.isWildcard {
					newChild.isRegexp = strings.Contains(part, "*") || strings.Contains(part, "{")
				}

				if newChild.isRegexp {
					newChild.regexp = getRegexp(part)
				}

				vertex.children = append(vertex.children, newChild)
				vertex = newChild
			}
		}

		vertex.pattern = pattern
		vertex.isTerminal = true
	}
}

func getRegexp(part string) *regexp.Regexp {
	part = strings.Replace(part, "*", ".*", -1)
	part = strings.Replace(part, "{", "(", -1)
	part = strings.Replace(part, ",", "|", -1)
	part = strings.Replace(part, "}", ")", -1)
	expr, _ := regexp.Compile(part)
	return expr
}

func printTrie(vertex *TrieVertex, level int) {
	var name string

	if vertex.isRegexp {
		name = "/" + vertex.regexp.String() + "/"
	} else {
		name = "-" + vertex.token + "-"
	}

	levelOffset := strings.Repeat("   ", level)
	childrenCount := "c:" + strconv.Itoa(len(vertex.children))
	isTerminal := "t:" + strconv.FormatBool(vertex.isTerminal)

	println(levelOffset + " " + name + " " + childrenCount + " " + isTerminal)

	for _, child := range vertex.children {
		printTrie(child, level+1)
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	matchingPatterns = make([]string, 0, )
	return detectMatchingPatterns(p.PatternTrieRoot, strings.Split(metricName, "."))
}

func detectMatchingPatterns(vertex *TrieVertex, parts []string) (matchingPatterns []string) {
	for i, part := range parts {
		for _, child := range vertex.children {
			if child.token == part || child.isWildcard || child.isRegexp && child.regexp.MatchString(part) {
				if child.isTerminal {
					matchingPatterns = append(matchingPatterns, vertex.pattern)
				} else {
					patterns := detectMatchingPatterns(child, parts[i+1:])
					matchingPatterns = append(matchingPatterns, patterns...)
				}
			}
		}
	}

	return
}
