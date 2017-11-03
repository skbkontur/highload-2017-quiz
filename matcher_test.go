package hlconf2017

import (
	"testing"
)

var testPatterns = []string{
	"Simple.matching.pattern",
	"Star.single.*",
	"Star.*.double.any*",
	"Bracket.{one,two,three}.pattern",
	"Bracket.pr{one,two,three}suf",
	"Complex.matching.pattern",
	"Complex.*.*",
	"Complex.*{one,two,three}suf*.pattern",
}

var nonMatchingMetrics = []string{
	"Simple.notmatching.pattern",
	"Star.nothing",
	"Bracket.one.nothing",
	"Bracket.nothing.pattern",
	"Complex.prefixonesuffix",
}

var matchingSingleMetrics = []string{
	"Simple.matching.pattern",
	"Star.single.anything",
	"Star.anything.double.anything",
	"Bracket.one.pattern",
	"Bracket.two.pattern",
	"Bracket.three.pattern",
	"Bracket.pronesuf",
	"Bracket.prtwosuf",
	"Bracket.prthreesuf",
	"Complex.anything.pattern",
	"Complex.prefixtwofix.pattern",
	"Complex.anything.pattern",
}

var matchingMultipleMetrics = []string{
	"Complex.matching.pattern",
	"Complex.prefixonesuffix.pattern",
}

func TestPatternMatcher_DetectMatchingPatterns(t *testing.T) {
	pm := &PatternMatcher{}
	pm.InitPatterns(testPatterns)

	for _, metricName := range nonMatchingMetrics {
		if len(pm.DetectMatchingPatterns(metricName)) != 0 {
			t.Errorf("%s should not match any patterns, but it does", metricName)
		}
	}

	for _, metricName := range matchingSingleMetrics {
		if len(pm.DetectMatchingPatterns(metricName)) != 1 {
			t.Errorf("%s should match exactly one pattern, but it doesn't", metricName)
		}
	}

	for _, metricName := range matchingMultipleMetrics {
		if len(pm.DetectMatchingPatterns(metricName)) < 2 {
			t.Errorf("%s should match more than one pattern, but it doesn't", metricName)
		}
	}
}

func BenchmarkPatternMatcher_DetectMatchingPatterns(b *testing.B) {
	pm := &PatternMatcher{}
	pm.InitPatterns(testPatterns)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, metricName := range nonMatchingMetrics {
			pm.DetectMatchingPatterns(metricName)
		}
		for _, metricName := range matchingSingleMetrics {
			pm.DetectMatchingPatterns(metricName)
		}
		for _, metricName := range matchingMultipleMetrics {
			pm.DetectMatchingPatterns(metricName)
		}
	}
}
