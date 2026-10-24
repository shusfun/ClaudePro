package core

type Preset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Target      Target `json:"target"`
	Unavailable string `json:"unavailable"`
}

func Presets(c Catalog) []Preset {
	items := []Preset{
		{"us-west", "美国西部", "US", Target{"Pacific Standard Time", "en-US", "244"}, ""},
		{"us-east", "美国东部", "US", Target{"Eastern Standard Time", "en-US", "244"}, ""},
		{"uk", "英国", "GB", Target{"GMT Standard Time", "en-GB", "242"}, ""},
		{"germany", "德国", "DE", Target{"W. Europe Standard Time", "de-DE", "94"}, ""},
		{"japan", "日本", "JP", Target{"Tokyo Standard Time", "ja-JP", "122"}, ""},
		{"singapore", "新加坡", "SG", Target{"Singapore Standard Time", "en-SG", "215"}, ""},
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
				items[i].Unavailable = "本机 Windows 缺少选项：" + part.value
				break
			}
		}
	}
	return items
}
