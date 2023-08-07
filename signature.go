package polygon

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"net/url"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

func Signature(method string, secret string, params url.Values) string {
	salt := fmt.Sprint(100000 + rnd.Int()%899999)

	hash := sha512.New()
	hash.Write([]byte(salt + "/" + method + "?" + params.Encode() + "#" + secret))

	return salt + fmt.Sprintf("%x", hash.Sum(nil))
}
