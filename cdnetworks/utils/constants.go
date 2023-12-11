package utils

import (
	"errors"
	"fmt"
	"strings"
)

var (
	Separator        = ";"
	All              = "all"
	ValidHttpMethods = [...]string{"GET", "POST", "PUT", "HEAD", "DELETE", "OPTIONS"}
	ValidFileTypes   = [...]string{"gif", "png", "bmp", "jpeg", "jpg", "html", "htm", "shtml", "mp3", "wma", "flv", "mp4", "wmv", "zip", "exe", "rar", "css", "txt", "ico", "js", "swf", "m3u8", "xml", "f4m", "bootstarp", "ts"}
)

func CheckFileTypes(types string) error {
	if strings.TrimSpace(types) == "" {
		return errors.New("file type is empty")
	}
	values := strings.Split(types, Separator)
	for _, v := range values {
		if v == All {
			return fmt.Errorf("%s and specific file types cannot be configured at the same time.", All)
		}
		valid := false
		for _, t := range ValidFileTypes {
			if v == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid file type:%s", v)
		}
	}
	return nil
}
