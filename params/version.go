package params

import (
	"time"
)

var (
	Version   string
	BuildDate time.Time
	RunDate   = time.Now()
)
