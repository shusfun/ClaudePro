package platform

import (
	"claudepro/internal/core"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Demo uses its own directory and never invokes system settings or clients.
type Demo struct {
	mu  sync.Mutex
	dir string
}

func NewDemo(dir string) *Demo { return &Demo{dir: dir} }
func (d *Demo) Read() (core.Environment, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	env := core.Environment{}
	err := core.ReadJSON(filepath.Join(d.dir, "environment.json"), &env)
	if errors.Is(err, os.ErrNotExist) {
		if Kind == "darwin" {
			return core.Environment{
				"timezone": {Label: "Asia/Shanghai", Data: map[string]string{"id": "Asia/Shanghai", "auto": "1"}},
				"culture":  {Label: "zh_CN", Data: map[string]string{"value": "zh_CN", "present": "1"}},
				"region":   {Label: "CN", Data: map[string]string{"value": "CN", "present": "1"}},
			}, nil
		}
		return core.Environment{
			"timezone": {Label: "China Standard Time", Data: map[string]string{"id": "China Standard Time"}},
			"culture":  {Label: "zh-CN", Data: map[string]string{"LocaleName": "zh-CN", "sShortDate": "yyyy/MM/dd"}},
			"region":   {Label: "45", Data: map[string]string{"Nation": "45", "Name": "CN"}},
		}, nil
	}
	return env, err
}
func cultureName(id string) string {
	names := map[string]string{
		"zh-CN": "中文（中国） / zh-CN", "zh_CN": "中文（中国） / zh_CN",
		"en-US": "英语（美国） / en-US", "en_US": "英语（美国） / en_US",
		"en-GB": "英语（英国） / en-GB", "en_GB": "英语（英国） / en_GB",
		"de-DE": "德语（德国） / de-DE", "de_DE": "德语（德国） / de_DE",
		"ja-JP": "日语（日本） / ja-JP", "ja_JP": "日语（日本） / ja_JP",
		"en-SG": "英语（新加坡） / en-SG", "en_SG": "英语（新加坡） / en_SG",
	}
	if name := names[id]; name != "" {
		return name
	}
	return id
}

func (d *Demo) Catalog() (core.Catalog, error) {
	catalog := core.Catalog{
		Timezones: []core.Option{{ID: "China Standard Time", Name: "(UTC+08:00) 北京"}, {ID: "Eastern Standard Time", Name: "(UTC-05:00) 美国东部"}, {ID: "Pacific Standard Time", Name: "(UTC-08:00) 美国西部"}},
		Cultures:  []core.Option{{ID: "zh-CN", Name: cultureName("zh-CN")}, {ID: "en-US", Name: cultureName("en-US")}},
		Regions:   []core.Option{{ID: "45", Name: "中国 (CN)"}, {ID: "244", Name: "美国 (US)"}},
	}
	if Kind == "darwin" {
		catalog = core.Catalog{
			Timezones: []core.Option{{ID: "Asia/Shanghai", Name: "(UTC+08:00) Asia/Shanghai"}},
			Cultures:  []core.Option{{ID: "zh_CN", Name: cultureName("zh_CN")}},
			Regions:   []core.Option{{ID: "CN", Name: "中国 (CN)"}},
		}
	}
	for _, preset := range core.Presets(core.Catalog{}, Kind) {
		catalog.Timezones = append(catalog.Timezones, core.Option{ID: preset.Target.Timezone, Name: preset.Name + " / " + preset.Target.Timezone})
		catalog.Cultures = append(catalog.Cultures, core.Option{ID: preset.Target.Culture, Name: cultureName(preset.Target.Culture)})
		catalog.Regions = append(catalog.Regions, core.Option{ID: preset.Target.Region, Name: preset.Name + " (" + preset.Country + ")"})
	}
	dedup := func(options []core.Option) []core.Option {
		seen := map[string]bool{}
		result := []core.Option{}
		for _, o := range options {
			if !seen[o.ID] {
				seen[o.ID] = true
				result = append(result, o)
			}
		}
		return result
	}
	catalog.Timezones = dedup(catalog.Timezones)
	catalog.Cultures = dedup(catalog.Cultures)
	catalog.Regions = dedup(catalog.Regions)
	return catalog, nil
}
func (d *Demo) Validate(t core.Target) error {
	c, _ := d.Catalog()
	for _, x := range []struct {
		v    string
		list []core.Option
	}{{t.Timezone, c.Timezones}, {t.Culture, c.Cultures}, {t.Region, c.Regions}} {
		found := false
		for _, o := range x.list {
			if o.ID == x.v {
				found = true
			}
		}
		if !found {
			return errors.New("请选择有效的模拟环境")
		}
	}
	return nil
}
func (d *Demo) Apply(field, value string) error {
	env, err := d.Read()
	if err != nil {
		return err
	}
	env[field] = core.Value{Label: value, Data: map[string]string{"value": value}}
	d.mu.Lock()
	defer d.mu.Unlock()
	return core.AtomicJSON(filepath.Join(d.dir, "environment.json"), env)
}
func (d *Demo) Restore(field string, value core.Value) error {
	env, err := d.Read()
	if err != nil {
		return err
	}
	env[field] = value
	d.mu.Lock()
	defer d.mu.Unlock()
	return core.AtomicJSON(filepath.Join(d.dir, "environment.json"), env)
}
func (d *Demo) Running() ([]string, error) { return []string{}, nil }
func (d *Demo) Discover() (core.Clients, error) {
	return core.Clients{CLI: "模拟 Claude CLI", Desktop: "模拟 Claude Desktop", DesktopKind: "demo"}, nil
}
func (d *Demo) LaunchCLI(string, string) error {
	return errors.New("模拟模式不会启动真实客户端")
}
func (d *Demo) LaunchDesktop(string, string) error {
	return errors.New("模拟模式不会启动真实客户端")
}
