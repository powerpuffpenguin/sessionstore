package bolt

import (
	"fmt"
	"io/fs"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// bolt://mode/path?k0=v0&k1=v1
// bolt://0600/bbolt.db?Timeout=1s
// bolt://0600//bbolt.db?Timeout=1s
func ParseURL(rawURL string) (path string, mode fs.FileMode, options *bolt.Options, e error) {
	u, e := url.Parse(rawURL)
	if e != nil {
		return
	}
	if u.Scheme != `bbolt` && u.Scheme != `bolt` {
		e = fmt.Errorf(`unknow scheme %s`, u.Scheme)
		return
	}
	i, e := strconv.ParseUint(u.Host, 8, 32)
	if e != nil {
		return
	}
	path = strings.TrimPrefix(u.Path, `/`)
	mode = fs.FileMode(i)

	query := u.Query()
	options = &bolt.Options{
		NoGrowSync: isTrue(query.Get(`NoGrowSync`)),
		ReadOnly:   isTrue(query.Get(`ReadOnly`)),
	}
	s := query.Get(`Timeout`)
	if s != `` {
		options.Timeout, e = time.ParseDuration(s)
		if e != nil {
			return
		}
	}
	s = query.Get(`MmapFlags`)
	if s != `` && s != `0` {
		options.MmapFlags, e = strconv.Atoi(s)
		if e != nil {
			return
		}
	}
	s = query.Get(`InitialMmapSize`)
	if s != `` && s != `0` {
		options.InitialMmapSize, e = strconv.Atoi(s)
		if e != nil {
			return
		}
	}
	return
}

func isTrue(str string) bool {
	return str != `` && str != `0` && str != `false` && str != `undefined` && str != `null` && str != `nil`
}
