package utils

import (
	"os"
)

// FileExists 文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
