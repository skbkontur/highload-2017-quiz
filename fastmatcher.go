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
	// your own fast implementation here
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	// your own fast implementation here
	return
}
