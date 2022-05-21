package sessionstore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/powerpuffpenguin/sessionstore"
	"github.com/powerpuffpenguin/sessionstore/cryptoer"
	"github.com/powerpuffpenguin/sessionstore/store"
	"github.com/powerpuffpenguin/sessionstore/store/bbolt"
)

type Session struct {
	ID   string
	Name string
}

type Coder struct {
}

func (Coder) Unmarshal(b []byte) (session interface{}, e error) {
	var result Session
	e = json.Unmarshal(b, &result)
	if e != nil {
		return
	}
	session = &result
	return
}
func (Coder) Marshal(session interface{}) (b []byte, e error) {
	b, e = json.Marshal(session)
	return
}
func testManager(t *testing.T, store sessionstore.Store) {
	m := sessionstore.New(Coder{},
		sessionstore.WithStore(store),
		sessionstore.WithAccess(time.Second*1),
		sessionstore.WithRefresh(time.Second*3),
	)

	ctx := context.Background()
	platform := `test`
	tokens := make([]*sessionstore.Token, 0, 3)
	for i := 0; i < 4; i++ {
		s := &Session{
			ID:   fmt.Sprint(i),
			Name: fmt.Sprint(`name `, i),
		}
		token, e := m.Put(ctx, s.ID, platform, s)
		if i == 3 {
			if e == nil {
				t.Fatal(`full but put success`)
			}
		} else {
			if e != nil {
				t.Fatal(e)
			}
			tokens = append(tokens, token)
		}
	}
	for i, token := range tokens {
		t0, s, e := m.Get(ctx, token.Access)
		if e != nil {
			t.Fatal(e)
		}
		if t0.Access != token.Access {
			t.Fatal(`Access not equal`)
		}
		if t0.Refresh != token.Refresh {
			t.Fatal(`Refresh not equal`)
		}
		if t0.AccessDeadline != token.AccessDeadline {
			t.Fatal(`AccessDeadline not equal`)
		}
		if t0.RefreshDeadline != token.RefreshDeadline {
			t.Fatal(`RefreshDeadline not equal`)
		}
		if t0.Deadline != token.Deadline {
			t.Fatal(`Deadline not equal`)
		}
		s0 := s.(*Session)
		id := fmt.Sprint(i)
		if s0.ID != id {
			t.Fatal(`ID not equal`)
		}
		name := `name ` + id
		if s0.Name != name {
			t.Fatal(`Name not equal`)
		}
	}

	time.Sleep(time.Second * 2)

	for _, token := range tokens {
		_, _, e := m.Get(ctx, token.Access)
		if e != cryptoer.ErrExpired {
			t.Fatal(`not ErrExpired`, e)
		}
	}

	time.Sleep(time.Second * 2)
	for _, token := range tokens {
		_, _, e := m.Get(ctx, token.Access)
		if e != cryptoer.ErrNotExistsToken {
			t.Fatal(`not ErrExpired`, e)
		}
	}
}
func TestManager(t *testing.T) {
	testManager(t, store.NewMemory(3))
}
func testManagerRefresh(t *testing.T, store sessionstore.Store) {
	opts := []sessionstore.Option{
		sessionstore.WithStore(store),
	}
	w0 := time.Second * 2
	w1 := time.Second * 2
	w2 := time.Second
	if _, ok := store.(*bbolt.Store); ok {
		opts = append(opts,
			sessionstore.WithAccess(time.Second*1),
			sessionstore.WithRefresh(time.Second*5),
			sessionstore.WithDeadline(time.Second*10),
		)
		w0 = time.Second * 3
		w1 = time.Second * 3
		w2 = time.Second * 4
	} else {
		opts = append(opts,
			sessionstore.WithAccess(time.Second*1),
			sessionstore.WithRefresh(time.Second*3),
			sessionstore.WithDeadline(time.Second*4),
		)
	}
	m := sessionstore.New(Coder{},
		opts...,
	)
	ctx := context.Background()
	platform := `test`
	tokens := make([]*sessionstore.Token, 0, 3)
	last := time.Now()
	for i := 0; i < 4; i++ {
		s := &Session{
			ID:   fmt.Sprint(i),
			Name: fmt.Sprint(`name `, i),
		}
		token, e := m.Put(ctx, s.ID, platform, s)
		if i == 3 {
			if e == nil {
				t.Fatal(`full but put success`)
			}
		} else {
			if e != nil {
				t.Fatal(e)
			}
			tokens = append(tokens, token)
		}
	}
	for i, token := range tokens {
		t0, s, e := m.Get(ctx, token.Access)
		if e != nil {
			t.Fatal(e)
		}
		if t0.Access != token.Access {
			t.Fatal(`Access not equal`)
		}
		if t0.Refresh != token.Refresh {
			t.Fatal(`Refresh not equal`)
		}
		if t0.AccessDeadline != token.AccessDeadline {
			t.Fatal(`AccessDeadline not equal`)
		}
		if t0.RefreshDeadline != token.RefreshDeadline {
			t.Fatal(`RefreshDeadline not equal`)
		}
		if t0.Deadline != token.Deadline {
			t.Fatal(`Deadline not equal`)
		}
		s0 := s.(*Session)
		id := fmt.Sprint(i)
		if s0.ID != id {
			t.Fatal(`ID not equal`)
		}
		name := `name ` + id
		if s0.Name != name {
			t.Fatal(`Name not equal`)
		}
	}
	fmt.Println(`set get `, time.Since(last))

	time.Sleep(w0)
	last = time.Now()
	for i, token := range tokens {
		_, _, e := m.Get(ctx, token.Access)
		if e != cryptoer.ErrExpired {
			t.Fatal(`not ErrExpired`, e)
		}

		_, _, e = m.Refresh(ctx, token.Access, token.Access)
		if e == nil {
			t.Fatal(`Refresh not match but success`)
		}

		t0, _, e := m.Refresh(ctx, token.Access, token.Refresh)
		if e != nil {
			t.Fatal(`Refresh`, e)
		}

		_, _, e = m.Get(ctx, token.Access)
		if e != cryptoer.ErrNotExistsToken {
			t.Fatal(`not ErrNotExistsToken`, e)
		}

		_, s, e := m.Get(ctx, t0.Access)
		if e != nil {
			t.Fatal(`Get Refresh err`, e)
		}

		s0 := s.(*Session)
		id := fmt.Sprint(i)
		if s0.ID != id {
			t.Fatal(`ID not equal`)
		}
		name := `name ` + id
		if s0.Name != name {
			t.Fatal(`Name not equal`)
		}

		tokens[i] = t0
	}
	fmt.Println(`refresh `, time.Since(last))
	time.Sleep(w1)

	last = time.Now()
	for i, token := range tokens {
		_, _, e := m.Get(ctx, token.Access)
		if e != cryptoer.ErrExpired {
			t.Fatal(`not ErrExpired`, e)
		}

		t0, s, e := m.Refresh(ctx, token.Access, token.Refresh)
		if e != nil {
			t.Fatal(`Refresh`, e)
		}
		s0 := s.(*Session)
		id := fmt.Sprint(i)
		if s0.ID != id {
			t.Fatal(`ID not equal`)
		}
		name := `name ` + id
		if s0.Name != name {
			t.Fatal(`Name not equal`)
		}

		tokens[i] = t0
	}
	fmt.Println(`refresh2 `, time.Since(last))

	time.Sleep(w2)
	for _, token := range tokens {
		_, _, e := m.Refresh(ctx, token.Access, token.Refresh)
		if e == nil {
			t.Fatal(`Refresh success on deadline`)
		}
	}
}
func TestMemoryRefresh(t *testing.T) {
	testManagerRefresh(t, store.NewMemory(3))
}
