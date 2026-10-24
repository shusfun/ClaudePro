//go:build !windows

package platform

func SystemProxy() (bool, bool, error) { return false, false, nil }
