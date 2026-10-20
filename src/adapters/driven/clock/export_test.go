package clock

import "time"

// NeuMit macht neuMit für die Tests erreichbar.
func NeuMit(quelle func() time.Time) *Systemuhr { return neuMit(quelle) }
