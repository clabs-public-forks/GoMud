package web

import (
	"fmt"
	htemplate "html/template"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
)

func rpadString(totalWidth int, padValues ...any) string {
	var stringIn string = ""
	var padString string = " "

	if len(padValues) > 0 {
		stringIn = fmt.Sprintf(`%v`, padValues[0])
		if len(padValues) > 1 {
			padString = fmt.Sprintf(`%v`, padValues[1])
		}
	}

	if len(stringIn) >= totalWidth {
		return stringIn
	}
	paddingLength := totalWidth - len(stringIn)
	if paddingLength < 1 {
		return stringIn
	}
	return stringIn + strings.Repeat(padString, paddingLength)
}

var (
	funcMap = htemplate.FuncMap{
		"pad": func(totalWidth int, padValues ...any) string {
			var stringIn string = ""
			var padString string = " "

			if len(padValues) > 0 {
				stringIn = fmt.Sprintf(`%v`, padValues[0])
				if len(padValues) > 1 {
					padString = fmt.Sprintf(`%v`, padValues[1])
				}
			}

			if len(stringIn) >= totalWidth {
				return stringIn
			}
			paddingLength := totalWidth - len(stringIn)
			leftPad := paddingLength >> 1
			if leftPad < 1 {
				return stringIn
			}
			if paddingLength-leftPad < 1 {
				return strings.Repeat(padString, leftPad) + stringIn
			}
			return strings.Repeat(padString, leftPad) + stringIn + strings.Repeat(padString, paddingLength-leftPad)
		},
		"lpad": func(totalWidth int, padValues ...any) string {
			var stringIn string = ""
			var padString string = " "

			if len(padValues) > 0 {
				stringIn = fmt.Sprintf(`%v`, padValues[0])
				if len(padValues) > 1 {
					padString = fmt.Sprintf(`%v`, padValues[1])
				}
			}

			if len(stringIn) >= totalWidth {
				return stringIn
			}
			paddingLength := totalWidth - len(stringIn)
			if paddingLength < 1 {
				return stringIn
			}
			return strings.Repeat(padString, paddingLength) + stringIn
		},
		"rpad": func(totalWidth int, padValues ...any) string {
			return rpadString(totalWidth, padValues...)
		},
		"rpadhtml": func(totalWidth int, padValues ...any) htemplate.HTML {
			return htemplate.HTML(rpadString(totalWidth, padValues...))
		},
		"join": func(items []string, sep string) string { return strings.Join(items, sep) },
		"lte":  func(a, b int) bool { return a <= b },
		"gte":  func(a, b int) bool { return a >= b },
		"lt":   func(a, b int) bool { return a < b },
		//"gt":   func(a, b int) bool { return a > b },
		"uc":  func(s string) string { return strings.Title(s) },
		"lc":  func(s string) string { return strings.ToLower(s) },
		"add": func(num int, amt int) int { return num + amt },
		"sub": func(num int, amt int) int { return num - amt },
		"mul": func(num int, amt int) int { return num * amt },
		"intRange": func(start, end int) []int {
			n := end - start + 1
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = start + i
			}
			return result
		},
		"escapehtml": func(str string) string {
			return str
		},
		"lowercase": func(str string) string {
			return strings.ToLower(str)
		},
		"getconfig": func() configs.Config {
			return configs.GetConfig()
		},
	}
)
