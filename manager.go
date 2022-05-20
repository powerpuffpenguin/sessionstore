package sessionstore

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	protoc_session "github.com/powerpuffpenguin/sessionstore/sessionstore/session"

	"github.com/powerpuffpenguin/sessionstore/cryptoer"
	"google.golang.org/protobuf/proto"
)

type Manager struct {
	coder Coder
	opts  options
}

func New(coder Coder, opt ...Option) (m *Manager) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	if opts.refresh < opts.access {
		opts.refresh = opts.access
	}
	return &Manager{
		coder: coder,
		opts:  opts,
	}
}
func (m *Manager) Close() error {
	return m.opts.store.Close()
}

// 驗證簽名是否有效
func (m *Manager) Verify(token string) (playdata string, e error) {
	i := strings.LastIndex(token, `.`)
	if i == -1 {
		e = cryptoer.ErrInvalidToken
		return
	}
	value := token[:i]
	sign := token[i+1:]

	e = m.opts.method.Verify(m.opts.key, StringToBytes(value), sign)
	if e != nil {
		return
	}
	playdata = value
	return
}

// 簽名數據
func (m *Manager) Sin(playdata string) (token string, e error) {
	sign, e := m.opts.method.Sign(m.opts.key, StringToBytes(playdata))
	if e != nil {
		return
	}
	return playdata + `.` + sign, nil
}

// 返回 token 關聯的 後端原始數據
func (m *Manager) GetRaw(ctx context.Context, access string) (key string, token *Token, b []byte, e error) {
	playdata, e := m.Verify(access)
	if e != nil {
		return
	}
	i := strings.LastIndex(playdata, `.`)
	if i == -1 {
		e = cryptoer.ErrInvalidToken
		return
	}
	key = playdata[:i]

	b, e = m.opts.store.Get(ctx, key)
	if e != nil {
		return
	}
	var raw protoc_session.Raw
	e = proto.Unmarshal(b, &raw)
	if e != nil {
		return
	}
	if raw.Token == nil {
		e = errors.New(`raw.Token nil`)
		return
	} else if raw.Data == nil {
		e = errors.New(`raw.Data nil`)
		return
	} else if raw.Token.Access != access {
		e = cryptoer.ErrNotExistsToken
		return
	}
	token = NewToken(
		raw.Token.Access, raw.Token.Refresh,
		raw.Token.AccessDeadline, raw.Token.RefreshDeadline,
		raw.Token.Deadline,
	)
	b = raw.Data
	return
}

// 返回 token 關聯的 session 數據
func (m *Manager) Get(ctx context.Context, access string) (token *Token, session interface{}, e error) {
	_, token, b, e := m.GetRaw(ctx, access)
	if e != nil {
		return
	}
	// token
	if token.IsDeleted() {
		e = cryptoer.ErrNotExistsToken
		return
	} else if token.IsExpired() {
		e = cryptoer.ErrExpired
		return
	}

	// session
	session, e = m.coder.Unmarshal(b)
	return
}

// 刪除 token
func (m *Manager) Delete(ctx context.Context, access string) (e error) {
	playdata, e := m.Verify(access)
	if e != nil {
		return
	}
	i := strings.LastIndex(playdata, `.`)
	if i == -1 {
		e = cryptoer.ErrInvalidToken
		return
	}
	key := playdata[:i]
	e = m.opts.store.Del(context.Background(), key, access)
	return
}

// 刪除指定 用戶 id 的所有 session
func (m *Manager) DeleteID(ctx context.Context, id string) (e error) {
	prefix := base64.RawURLEncoding.EncodeToString(StringToBytes(id)) + `.`
	e = m.opts.store.DelPrefix(ctx, prefix)
	return
}

// 刪除指定用戶 id 在 指定平臺 platform 的所有 session
func (m *Manager) DeletePlatform(ctx context.Context, id, platform string) (e error) {
	prefix := base64.RawURLEncoding.EncodeToString(StringToBytes(id)) + `.` +
		base64.RawURLEncoding.EncodeToString(StringToBytes(platform)) + `.`
	e = m.opts.store.DelPrefix(ctx, prefix)
	return
}

// 創建一個 token
func (m *Manager) NewToken(prefix string) (token string, e error) {
	u, e := uuid.NewUUID()
	if e != nil {
		return
	}
	playdata := prefix + `.` + base64.RawURLEncoding.EncodeToString(u[:])
	token, e = m.Sin(playdata)
	return
}

// 創建 session 關聯的 token
func (m *Manager) Put(ctx context.Context, id, platform string, session interface{}) (token *Token, e error) {
	// create token
	key := base64.RawURLEncoding.EncodeToString(StringToBytes(id)) + `.` +
		base64.RawURLEncoding.EncodeToString(StringToBytes(platform))
	access, e := m.NewToken(key)
	if e != nil {
		return
	}
	refresh, e := m.NewToken(key)
	if e != nil {
		return
	}
	now := time.Now()
	refreshDeadline := now.Add(m.opts.refresh)
	var deadline int64
	if m.opts.deadline != 0 {
		deadline = now.Add(m.opts.deadline).Unix()
	}
	token = NewToken(access, refresh,
		now.Add(m.opts.access).Unix(), refreshDeadline.Unix(),
		deadline,
	)
	// marshal session
	b, e := m.coder.Marshal(session)
	if e != nil {
		return
	}

	// marshal
	b, e = proto.Marshal(&protoc_session.Raw{
		Token: &protoc_session.Token{
			Access:          token.Access,
			Refresh:         token.Refresh,
			AccessDeadline:  token.AccessDeadline,
			RefreshDeadline: token.RefreshDeadline,
			Deadline:        token.Deadline,
		},
		Data: b,
	})
	if e != nil {
		return
	}

	e = m.opts.store.Put(ctx, key, b, refreshDeadline)
	return
}

func (m *Manager) Refresh(ctx context.Context, access, refresh string) (token *Token, session interface{}, e error) {
	key, token, b, e := m.GetRaw(ctx, access)
	if e != nil {
		return
	}

	// token
	if token.IsDeleted() {
		e = cryptoer.ErrNotExistsToken
		return
	} else if refresh != token.Refresh {
		e = cryptoer.ErrRefreshTokenNotMatched
		return
	} else if !token.CanRefresh() {
		e = cryptoer.ErrCannotRefresh
		return
	}

	access, e = m.NewToken(key)
	if e != nil {
		return
	}
	refresh, e = m.NewToken(key)
	if e != nil {
		return
	}
	now := time.Now()
	refreshDeadline := now.Add(m.opts.refresh)
	token = NewToken(access, refresh,
		now.Add(m.opts.access).Unix(), refreshDeadline.Unix(),
		token.Deadline,
	)
	// unmarshal
	session, e = m.coder.Unmarshal(b)
	if e != nil {
		return
	}

	// marshal
	raw := &protoc_session.Raw{
		Token: &protoc_session.Token{
			Access:          token.Access,
			Refresh:         token.Refresh,
			AccessDeadline:  token.AccessDeadline,
			RefreshDeadline: token.RefreshDeadline,
		},
		Data: b,
	}
	b, e = proto.Marshal(raw)
	if e != nil {
		return
	}
	e = m.opts.store.Put(ctx, key, b, refreshDeadline)
	return
}
