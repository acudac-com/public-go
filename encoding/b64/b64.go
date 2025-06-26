/*
Package b64 provides utility functions for encoding and decoding data using
Base64 encoding. It supports both standard Base64 and URL-safe Base64 (for use
in URLs and cookies).
*/
package b64

import "encoding/base64"

// Base64 encodes the payload, using URLEncoding to be used in urls and cookies.
func UrlEncode(data []byte) []byte {
	return encode(base64.URLEncoding, data)
}

// Base64 decodes the payload, using URLEncoding to be used in urls and cookies.
func UrlDecode(data []byte) ([]byte, error) {
	return decode(base64.URLEncoding, data)
}

// Base64 encodes the payload.
func StdEncode(data []byte) []byte {
	return encode(base64.StdEncoding, data)
}

// Base64 decodes the payload.
func StdDecode(data []byte) ([]byte, error) {
	return decode(base64.StdEncoding, data)
}

func encode(encoding *base64.Encoding, data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	encodedLen := encoding.EncodedLen(len(data))
	encodedBytes := make([]byte, encodedLen)
	encoding.Encode(encodedBytes, data)
	return encodedBytes
}

func decode(encoding *base64.Encoding, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}
	decodedLen := encoding.DecodedLen(len(data))
	decodedBytes := make([]byte, decodedLen)
	n, err := encoding.Decode(decodedBytes, data)
	if err != nil {
		return nil, err
	}
	if n != decodedLen {
		return decodedBytes[:n], nil
	}
	return decodedBytes, nil
}
