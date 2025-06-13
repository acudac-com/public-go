package b64

import "encoding/base64"

func UrlEncode(data []byte) []byte {
	return encode(base64.URLEncoding, data)
}

func UrlDecode(data []byte) ([]byte, error) {
	return decode(base64.URLEncoding, data)
}

func StdEncode(data []byte) []byte {
	return encode(base64.StdEncoding, data)
}

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
