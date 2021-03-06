package matrix

import (
	"strings"
)

func ParseMatrixFilter(arr []string) *Filter {
	ret := NewFilter(make(map[string][]string))

	for _, v := range arr {
		if idx := strings.Index(v, "!="); idx > 0 {
			ret.AddIgnore(v[:idx], v[idx+2:])
			continue
		}

		if idx := strings.IndexByte(v, '='); idx > 0 {
			ret.AddMatch(v[:idx], v[idx+1:])
			continue
		}
	}

	return ret
}
