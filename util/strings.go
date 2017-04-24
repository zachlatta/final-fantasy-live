package util

import (
	pad "github.com/willf/pad/utf8"
)

func LeftPad(str, padChar string, length int) string {
	return pad.Left(str, length, padChar)
}

func RightPad(str, padChar string, length int) string {
	return pad.Right(str, length, padChar)
}
