//go:build darwin

package platform

import (
	"strings"
	"time"
)

// Only return presence flags, never proxy addresses, PAC URLs or credentials.
func SystemProxy() (bool, bool, error) {
	out, err := command(10*time.Second, "/usr/sbin/scutil", "--proxy")
	if err != nil {
		return false, false, err
	}
	return parseProxy(out)
}

func parseProxy(out string) (manual, pac bool, err error) {
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok || strings.TrimSpace(value) != "1" {
			continue
		}
		switch strings.TrimSpace(key) {
		case "HTTPEnable", "HTTPSEnable", "SOCKSEnable":
			manual = true
		case "ProxyAutoConfigEnable":
			pac = true
		}
	}
	return manual, pac, nil
}
