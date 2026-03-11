package combo

import (
	"strings"
	"testing"
)

func TestUserAgent(t *testing.T) {
	ua := userAgent("test_game")

	if !strings.HasPrefix(ua, SdkName+"/"+SdkVersion) {
		t.Fatalf("user agent should start with SDK name/version, got %s", ua)
	}
	if !strings.Contains(ua, "game/test_game") {
		t.Fatalf("user agent should contain game ID, got %s", ua)
	}
	if !strings.Contains(ua, "go/") {
		t.Fatalf("user agent should contain go version, got %s", ua)
	}
	if !strings.Contains(ua, "GOOS/") {
		t.Fatalf("user agent should contain GOOS, got %s", ua)
	}
	if !strings.Contains(ua, "GOARCH/") {
		t.Fatalf("user agent should contain GOARCH, got %s", ua)
	}
}
