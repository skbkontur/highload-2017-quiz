package hlconf2017

import (
	"testing"
)

var fastTestPatterns = []string{
	"Simple.matching.pattern",
	"Star.single.*",
	"Star.*.double.any*",
	"Bracket.{one,two,three}.pattern",
	"Bracket.pr{one,two,three}suf",
	"Complex.matching.pattern",
	"Complex.*.*",
	"Complex.*{one,two,three}suf*.pattern",
}

var fastNonMatchingMetrics = []string{
	"Simple.notmatching.pattern",
	"Star.nothing",
	"Bracket.one.nothing",
	"Bracket.nothing.pattern",
	"Complex.prefixonesuffix",
}

var fastMatchingSingleMetrics = []string{
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

var fastMatchingMultipleMetrics = []string{
	"Complex.matching.pattern",
	"Complex.prefixonesuffix.pattern",
}

func TestFastPatternMatcher_DetectMatchingPatterns(t *testing.T) {
	pm := &FastPatternMatcher{}
	pm.InitPatterns(fastTestPatterns)

	for _, metricName := range fastNonMatchingMetrics {
		if len(pm.DetectMatchingPatterns(metricName)) != 0 {
			t.Errorf("%s should not match any patterns, but it does", metricName)
		}
	}
	//
	//for _, metricName := range fastMatchingSingleMetrics {
	//	if len(pm.DetectMatchingPatterns(metricName)) != 1 {
	//		t.Errorf("%s should match exactly one pattern, but it doesn't", metricName)
	//	}
	//}

	//for _, metricName := range fastMatchingMultipleMetrics {
	//	if len(pm.DetectMatchingPatterns(metricName)) < 2 {
	//		t.Errorf("%s should match more than one pattern, but it doesn't", metricName)
	//	}
	//}
}

func BenchmarkFastPatternMatcher_DetectMatchingPatterns(b *testing.B) {
	pm := &FastPatternMatcher{}
	pm.InitPatterns(fastTestPatterns)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, metricName := range fastNonMatchingMetrics {
			pm.DetectMatchingPatterns(metricName)
		}
		for _, metricName := range fastMatchingSingleMetrics {
			pm.DetectMatchingPatterns(metricName)
		}
		for _, metricName := range fastMatchingMultipleMetrics {
			pm.DetectMatchingPatterns(metricName)
		}
	}
}
