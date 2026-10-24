package platform

import "testing"

func TestDemoCatalogCultureNames(t *testing.T) {
	catalog, err := NewDemo(t.TempDir()).Catalog()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, option := range catalog.Cultures {
		got[option.ID] = option.Name
	}
	wantID, wantName := "en-US", "英语（美国） / en-US"
	if Kind == "darwin" {
		wantID, wantName = "en_US", "英语（美国） / en_US"
	}
	if got[wantID] != wantName {
		t.Fatalf("%s name = %q, want %q", wantID, got[wantID], wantName)
	}
}
