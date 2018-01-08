package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"fmt"
)

var encoders = map[string]Encoder{}

func _url(str string) string {
	tmp := ""
	for _, val := range str {
		tmp += fmt.Sprintf("%%%x", val)
	}
	return tmp
}

func urlurl(str string) string {
	return _url(_url(str))
}

func _html(str string) string {
	tmp := ""
	for _, val := range str {
		tmp += fmt.Sprintf("&#%d;", val)
	}
	return tmp
}

func htmlhex(str string) string {
	tmp := ""
	for _, val := range str {
		tmp += fmt.Sprintf("&#x%x;", val)
	}
	return tmp
}

func unicode(str string) string {
	tmp := ""
	for _, val := range str {
		tmp += fmt.Sprintf("\\u00%x", val)
	}
	return tmp
}

func hex(str string) string {
	tmp := ""
	for _, val := range str {
		tmp += fmt.Sprintf("\\x%x", val)
	}
	return tmp
}

func md5hash(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func sha1hash(str string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(str)))
}

func sha2hash(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}

func b64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func b32(str string) string {
	return base32.StdEncoding.EncodeToString([]byte(str))
}

func plain(str string) string {
	return str
}

func initEncoders() {
	encoders["url"] = Encoder(_url)
	encoders["urlurl"] = Encoder(urlurl)
	encoders["html"] = Encoder(_html)
	encoders["htmlhex"] = Encoder(htmlhex)
	encoders["unicode"] = Encoder(unicode)
	encoders["hex"] = Encoder(hex)
	encoders["md5hash"] = Encoder(md5hash)
	encoders["sha1hash"] = Encoder(sha1hash)
	encoders["sha2hash"] = Encoder(sha2hash)
	encoders["b64"] = Encoder(b64)
	encoders["b32"] = Encoder(b32)
	encoders["plain"] = Encoder(plain)
}
