package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type fakeHost struct {
	env         Environment
	failApply   string
	failRestore string
	mismatch    string
	processes   []string
	writes      []string
}

func newFake() *fakeHost {
	return &fakeHost{env: Environment{
		"timezone": {Label: "China Standard Time", Data: map[string]string{"id": "China Standard Time"}},
		"culture":  {Label: "zh-CN", Data: map[string]string{"LocaleName": "zh-CN", "sShortDate": "yyyy~MM~dd", "sDecimal": ","}},
		"region":   {Label: "45", Data: map[string]string{"Nation": "45", "Name": "CN"}},
	}}
}
func clone[T any](v T) T                       { b, _ := json.Marshal(v); var r T; json.Unmarshal(b, &r); return r }
func (f *fakeHost) Read() (Environment, error) { return clone(f.env), nil }
func (f *fakeHost) Validate(t Target) error {
	if t.Timezone == "" || t.Culture == "" || t.Region == "" {
		return errors.New("invalid")
	}
	return nil
}
func (f *fakeHost) Running() ([]string, error) { return f.processes, nil }
func (f *fakeHost) Apply(k, v string) error {
	f.writes = append(f.writes, "apply:"+k)
	if f.failApply == k {
		return errors.New("UAC cancelled / adapter failed")
	}
	if f.mismatch != k {
		f.env[k] = Value{Label: v, Data: map[string]string{"value": v}}
	}
	return nil
}
func (f *fakeHost) Restore(k string, v Value) error {
	f.writes = append(f.writes, "restore:"+k)
	if f.failRestore == k {
		return errors.New("restore failed")
	}
	f.env[k] = clone(v)
	return nil
}

var target = Target{Timezone: "Eastern Standard Time", Culture: "en-US", Region: "244"}

func setup(t *testing.T) (*Engine, *fakeHost, FileStore) {
	t.Helper()
	h := newFake()
	s := FileStore{Dir: t.TempDir()}
	return New(h, s, nil), h, s
}

