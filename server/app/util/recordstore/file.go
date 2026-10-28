package recordstore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func validateDays(days int) error {
	if days < 1 || days > 365 {
		return errors.New("历史保留天数必须为 1 至 365 天")
	}
	return nil
}

// writeJSON 原子保存配置，写入失败时保留已有文件。
func writeJSON(dir, name string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".recordstore-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temporary, filepath.Join(dir, name))
}
