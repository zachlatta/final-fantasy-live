package util

import (
	"crypto/md5"
	"fmt"
	"io"

	pad "github.com/willf/pad/utf8"
)

func LeftPad(str, padChar string, length int) string {
	return pad.Left(str, length, padChar)
}

func RightPad(str, padChar string, length int) string {
	return pad.Right(str, length, padChar)
}

func MD5HashString(romPath string) string {
	h := md5.New()
	io.WriteString(h, romPath)

	return fmt.Sprintf("%x", h.Sum(nil))
}
