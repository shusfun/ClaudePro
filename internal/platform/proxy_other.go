//go:build !windows && !darwin

package platform

func SystemProxy() (bool, bool, error) { return false, false, nil }
