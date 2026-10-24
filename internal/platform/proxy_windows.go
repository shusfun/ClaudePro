//go:build windows

package platform

import "golang.org/x/sys/windows/registry"

// Only return presence flags, never proxy addresses, PAC URLs or credentials.
func SystemProxy() (bool, bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return false, false, err
	}
	defer k.Close()
	enabled, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil && err != registry.ErrNotExist {
		return false, false, err
	}
	pac, _, err := k.GetStringValue("AutoConfigURL")
	if err != nil && err != registry.ErrNotExist {
		return false, false, err
	}
	return enabled != 0, pac != "", nil
}
