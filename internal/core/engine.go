package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var Fields = []string{"timezone", "culture", "region"}

type Value struct {
	Label string            `json:"label"`
	Data  map[string]string `json:"data"`
}
type Environment map[string]Value
type Target struct {
	Timezone string `json:"timezone"`
	Culture  string `json:"culture"`
	Region   string `json:"region"`
}

func (t Target) Value(field string) string {
	switch field {
	case "timezone":
		return t.Timezone
	case "culture":
		return t.Culture
	default:
		return t.Region
	}
}

type Config struct {
	Version     int    `json:"version"`
	Target      Target `json:"target"`
	CLIPath     string `json:"cliPath"`
	DesktopPath string `json:"desktopPath"`
	WorkDir     string `json:"workDir"`
}
type Option struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Catalog struct {
	Timezones []Option `json:"timezones"`
	Cultures  []Option `json:"cultures"`
	Regions   []Option `json:"regions"`
}
type Clients struct {
	CLI         string `json:"cli"`
	Desktop     string `json:"desktop"`
	DesktopKind string `json:"desktopKind"`
}
type Adapter interface {
	Read() (Environment, error)
	Validate(Target) error
	Apply(string, string) error
	Restore(string, Value) error
	Running() ([]string, error)
}
type Step struct {
	Field      string `json:"field"`
	Before     Value  `json:"before"`
	After      *Value `json:"after,omitempty"`
	Pending    bool   `json:"pending"`
	Done       bool   `json:"done"`
	Resolution string `json:"resolution,omitempty"`
}
type Entry struct {
	Time    string `json:"time"`
	Message string `json:"message"`
}
type Journal struct {
	Version  int         `json:"version"`
	ID       string      `json:"id"`
	Status   string      `json:"status"`
	Original Environment `json:"original"`
	Target   Target      `json:"target"`
	Steps    []Step      `json:"steps"`
	Log      []Entry     `json:"log"`
}
type Store interface {
	Load() (*Journal, error)
	Save(*Journal) error
	Archive(*Journal) error
}
type Conflict struct {
	Field    string `json:"field"`
	Current  Value  `json:"current"`
	Original Value  `json:"original"`
	Expected *Value `json:"expected,omitempty"`
}
type Choice struct {
	Action   string `json:"action"`
	Observed Value  `json:"observed"`
}
type View struct {
	Current   Environment `json:"current"`
	Journal   *Journal    `json:"journal"`
	Conflicts []Conflict  `json:"conflicts"`
}
type Engine struct {
	mu      sync.Mutex
	adapter Adapter
	store   Store
	notify  func(string)
}

