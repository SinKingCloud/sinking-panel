package sqlite

import (
	"errors"
	"strings"
	"unicode"
)

// schemaTokens 读取结构语句的标识符与边界；跳过注释，保留引号标识符与字符串的区别。
func (m *Manager) schemaTokens(statement string) ([]string, error) {
	if len(statement) > 1<<20 || strings.ContainsRune(statement, 0) {
		return nil, errors.New("结构定义不合法或超过 1 MB")
	}
	tokens := make([]string, 0)
	for index := 0; index < len(statement); {
		c := statement[index]
		if unicode.IsSpace(rune(c)) {
			index++
			continue
		}
		if strings.HasPrefix(statement[index:], "--") {
			if end := strings.IndexByte(statement[index:], '\n'); end >= 0 {
				index += end + 1
			} else {
				break
			}
			continue
		}
		if strings.HasPrefix(statement[index:], "/*") {
			end := strings.Index(statement[index+2:], "*/")
			if end < 0 {
				return nil, errors.New("结构定义中存在未闭合注释")
			}
			index += end + 4
			continue
		}
		if c == '\'' || c == '"' || c == '`' || c == '[' {
			endChar := c
			if c == '[' {
				endChar = ']'
			}
			index++
			var value strings.Builder
			closed := false
			for index < len(statement) {
				next := statement[index]
				index++
				if next == endChar {
					if index < len(statement) && statement[index] == endChar {
						value.WriteByte(endChar)
						index++
						continue
					}
					closed = true
					break
				}
				value.WriteByte(next)
			}
			if !closed {
				return nil, errors.New("结构定义中存在未闭合引号")
			}
			if c == '\'' {
				tokens = append(tokens, "#string")
			} else {
				tokens = append(tokens, "\x01"+value.String())
			}
			continue
		}
		if strings.ContainsRune(".;(),=+-*/", rune(c)) {
			tokens = append(tokens, string(c))
			index++
			continue
		}
		end := index + 1
		for end < len(statement) && !unicode.IsSpace(rune(statement[end])) && !strings.ContainsRune(".;(),=+-*/'\"`[]", rune(statement[end])) {
			end++
		}
		tokens = append(tokens, strings.ToUpper(statement[index:end]))
		index = end
	}
	return tokens, nil
}
