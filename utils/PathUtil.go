package utils

import (
	"os"
	"path"
)

func GetUserDesktopPath() string {
	// 用户目录
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return path.Join(userHomeDir, "Desktop")
}
