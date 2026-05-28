package tui

// Tab is one of the three top-level screens.
type Tab int

const (
	TabActivity Tab = iota
	TabMetrics
	TabStatements
)

func (t Tab) String() string {
	switch t {
	case TabActivity:
		return "Activity"
	case TabMetrics:
		return "Metrics"
	case TabStatements:
		return "Statements"
	}

	return ""
}

func (t Tab) next() Tab {
	if t == TabStatements {
		return TabActivity
	}

	return t + 1
}
