package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct{ Dir string }

func ReadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !json.Valid(b) {
		return fmt.Errorf("文件损坏，保留原文件并停止修改：%s", path)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(v); err != nil {
		return fmt.Errorf("无法读取 %s：%w", path, err)
	}
	return nil
}
func (s FileStore) Load() (*Journal, error) {
	var j Journal
	err := ReadJSON(filepath.Join(s.Dir, "journal.json"), &j)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	validStatus := map[string]bool{"switching": true, "active": true, "restoring": true, "needs_restore": true, "restored": true}
	if j.Version != 1 || j.ID == "" || !validStatus[j.Status] || len(j.Original) != 3 {
		return nil, errors.New("备份格式或版本无效，请保留文件并检查")
	}
	for _, f := range Fields {
		v, ok := j.Original[f]
		if !ok || !validStoredValue(f, v) || j.Target.Value(f) == "" {
			return nil, errors.New("原环境备份不完整")
		}
	}
	seen := map[string]bool{}
	for _, step := range j.Steps {
		if seen[step.Field] || (step.Field != "timezone" && step.Field != "culture" && step.Field != "region") || !Equal(step.Before, j.Original[step.Field]) {
			return nil, errors.New("操作日志步骤无效")
		}
		seen[step.Field] = true
		if step.After != nil && !validStoredValue(step.Field, *step.After) {
			return nil, errors.New("操作日志回读数据不完整")
		}
		if j.Status == "active" && (step.After == nil || step.Pending || step.Done || step.After.Label != j.Target.Value(step.Field)) {
			return nil, errors.New("已切换记录缺少成功验证步骤")
		}
		if j.Status == "restored" && !step.Done {
			return nil, errors.New("恢复记录包含未完成步骤")
		}
	}
	if j.Status == "active" {
		for _, f := range Fields {
			if j.Original[f].Label != j.Target.Value(f) && !seen[f] {
				return nil, errors.New("已切换记录缺少修改步骤")
			}
		}
	}
	return &j, nil
}

// macOS 可完整记录一个不存在的偏好键；这与丢失备份数据不同。
func validStoredValue(field string, v Value) bool {
	if len(v.Data) == 0 {
		return false
	}
	if v.Label != "" {
		return true
	}
	value, hasValue := v.Data["value"]
	return (field == "culture" || field == "region") && v.Data["present"] == "0" && hasValue && value == "" && len(v.Data) == 2
}

func (s FileStore) Save(j *Journal) error { return AtomicJSON(filepath.Join(s.Dir, "journal.json"), j) }
func (s FileStore) Archive(j *Journal) error {
	// Never derive a path from user-editable journal identifiers.
	return AtomicJSON(filepath.Join(s.Dir, "history", fmt.Sprintf("%x.json", []byte(j.ID))), j)
}
func AtomicJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".pending-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return replaceFile(tmp, path)
}
