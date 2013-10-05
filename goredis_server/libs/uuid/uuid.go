package uuid

import (
	"crypto/md5"
	"io"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// 生成uuid， n <= 32
func UUID(n int) string {
	// 先用系统的生成，失败后用时间生成
	uid, e1 := uuidFromOS()
	if e1 != nil {
		uid = uuidFromTime()
	}
	return uid[:n]
}

func uuidFromTime() (uid string) {
	nano := time.Now().UnixNano()
	rand.Seed(nano)
	rndNum := rand.Int63()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(nano, 10))
	io.WriteString(h, strconv.FormatInt(rndNum, 10))
	bs := h.Sum(nil)
	uid = string(bs)
	return
}

func uuidFromOS() (uid string, err error) {
	var out []byte
	out, err = exec.Command("uuidgen").Output()
	if err != nil {
		return
	}
	uid = strings.ToLower(string(out))
	uid = strings.Replace(uid, "-", "", -1)
	return
}
