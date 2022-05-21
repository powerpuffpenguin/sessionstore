package sessionstore_test

import (
	"testing"

	"github.com/powerpuffpenguin/sessionstore/store/bolt"
)

func TestBolt(t *testing.T) {
	store, e := bolt.New(bolt.WithLimit(3))
	if e != nil {
		t.Fatal(e)
	}
	e = store.Reset()
	if e != nil {
		t.Fatal(e)
	}
	testManager(t, store)
}
func TestBoltRefresh(t *testing.T) {
	store, e := bolt.New(bolt.WithLimit(3))
	if e != nil {
		t.Fatal(e)
	}
	e = store.Reset()
	if e != nil {
		t.Fatal(e)
	}

	testManagerRefresh(t, store)
}
