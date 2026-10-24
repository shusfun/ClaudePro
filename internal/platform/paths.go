package platform

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func ValidateLaunchPath(path string, extensions ...string) error {
	if !filepath.IsAbs(path) {
		return errors.New("客户端路径必须是绝对路径")
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return errors.New("请选择客户端文件")
	}
	for _, ext := range extensions {
		if strings.EqualFold(filepath.Ext(path), ext) {
			return nil
		}
	}
	return errors.New("不支持的客户端文件类型")
}
