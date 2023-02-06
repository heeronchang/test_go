package memory

import (
	"container/list"
	"go_test/session"
	"sync"
	"time"
)

var provider = &Provider{list: list.New()}

func init() {
	provider.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", provider)
}

//	type Session interface {
//		Set(key, value interface{}) error
//		Get(key interface{}) interface{}
//		Delete(key interface{}) error
//		SessionID() string
//	}
type SessionStore struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

func (s *SessionStore) Set(key, value interface{}) error {
	s.value[key] = value
	provider.SessionUpdate(s.sid)
	return nil
}

func (s *SessionStore) Get(key interface{}) interface{} {
	provider.SessionUpdate(s.sid)
	if v, ok := s.value[key]; ok {
		return v
	}

	return nil
}

func (s *SessionStore) Delete(key interface{}) error {
	delete(s.value, key)
	provider.SessionUpdate(s.sid)
	return nil
}

func (s *SessionStore) SessionID() string {
	return s.sid
}

//	type Provider interface {
//		SessionInit(sid string) (Session, error)
//		SessionRead(sid string) (Session, error)
//		SessionDestroy(sid string) error
//		SessionGC(maxLifeTime int64)
//	}
type Provider struct {
	lock     sync.Mutex
	sessions map[string]*list.Element
	list     *list.List
}

func (p *Provider) SessionInit(sid string) (session.Session, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	v := make(map[interface{}]interface{}, 0)
	s := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := p.list.PushBack(s)
	p.sessions[sid] = element

	return s, nil
}

func (p *Provider) SessionRead(sid string) (session.Session, error) {
	if element, ok := p.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := p.SessionInit(sid)
		return sess, err
	}
}

func (p *Provider) SessionDestroy(sid string) error {
	if element, ok := p.sessions[sid]; ok {
		delete(p.sessions, sid)
		p.list.Remove(element)
		return nil
	}
	return nil
}

func (p *Provider) SessionGC(maxLifetime int64) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		element := p.list.Back()
		if element == nil {
			break
		}
		// 如果最后一个session没有超时，则前面的也不会超时，因为 SessionUpdate 方法会把最新访问的session移到链表最前面
		// 如果最有一个session超时了，则销毁该session，然后遍历它前一个session
		if element.Value.(*SessionStore).timeAccessed.Unix()+maxLifetime < time.Now().Unix() {
			p.list.Remove(element)
			delete(p.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

// SessionUpdate 更新sessoin最后访问时间，并把session移到链表最前面
func (p *Provider) SessionUpdate(sid string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if element, ok := p.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		p.list.MoveToFront(element)
		return nil
	}
	return nil
}
