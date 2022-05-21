package sessionstore_test

import (
	"testing"

	"github.com/powerpuffpenguin/sessionstore/store/bbolt"
)

func TestBBolt(t *testing.T) {
	store, e := bbolt.New(bbolt.WithLimit(3))
	if e != nil {
		t.Fatal(e)
	}
	e = store.Reset()
	if e != nil {
		t.Fatal(e)
	}
	testManager(t, store)
}
func TestBBoltRefresh(t *testing.T) {
	store, e := bbolt.New(bbolt.WithLimit(3))
	if e != nil {
		t.Fatal(e)
	}
	e = store.Reset()
	if e != nil {
		t.Fatal(e)
	}

	testManagerRefresh(t, store)
}
