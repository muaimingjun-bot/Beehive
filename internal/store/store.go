package store

import (
	"os"

	"gopkg.in/yaml.v3"
)

// MarshalYAML 将结构体序列化为 YAML 字节
func MarshalYAML(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

// UnmarshalYAML 从文件读取并反序列化
func UnmarshalYAML(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// WriteFile 写 YAML 文件
func WriteFile(path string, v any) error {
	data, err := MarshalYAML(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
