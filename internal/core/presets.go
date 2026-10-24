package core

type Preset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Target      Target `json:"target"`
	Unavailable string `json:"unavailable"`
}

// 每个模板同时记录 Windows 与 macOS 的系统标识，按宿主平台取用。
type presetSpec struct {
	id, name, country string
	windows, darwin   Target
}

var presetSpecs = []presetSpec{
	{"us-west", "美国西部", "US", Target{"Pacific Standard Time", "en-US", "244"}, Target{"America/Los_Angeles", "en_US", "US"}},
	{"us-east", "美国东部", "US", Target{"Eastern Standard Time", "en-US", "244"}, Target{"America/New_York", "en_US", "US"}},
	{"uk", "英国", "GB", Target{"GMT Standard Time", "en-GB", "242"}, Target{"Europe/London", "en_GB", "GB"}},
	{"germany", "德国", "DE", Target{"W. Europe Standard Time", "de-DE", "94"}, Target{"Europe/Berlin", "de_DE", "DE"}},
	{"japan", "日本", "JP", Target{"Tokyo Standard Time", "ja-JP", "122"}, Target{"Asia/Tokyo", "ja_JP", "JP"}},
	{"singapore", "新加坡", "SG", Target{"Singapore Standard Time", "en-SG", "215"}, Target{"Asia/Singapore", "en_SG", "SG"}},
}

// Presets 返回内置模板。kind 为 "darwin" 时使用 macOS 标识，其余使用 Windows 标识。
func Presets(c Catalog, kind string) []Preset {
	items := make([]Preset, 0, len(presetSpecs))
	for _, s := range presetSpecs {
		target := s.windows
		if kind == "darwin" {
			target = s.darwin
		}
		items = append(items, Preset{s.id, s.name, s.country, target, ""})
	}
	for i := range items {
		for _, part := range []struct {
			value   string
			options []Option
		}{{items[i].Target.Timezone, c.Timezones}, {items[i].Target.Culture, c.Cultures}, {items[i].Target.Region, c.Regions}} {
			found := false
			for _, o := range part.options {
				if o.ID == part.value {
					found = true
					break
				}
			}
			if !found {
				items[i].Unavailable = "本机系统缺少选项：" + part.value
				break
			}
		}
	}
	return items
}