func New(a Adapter, s Store, notify func(string)) *Engine {
	return &Engine{adapter: a, store: s, notify: notify}
}
func Equal(a, b Value) bool  { return a.Label == b.Label && reflect.DeepEqual(a.Data, b.Data) }
func Active(j *Journal) bool { return j != nil && j.Status != "restored" }
func (e *Engine) Lock() bool { return e.mu.TryLock() }
func (e *Engine) Unlock()    { e.mu.Unlock() }
func (e *Engine) HasActive() (bool, error) {
	if !e.mu.TryLock() {
		return true, nil
	}
	defer e.mu.Unlock()
	j, err := e.store.Load()
	return Active(j), err
}
func (e *Engine) record(j *Journal, msg string) error {
	j.Log = append(j.Log, Entry{Time: time.Now().UTC().Format(time.RFC3339Nano), Message: msg})
	if err := e.store.Save(j); err != nil {
		return fmt.Errorf("备份写入失败，操作已停止：%w", err)
	}
	if e.notify != nil {
		e.notify(msg)
	}
	return nil
}
func conflicts(j *Journal, env Environment) []Conflict {
	result := []Conflict{}
	if !Active(j) {
		return result
	}
	for _, s := range j.Steps {
		if s.Done || Equal(env[s.Field], s.Before) {
			continue
		}
		if s.After == nil || !Equal(env[s.Field], *s.After) {
			result = append(result, Conflict{Field: s.Field, Current: env[s.Field], Original: s.Before, Expected: s.After})
		}
	}
	return result
}
func (e *Engine) Inspect() (View, error) {
	if !e.mu.TryLock() {
		return View{}, errors.New("操作正在进行，请稍候")
	}
	defer e.mu.Unlock()
	j, err := e.store.Load()
	if err != nil {
		return View{}, err
	}
	env, err := e.adapter.Read()
	if err != nil {
		return View{}, err
	}
	if Active(j) && j.Status != "needs_restore" {
		drift := false
		for _, step := range j.Steps {
			if !step.Done && (step.After == nil || !Equal(env[step.Field], *step.After)) {
				drift = true
			}
		}
		if j.Status != "active" || drift {
			j.Status = "needs_restore"
			if err = e.record(j, "发现未完成操作或外部修改，请恢复原环境"); err != nil {
				return View{}, err
			}
		}
	}
	return View{Current: env, Journal: j, Conflicts: conflicts(j, env)}, nil
}
func (e *Engine) ensureStopped() error {
	p, err := e.adapter.Running()
	if err != nil {
		return err
	}
	if len(p) > 0 {
		return fmt.Errorf("请保存工作并关闭 Claude 后重试：%v", p)
	}
	return nil
}
func (e *Engine) Switch(target Target) error {
	if !e.mu.TryLock() {
		return errors.New("操作正在进行")
	}
	defer e.mu.Unlock()
	old, err := e.store.Load()
	if err != nil {
		return err
	}
	if Active(old) {
		return errors.New("原环境尚未恢复，不能重复切换或覆盖备份")
	}
	if err = e.adapter.Validate(target); err != nil {
		return err
	}
	if err = e.ensureStopped(); err != nil {
		return err
	}
	original, err := e.adapter.Read()
	if err != nil {
		return err
	}
	if old != nil {
		if err = e.store.Archive(old); err != nil {
			return err
		}
	}
	j := &Journal{Version: 1, ID: time.Now().UTC().Format("20060102T150405.000000000"), Status: "switching", Original: original, Target: target, Steps: []Step{}, Log: []Entry{}}
	if err = e.record(j, "原始配置已保存，开始切换"); err != nil {
		return err
	}
	for _, field := range Fields {
		before, readErr := e.adapter.Read()
		if readErr != nil {
			return e.fail(j, readErr)
		}
		if !Equal(before[field], original[field]) {
			return e.fail(j, fmt.Errorf("%s 在切换期间被外部修改", field))
		}
		if original[field].Label == target.Value(field) {
			if err = e.record(j, field+" 已匹配目标，保留现有设置"); err != nil {
				return err
			}
			continue
		}
		j.Steps = append(j.Steps, Step{Field: field, Before: original[field], Pending: true})
		if err = e.record(j, "准备修改 "+field); err != nil {
			return err
		}
		applyErr := e.adapter.Apply(field, target.Value(field))
		actual, readErr := e.adapter.Read()
		if readErr == nil {
			v := actual[field]
			j.Steps[len(j.Steps)-1].After = &v
		}
		if err = e.record(j, "已执行 "+field+"，核对结果"); err != nil {
			return err
		}
		if applyErr != nil {
			return e.fail(j, applyErr)
		}
		if readErr != nil {
			return e.fail(j, readErr)
		}
		if actual[field].Label != target.Value(field) {
			return e.fail(j, fmt.Errorf("%s 回读值与目标不一致", field))
		}
		j.Steps[len(j.Steps)-1].Pending = false
		if err = e.record(j, field+" 已验证"); err != nil {
			return err
		}
	}
	j.Status = "active"
	return e.record(j, "切换完成，可以启动 Claude")
}
func (e *Engine) fail(j *Journal, cause error) error {
	j.Status = "needs_restore"
	if err := e.record(j, "切换失败："+cause.Error()); err != nil {
		return errors.Join(cause, err)
	}
	// Rollback uses the same conflict checks as a manual restore. Unknown writes
	// (e.g. a failed read) require explicit review rather than overwriting changes.
	err := e.restore(j, nil)
	if err != nil {
		return fmt.Errorf("%v；自动恢复未完成：%w", cause, err)
	}
	return fmt.Errorf("%v；已恢复本次修改", cause)
}
func (e *Engine) Restore(choices map[string]Choice) error {
	if !e.mu.TryLock() {
		return errors.New("操作正在进行")
	}
	defer e.mu.Unlock()
	j, err := e.store.Load()
	if err != nil {
		return err
	}
	if !Active(j) {
		return errors.New("没有需要恢复的配置")
	}
	if err = e.ensureStopped(); err != nil {
		return err
	}
	return e.restore(j, choices)
}
func (e *Engine) restore(j *Journal, choices map[string]Choice) error {
	env, err := e.adapter.Read()
	if err != nil {
		return err
	}
	for _, c := range conflicts(j, env) {
		choice, ok := choices[c.Field]
		if !ok || (choice.Action != "keep" && choice.Action != "restore") || !Equal(choice.Observed, c.Current) {
			j.Status = "needs_restore"
			saveErr := e.record(j, "存在外部修改或未确认写入，请选择冲突处理方式")
			return errors.Join(errors.New("请检查并处理配置冲突"), saveErr)
		}
	}
	j.Status = "restoring"
	if err = e.record(j, "开始逆序恢复原环境"); err != nil {
		return err
	}
	for i := len(j.Steps) - 1; i >= 0; i-- {
		s := &j.Steps[i]
		if s.Done {
			continue
		}
		env, err = e.adapter.Read()
		if err != nil {
			return e.restoreFailed(j, err)
		}
		if Equal(env[s.Field], s.Before) {
			s.Done = true
			s.Resolution = "unchanged"
			if err = e.record(j, s.Field+" 已是原值"); err != nil {
				return err
			}
			continue
		}
		unexpected := s.After == nil || !Equal(env[s.Field], *s.After)
		if unexpected {
			c, ok := choices[s.Field]
			if !ok || !Equal(c.Observed, env[s.Field]) || (c.Action != "keep" && c.Action != "restore") {
				return e.restoreFailed(j, errors.New("配置再次变化，请刷新后重新选择"))
			}
			if c.Action == "keep" {
				s.Done = true
				s.Resolution = "kept_external"
				if err = e.record(j, s.Field+"：用户选择保留当前值，放弃该项还原"); err != nil {
					return err
				}
				continue
			}
		}
		if err = e.record(j, "准备恢复 "+s.Field); err != nil {
			return err
		}
		if err = e.adapter.Restore(s.Field, s.Before); err != nil {
			return e.restoreFailed(j, err)
		}
		env, err = e.adapter.Read()
		if err != nil {
			return e.restoreFailed(j, err)
		}
		if !Equal(env[s.Field], s.Before) {
			return e.restoreFailed(j, fmt.Errorf("%s 恢复后回读不一致", s.Field))
		}
		s.Done = true
		s.Resolution = "restored"
		if err = e.record(j, s.Field+" 原值已恢复并验证"); err != nil {
			return err
		}
	}
	j.Status = "restored"
	return e.record(j, "恢复流程完成，记录已保留")
}
func (e *Engine) restoreFailed(j *Journal, cause error) error {
	j.Status = "needs_restore"
	return errors.Join(cause, e.record(j, "恢复未完成："+cause.Error()))
}
