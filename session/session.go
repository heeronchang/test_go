package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Session interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error
	SessionID() string
}

type Manager struct {
	cookieName  string
	lock        sync.Mutex
	provider    Provider
	maxlifetime int64
}

func NewManager(provideName, cookieName string, maxlifetime int64) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}

	return &Manager{
		provider:    provider,
		cookieName:  cookieName,
		maxlifetime: maxlifetime,
	}, nil
}

// genSessionId create sessionid
func genSessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}

// SessionStart 检测来访用户是否有与之对应的session，没有则创建
func (m *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		sid := genSessionId()
		session, _ = m.provider.SessionInit(sid)
		cookie := http.Cookie{Name: m.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(m.maxlifetime)}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = m.provider.SessionRead(sid)
	}
	return
}

// Session 根据sid 获取session
func (m *Manager) Session(w http.ResponseWriter, r *http.Request) (session Session, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	cookie, e := r.Cookie(m.cookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		session = nil
		err = e
		return
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	session, _ = m.provider.SessionRead(sid)
	return
}

// GC 销毁Session
func (m *Manager) GC() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.provider.SessionGC(m.maxlifetime)
	time.AfterFunc(time.Duration(m.maxlifetime), func() {
		m.GC()
	})
}

// Provider 表示session管理器底层存储结构
type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

var provides = make(map[string]Provider)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panic.
func Register(name string, provider Provider) {
	if provider == nil {
		panic("Session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("Session: Register called twice for provide " + name)
	}
	provides[name] = provider
}
