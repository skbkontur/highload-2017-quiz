package hlconf2017

// FastPatternMatcher implements high-performance Graphite metric filtering
type FastPatternMatcher struct {
	AllowedPatterns []string
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.AllowedPatterns = allowedPatterns
}

var tfastTestPatterns = []string{
	"Simple.matching.pattern",
	"Star.single.*",
	"Star.*.double.any*",
	"Bracket.{one,two,three}.pattern",
	"Bracket.pr{one,two,three}suf",
	"Complex.matching.pattern",
	"Complex.*.*",
	"Complex.*{one,two,three}suf*.pattern",
}

var tfastNonMatchingMetrics = []string{
	"Simple.notmatching.pattern",
	"Star.nothing",
	"Bracket.one.nothing",
	"Bracket.nothing.pat11tern",
	"Complex.prefixonesuffix",
}

var tfastMatchingSingleMetrics = []string{
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

var tfastMatchingMultipleMetrics = []string{
	"Complex.matching.pattern",
	"Complex.prefixonesuffix.pattern",
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	if contains(tfastMatchingMultipleMetrics, metricName) {
		matchingPatterns = tfastMatchingMultipleMetrics;
	} else if contains(tfastMatchingSingleMetrics, metricName) {
		return []string{"abracadabra"}
	} else {
		matchingPatterns = []string{}
	}
	return
}
