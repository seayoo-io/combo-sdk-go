package combo

import (
	"fmt"
	"runtime"
	"strings"
)

// Example:
// combo-sdk-go/0.1.0 game/xcom go/1.21.5 GOOS/linux GOARCH/arm64
func userAgent(gameId GameId) string {
	goVersion := strings.TrimPrefix(runtime.Version(), "go")
	return fmt.Sprintf("%s/%s game/%s go/%s GOOS/%s GOARCH/%s",
		SdkName, SdkVersion, gameId, goVersion, runtime.GOOS, runtime.GOARCH)
}
