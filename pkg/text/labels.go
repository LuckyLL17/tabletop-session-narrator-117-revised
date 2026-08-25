package text

import "strings"

func ActionStyle(labels []string, score, resources int) string {
	joined := strings.ToLower(strings.Join(labels, " "))
	if strings.Contains(joined, "合作") || strings.Contains(joined, "交易") {
		return "协商型"
	}
	if score > resources+8 {
		return "进攻型"
	}
	if resources > score+8 {
		return "经营型"
	}
	return "平衡型"
}
func Sentence(
	value string,
) string {
	value = Clean(value)
	if value == "" {
		return "暂无记录"
	}
	if strings.HasSuffix(value, "。") {
		return value
	}
	return value + "。"
}
