package common

import "net/http"

func DetectContentType(content *[]byte) (contentType string) {
	c := *content
	var buffer []byte
	if len(c) > 512 {
		buffer = c[0:512]
	} else {
		buffer = c
	}
	contentType = http.DetectContentType(buffer)
	return
}