func TestRoundTripPreservesCustomFormatsAndBaseline(t *testing.T) {
	e, h, s := setup(t)
	before := clone(h.env)
	if err := e.Switch(target); err != nil {
		t.Fatal(err)
	}
	j, _ := s.Load()
	id := j.ID
	if err := e.Switch(Target{Timezone: "other", Culture: "other", Region: "other"}); err == nil {
		t.Fatal("duplicate switch permitted")
	}
	j, _ = s.Load()
	if j.ID != id || !reflect.DeepEqual(j.Original, before) {
		t.Fatal("baseline overwritten")
	}
	if err := e.Restore(nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(h.env, before) {
		t.Fatal("custom formats not restored")
	}
	if got := h.writes[len(h.writes)-3:]; !reflect.DeepEqual(got, []string{"restore:region", "restore:culture", "restore:timezone"}) {
		t.Fatal(got)
	}
	if err := e.Switch(target); err != nil {
		t.Fatal(err)
	}
	files, _ := os.ReadDir(filepath.Join(s.Dir, "history"))
	if len(files) != 1 {
		t.Fatal("missing history")
	}
}
func TestApplyFailureAndUACCancelRollback(t *testing.T) {
	for _, field := range Fields {
		t.Run(field, func(t *testing.T) {
			e, h, s := setup(t)
			before := clone(h.env)
			h.failApply = field
			if e.Switch(target) == nil {
				t.Fatal("failure not reported")
			}
			if !reflect.DeepEqual(before, h.env) {
				t.Fatal("rollback incomplete")
			}
			j, _ := s.Load()
			if j.Status != "restored" {
				t.Fatal(j.Status)
			}
		})
	}
}
func TestReadbackMismatchRollsBack(t *testing.T) {
	e, h, _ := setup(t)
	before := clone(h.env)
	h.mismatch = "culture"
	if e.Switch(target) == nil {
		t.Fatal("mismatch ignored")
	}
	if !reflect.DeepEqual(before, h.env) {
		t.Fatal("rollback incomplete")
	}
}
func TestRestoreFailureCanResume(t *testing.T) {
	e, h, s := setup(t)
	before := clone(h.env)
	if err := e.Switch(target); err != nil {
		t.Fatal(err)
	}
	h.failRestore = "culture"
	if e.Restore(nil) == nil {
		t.Fatal("expected failure")
	}
	j, _ := s.Load()
	if j.Status != "needs_restore" || !j.Steps[2].Done {
		t.Fatal(j)
	}
	h.failRestore = ""
	e = New(h, s, nil)
	if err := e.Restore(nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, h.env) {
		t.Fatal("resume failed")
	}
}
func TestConflictRequiresObservedValueAndRecordsKeep(t *testing.T) {
	e, h, s := setup(t)
	if err := e.Switch(target); err != nil {
		t.Fatal(err)
	}
	h.env["culture"] = Value{Label: "fr-FR", Data: map[string]string{"custom": "external"}}
	if e.Restore(nil) == nil {
		t.Fatal("overwrote conflict")
	}
	v, err := e.Inspect()
	if err != nil || len(v.Conflicts) != 1 {
		t.Fatal(v, err)
	}
	c := v.Conflicts[0]
	choice := map[string]Choice{c.Field: {Action: "keep", Observed: c.Current}}
	h.env["culture"].Data["custom"] = "changed again"
	if e.Restore(choice) == nil {
		t.Fatal("stale decision accepted")
	}
	choice[c.Field] = Choice{Action: "keep", Observed: clone(h.env["culture"])}
	if err = e.Restore(choice); err != nil {
		t.Fatal(err)
	}
	if h.env["culture"].Label != "fr-FR" {
		t.Fatal("external value lost")
	}
	j, _ := s.Load()
	if j.Steps[1].Resolution != "kept_external" {
		t.Fatal("decision not logged")
	}
}
func TestExplicitConflictRestore(t *testing.T) {
	e, h, _ := setup(t)
	original := clone(h.env)
	e.Switch(target)
	h.env["region"] = Value{Label: "other", Data: map[string]string{"a": "b"}}
	choices := map[string]Choice{"region": {Action: "restore", Observed: clone(h.env["region"])}}
	if err := e.Restore(choices); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(h.env, original) {
		t.Fatal("restore failed")
	}
}
func TestRestartAfterUnverifiedWrite(t *testing.T) {
	e, h, s := setup(t)
	original := clone(h.env)
	j := &Journal{Version: 1, ID: "crash", Status: "switching", Original: original, Target: target, Steps: []Step{{Field: "timezone", Before: original["timezone"], Pending: true}}}
	if err := s.Save(j); err != nil {
		t.Fatal(err)
	}
	h.Apply("timezone", target.Timezone)
	v, err := e.Inspect()
	if err != nil {
		t.Fatal(err)
	}
	if v.Journal.Status != "needs_restore" || len(v.Conflicts) != 1 {
		t.Fatal(v)
	}
	if e.Switch(target) == nil {
		t.Fatal("pending baseline replaced")
	}
	if err = e.Restore(map[string]Choice{"timezone": {Action: "restore", Observed: v.Current["timezone"]}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(h.env, original) {
		t.Fatal("crash restore failed")
	}
}
func TestInterruptedRestoreAfterWriteBeforeJournal(t *testing.T) {
	e, h, s := setup(t)
	e.Switch(target)
	j, _ := s.Load()
	j.Status = "restoring"
	s.Save(j)
	h.Restore("region", j.Original["region"])
	e = New(h, s, nil)
	if _, err := e.Inspect(); err != nil {
		t.Fatal(err)
	}
	if err := e.Restore(nil); err != nil {
		t.Fatal(err)
	}
}
func TestRunningClaudeBlocksAllWrites(t *testing.T) {
	e, h, s := setup(t)
	h.processes = []string{"claude.exe (PID 123)"}
	if e.Switch(target) == nil {
		t.Fatal("running client allowed")
	}
	if len(h.writes) > 0 {
		t.Fatal(h.writes)
	}
	j, _ := s.Load()
	if j != nil {
		t.Fatal("created baseline despite running client")
	}
}
func TestCorruptOrUnsupportedJournalBlocksWrites(t *testing.T) {
	for _, raw := range []string{"{bad", `{"version":2}`, `null`} {
		t.Run(raw, func(t *testing.T) {
			e, h, s := setup(t)
			os.WriteFile(filepath.Join(s.Dir, "journal.json"), []byte(raw), 0600)
			if e.Switch(target) == nil {
				t.Fatal("corruption accepted")
			}
			if len(h.writes) > 0 {
				t.Fatal("wrote settings")
			}
			b, _ := os.ReadFile(filepath.Join(s.Dir, "journal.json"))
			if string(b) != raw {
				t.Fatal("corrupt evidence overwritten")
			}
		})
	}
}

type brokenStore struct {
	Store
	saves  int
	failAt int
}

func (s *brokenStore) Save(j *Journal) error {
	s.saves++
	if s.saves == s.failAt {
		return errors.New("disk full")
	}
	return s.Store.Save(j)
}
func TestWriteAheadPersistenceFailure(t *testing.T) {
	for _, at := range []int{1, 2, 3} {
		t.Run(string(rune('0'+at)), func(t *testing.T) {
			h := newFake()
			s := FileStore{Dir: t.TempDir()}
			b := &brokenStore{Store: s, failAt: at}
			e := New(h, b, nil)
			if e.Switch(target) == nil {
				t.Fatal("disk failure ignored")
			}
			if at < 3 && len(h.writes) > 0 {
				t.Fatal("write without durable intent")
			}
			if at == 3 {
				j, err := s.Load()
				if err != nil || !j.Steps[0].Pending {
					t.Fatal("lost pending write", err)
				}
			}
		})
	}
}
func TestMutualExclusion(t *testing.T) {
	e, _, _ := setup(t)
	if !e.Lock() {
		t.Fatal("lock")
	}
	defer e.Unlock()
	if err := e.Switch(target); err == nil || !strings.Contains(err.Error(), "正在进行") {
		t.Fatal(err)
	}
}

func TestStructurallyDamagedBackupBlocksRestoration(t *testing.T) {
	for _, damage := range []string{"empty-original", "missing-step", "missing-readback", "pending-active"} {
		t.Run(damage, func(t *testing.T) {
			e, h, s := setup(t)
			if err := e.Switch(target); err != nil {
				t.Fatal(err)
			}
			j, _ := s.Load()
			switch damage {
			case "empty-original":
				v := j.Original["culture"]
				v.Data = map[string]string{}
				j.Original["culture"] = v
			case "missing-step":
				j.Steps = j.Steps[:2]
			case "missing-readback":
				j.Steps[0].After = nil
			case "pending-active":
				j.Steps[0].Pending = true
			}
			if err := s.Save(j); err != nil {
				t.Fatal(err)
			}
			writes := len(h.writes)
			if e.Restore(nil) == nil || len(h.writes) != writes {
				t.Fatal("damaged backup caused a write")
			}
		})
	}
}

func TestAbsentMacPreferenceRoundTrip(t *testing.T) {
	for _, field := range []string{"culture", "region"} {
		t.Run(field, func(t *testing.T) {
			e, h, s := setup(t)
			absent := Value{Label: "", Data: map[string]string{"present": "0", "value": ""}}
			h.env[field] = absent
			if err := e.Switch(target); err != nil {
				t.Fatal(err)
			}
			// 重启后也必须能读取备份并恢复不存在的键。
			e = New(h, s, nil)
			if _, err := e.Inspect(); err != nil {
				t.Fatal(err)
			}
			if err := e.Restore(nil); err != nil {
				t.Fatal(err)
			}
			if !Equal(h.env[field], absent) {
				t.Fatal(h.env[field])
			}
			if _, err := s.Load(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPendingTimezoneWithAbsentCountry(t *testing.T) {
	e, h, s := setup(t)
	h.env["region"] = Value{Data: map[string]string{"present": "0", "value": ""}}
	original := clone(h.env)
	h.env["timezone"] = Value{Label: target.Timezone, Data: map[string]string{"id": target.Timezone}}
	j := &Journal{Version: 1, ID: "mac-readback-failure", Status: "needs_restore", Original: original, Target: target, Steps: []Step{{Field: "timezone", Before: original["timezone"], Pending: true}}}
	if err := s.Save(j); err != nil {
		t.Fatal(err)
	}
	view, err := e.Inspect()
	if err != nil || len(view.Conflicts) != 1 {
		t.Fatalf("%+v %v", view, err)
	}
	if e.Restore(nil) == nil || len(h.writes) != 0 {
		t.Fatal("unknown write restored without choice")
	}
	choice := map[string]Choice{"timezone": {Action: "restore", Observed: h.env["timezone"]}}
	if err := e.Restore(choice); err != nil {
		t.Fatal(err)
	}
	if !Equal(h.env["timezone"], original["timezone"]) {
		t.Fatal(h.env)
	}
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyStoredValueRequiresExplicitAbsentPreference(t *testing.T) {
	for _, data := range []map[string]string{nil, {"present": "0"}, {"present": "1", "value": ""}, {"present": "0", "value": "US"}, {"Nation": ""}} {
		if validStoredValue("region", Value{Data: data}) {
			t.Fatalf("accepted incomplete data: %v", data)
		}
	}
	if validStoredValue("timezone", Value{Data: map[string]string{"present": "0", "value": ""}}) {
		t.Fatal("accepted absent timezone")
	}
}
