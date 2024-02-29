package live

import (
	"bytes"
	"sort"
	"strings"

	"github.com/vtb-link/bianka/basic"
)

func Md5(str string) (md5str string) {
	return basic.Md5(str)
}

func HmacSHA256(key string, data string) string {
	return basic.HmacSHA256(key, data)
}

func ToSortedString(payloadMap map[string]string) string {
	hSil := make([]string, 0, 8)
	for k := range payloadMap {
		hSil = append(hSil, k)
	}
	sort.Strings(hSil)

	sign := bytes.NewBufferString("")
	for _, v := range hSil {
		sign.WriteString(v + ":" + payloadMap[v] + "\n")
	}

	return strings.TrimRight(sign.String(), "\n")
}
