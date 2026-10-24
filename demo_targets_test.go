package main

import (
	"claudepro/internal/core"
	"claudepro/internal/platform"
)

func demoPreset(id string) core.Target {
	catalog, err := platform.NewDemo(".").Catalog()
	if err != nil {
		panic(err)
	}
	for _, preset := range core.Presets(catalog, platform.Kind) {
		if preset.ID == id && preset.Unavailable == "" {
			return preset.Target
		}
	}
	panic("missing demo preset " + id)
}
