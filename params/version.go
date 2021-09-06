package params

import (
	"fmt"
	"time"
)

var (
	BuildDate time.Time
	RunDate   = time.Now()
)

const (
	VersionMajor = 0          // Major version component of the current release
	VersionMinor = 0          // Minor version component of the current release
	VersionPatch = 1          // Patch version component of the current release
	VersionMeta  = "Unstable" // Version metadata to append to the version string
)

var Version = fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
var MetaVersion = fmt.Sprintf("%s-%s", Version, VersionMeta)
