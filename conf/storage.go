package conf

import (
	"fmt"
	"path/filepath"
)

// Storage 存储相关配置
type Storage struct {
	RootDir string `yaml:"root_dir"` // RootDir 文件系统上的存储根路径
}

// validate 本方法用于检查配置文件中的 storage 配置段中的root_dir配置项是否为绝对路径
func (s *Storage) validate() error {
	isAbs := filepath.IsAbs(s.RootDir)
	if !isAbs {
		return fmt.Errorf("storage.root_dir 必须为绝对路径, 当前值: %q", s.RootDir)
	}

	return nil
}