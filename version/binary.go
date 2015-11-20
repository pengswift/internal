package version

import (
	"fmt"
	"runtime"
)

func String(app string, version string) string {
	return fmt.Sprintf("%s v%s (built w/%s)", app, version, runtime.Version())
}
