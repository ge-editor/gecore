package pathscope

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSettingsFor(t *testing.T) {
	var s Settings
	if got := s.For(FeatureMark); got != All {
		t.Errorf("builtin default = %q, want all", got)
	}
	s.Set("", Current)
	if got := s.For(FeatureMark); got != Current {
		t.Errorf("default = %q, want current", got)
	}
	s.Set(FeatureMark, Workspaces)
	if got := s.For(FeatureMark); got != Workspaces {
		t.Errorf("feature = %q, want workspaces", got)
	}
	s.Set(FeatureMark, "") // inherit
	if got := s.For(FeatureMark); got != Current {
		t.Errorf("after inherit = %q, want current", got)
	}
	// An unknown value is skipped, not used.
	s.Set(FeatureMark, "future-kind")
	if got := s.For(FeatureMark); got != Current {
		t.Errorf("unknown feature value = %q, want current", got)
	}
	s.Set("", "")
	s.Set(FeatureMark, "")
	if !s.IsZero() {
		t.Errorf("settings should be zero: %+v", s)
	}
}

func TestSettingsJSONRoundTrip(t *testing.T) {
	in := `{"default":"current","mark":"all","future":"somekind","other":"workspaces"}`
	var s Settings
	if err := json.Unmarshal([]byte(in), &s); err != nil {
		t.Fatal(err)
	}
	if s.Default != Current || s.ByFeature[FeatureMark] != All {
		t.Fatalf("unexpected settings: %+v", s)
	}
	out, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	// Keys are sorted and unknown values are kept.
	want := `{"default":"current","future":"somekind","mark":"all","other":"workspaces"}`
	if string(out) != want {
		t.Errorf("marshal = %s, want %s", out, want)
	}
	if got := s.Invalid(); !reflect.DeepEqual(got, []string{"future=somekind"}) {
		t.Errorf("Invalid() = %v", got)
	}
}

func TestSettingsUnmarshalIgnoresBadEntries(t *testing.T) {
	var s Settings
	if err := json.Unmarshal([]byte(`{"default":1,"mark":"","":"all","x":null,"y":"current"}`), &s); err != nil {
		t.Fatal(err)
	}
	want := Settings{ByFeature: map[Feature]Kind{"y": Current}}
	if !reflect.DeepEqual(s, want) {
		t.Errorf("got %+v, want %+v", s, want)
	}
	if err := json.Unmarshal([]byte(`[1]`), &s); err == nil {
		t.Error("an array must be an error")
	}
}

func TestSettingsMarshalStable(t *testing.T) {
	var s Settings
	for _, f := range []Feature{"d", "b", "a", "c"} {
		s.Set(f, All)
	}
	first, _ := json.Marshal(s)
	for i := 0; i < 20; i++ {
		again, _ := json.Marshal(s)
		if string(again) != string(first) {
			t.Fatalf("marshal is not stable: %s vs %s", first, again)
		}
	}
}

func TestSettingsClone(t *testing.T) {
	var s Settings
	s.Set(FeatureMark, Current)
	c := s.Clone()
	c.Set(FeatureMark, All)
	if s.ByFeature[FeatureMark] != Current {
		t.Error("Clone shares the map with the original")
	}
}

func TestSearchFeatureDefaultsToCurrent(t *testing.T) {
	var s Settings
	if got := s.For(FeatureSearch); got != Current {
		t.Fatalf("builtin = %s, want current", got)
	}
	s.Set(FeatureSearch, Workspaces)
	if got := s.For(FeatureSearch); got != Workspaces {
		t.Fatalf("set = %s, want workspaces", got)
	}
}
