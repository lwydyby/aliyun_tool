package server

import (
	"fmt"
	"path/filepath"
	"regexp"
)

func generateNewFileName(oldFileName, newPrefix, exampleFile string) (string, error) {
	// 提取文件扩展名
	oldFileName = mergeStrings(exampleFile, oldFileName)
	ext := filepath.Ext(oldFileName)
	nameWithoutExt := oldFileName[:len(oldFileName)-len(ext)]

	// 根据示例文件名判断集号的提取模式
	var re *regexp.Regexp
	re = regexp.MustCompile(`\$(\d+)\$`)

	// 提取集号
	matches := re.FindStringSubmatch(nameWithoutExt)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法从文件名中提取集号")
	}

	episodeNumber := matches[1]
	// 构建新的文件名
	newFileName := fmt.Sprintf("%sE%s%s", newPrefix, episodeNumber, ext)
	return newFileName, nil
}

func mergeStrings(a, b string) string {
	runeA := []rune(a)
	runeB := []rune(b)

	// 创建一个新的rune slice用于存储结果
	result := make([]rune, 0, len(runeA))
	bIndex := 0
	aIndex := 0
	// 按字符比较和替换
	for ; aIndex < len(runeA) && bIndex < len(runeB); aIndex++ {
		if runeA[aIndex] == '$' {
			result = append(result, runeA[aIndex])
		} else {
			result = append(result, runeB[bIndex])
			bIndex++
		}
	}
	for ; aIndex < len(runeA); aIndex++ {
		result = append(result, runeA[aIndex])
	}

	return string(result)
}
