package util

import (
	"strings"
)

// Given a URL to a Facebook stream, split it into a stream URL and a stream key
// for use in OBS.
//
// rtmp://rtmp-api.facebook.com:80/rtmp/1234567890?ds=1&s_l=1&a=abh2mv1sdjf3sf
//
// Becomes:
//
//  Stream URL: rtmp://rtmp-api.facebook.com:80/rtmp/
//  Stream key: 1234567890?ds=1&s_l=1&a=abh2mv1sdjf3sf
func SplitStreamUrl(url string) (streamUrl, streamkey string) {
	parts := strings.SplitAfter(url, "/rtmp/")

	return parts[0], parts[1]
}
