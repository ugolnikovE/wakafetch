package main

// StatUnit is the time spent on a single entry,
// such as a language, an editor or a project.
type StatUnit struct {
	Name         string  `json:"name"`
	Percent      float64 `json:"percent"`
	TotalSeconds float64 `json:"total_seconds"`
}

// WakafetchStats holds the statistics displayed by wakafetch.
// It does not depend on the format of the API response.
type WakafetchStats struct {
	TotalSeconds float64

	// AIPercentage is the share of time in the "ai coding" category.
	AIPercentage float64

	// Names of the entries with the most time, empty if there are none.
	TopProject string
	TopOS      string
	TopMachine string
	TopEditor  string

	Languages []StatUnit
}
