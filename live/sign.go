package live

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func Md5(str string) (md5str string) {
	sum := md5.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func HmacSHA256(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func ToSortedString(payloadMap map[string]string) string {
	var hSil []string
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
