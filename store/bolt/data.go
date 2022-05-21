package bolt

import (
	"encoding/binary"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/powerpuffpenguin/sessionstore"
	protoc_session "github.com/powerpuffpenguin/sessionstore/sessionstore/session"

	"google.golang.org/protobuf/proto"
)

var (
	bucketData   = []byte(`data`)
	bucketSort   = []byte(`sort`)
	bucketSystem = []byte(`system`)
	keyCount     = []byte(`count`)
)

func createBuckets(t *bolt.Tx) (bsystem, bdata *bolt.Bucket, bsort *bolt.Bucket, e error) {
	bsystem, e = t.CreateBucketIfNotExists(bucketSystem)
	if e != nil {
		return
	}
	bdata, e = t.CreateBucketIfNotExists(bucketData)
	if e != nil {
		return
	}
	bsort, e = t.CreateBucketIfNotExists(bucketSort)
	return
}
func getBuckets(t *bolt.Tx) (bsystem, bdata, bsort *bolt.Bucket) {
	bsystem = t.Bucket(bucketSystem)
	bdata = t.Bucket(bucketData)
	bsort = t.Bucket(bucketSort)
	return
}
func getData(bucket *bolt.Bucket, key []byte) (result *protoc_session.BBoltData, e error) {
	b := bucket.Get(key)
	if b == nil {
		return
	}
	var m protoc_session.BBoltData
	e = proto.Unmarshal(b, &m)
	if e != nil {
		return
	}
	result = &m
	return
}
func getCount(bsystem *bolt.Bucket) (count int64) {
	b := bsystem.Get(keyCount)
	if len(b) == 8 {
		count = int64(binary.LittleEndian.Uint64(b))
	}
	return
}
func setCount(bsystem *bolt.Bucket, count int64) (e error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(count))
	e = bsystem.Put(keyCount, b)
	return
}
func addCount(bsystem *bolt.Bucket, count int64) (e error) {
	e = setCount(bsystem, getCount(bsystem)+count)
	return
}
func delKey(bsystem, bdata, bsort *bolt.Bucket, key []byte) (e error) {
	data, e := getData(bdata, key)
	if e != nil || data == nil {
		return
	}
	e = bdata.Delete(key)
	if e != nil {
		return
	}
	e = bsort.Delete(data.Id)
	if e != nil {
		return
	}
	e = addCount(bsystem, -1)
	return
}
func putSort(bucket *bolt.Bucket, key []byte, deadline int64) (id []byte, e error) {
	u, e := bucket.NextSequence()
	if e != nil {
		return
	}
	id = make([]byte, 8)
	binary.BigEndian.PutUint64(id, u)

	m := &protoc_session.BBoltSort{
		Id:       id,
		Key:      key,
		Deadline: deadline,
	}
	val, e := proto.Marshal(m)
	if e != nil {
		return
	}

	e = bucket.Put(id, val)
	return
}
func putData(bucket *bolt.Bucket, id, key, value []byte, deadline int64) (e error) {
	m := &protoc_session.BBoltData{
		Id:       id,
		Data:     value,
		Deadline: deadline,
	}
	val, e := proto.Marshal(m)
	if e != nil {
		return
	}
	e = bucket.Put(key, val)
	return
}
func popKey(bsystem, bdata, bsort *bolt.Bucket) (e error) {
	var (
		c   = bsort.Cursor()
		now = time.Now().Unix()
		add int64
	)
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}
		var m protoc_session.BBoltSort
		e = proto.Unmarshal(v, &m)
		if e != nil {
			return
		}
		if now < m.Deadline {
			break
		}
		add--
		e = c.Delete()
		if e != nil {
			return
		}

		e = bdata.Delete(m.Key)
		if e != nil {
			return
		}
	}
	if add != 0 {
		e = addCount(bsystem, add)
	}
	return
}
func delKeyPrefix(bsystem, bdata, bsort *bolt.Bucket, prefix string) (e error) {
	var (
		c    = bdata.Cursor()
		bkey = sessionstore.StringToBytes(prefix)
		add  int64
	)
	for k, v := c.Seek(bkey); k != nil && strings.HasPrefix(sessionstore.BytesToString(k), prefix); k, v = c.Next() {
		add--
		e = c.Delete()
		if e != nil {
			return
		}
		if v == nil {
			continue
		}
		var m protoc_session.BBoltData
		e = proto.Unmarshal(v, &m)
		if e != nil {
			return
		}
		e = bsort.Delete(m.Id)
		if e != nil {
			return
		}
	}
	if add != 0 {
		e = addCount(bsystem, add)
	}
	return
}
