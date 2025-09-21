package utils

// 辅助函数：将驼峰命名转换为蛇形命名
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return string(result)
}
