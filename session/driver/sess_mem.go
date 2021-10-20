package driver

import (
	"container/list"
	"net/http"
	"sync"
	"time"
)

var memProvider = &MemProvider{list: list.New(), sessions: make(map[string]*list.Element)}

type MemSessionStore struct {
	sid          string                      //session id
	timeAccessed time.Time                   //last access time
	value        map[interface{}]interface{} //session store
	lock         sync.RWMutex
}

func (st *MemSessionStore) Set(key, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value[key] = value
	return nil
}

func (st *MemSessionStore) Get(key interface{}) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil
}

func (st *MemSessionStore) Delete(key interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
	return nil
}

func (st *MemSessionStore) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[interface{}]interface{})
	return nil
}

func (st *MemSessionStore) SessionID() string {
	return st.sid
}

func (st *MemSessionStore) SessionRelease(_ http.ResponseWriter) {
}

type MemProvider struct {
	lock        sync.RWMutex             // locker
	sessions    map[string]*list.Element // map in memory
	list        *list.List               // for gc
	maxLifeTime int64
	savePath    string
}

// SessionInit init memory session
func (mp *MemProvider) SessionInit(maxLifeTime int64, savePath string) error {
	mp.maxLifeTime = maxLifeTime
	mp.savePath = savePath
	return nil
}

// SessionRead get memory session store by sid
func (mp *MemProvider) SessionRead(sid string) (Store, error) {
	mp.lock.RLock()
	if element, ok := mp.sessions[sid]; ok {
		go func() {
			_ = mp.SessionUpdate(sid)
		}()
		mp.lock.RUnlock()
		return element.Value.(*MemSessionStore), nil
	}
	mp.lock.RUnlock()
	mp.lock.Lock()
	newSess := &MemSessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	element := mp.list.PushFront(newSess)
	mp.sessions[sid] = element
	mp.lock.Unlock()
	return newSess, nil
}

func (mp *MemProvider) SessionExist(sid string) bool {
	mp.lock.RLock()
	defer mp.lock.RUnlock()
	if _, ok := mp.sessions[sid]; ok {
		return true
	}
	return false
}

func (mp *MemProvider) SessionRegenerate(oldSid, sid string) (Store, error) {
	mp.lock.RLock()
	if element, ok := mp.sessions[oldSid]; ok {
		go func() {
			_ = mp.SessionUpdate(oldSid)
		}()
		mp.lock.RUnlock()
		mp.lock.Lock()
		element.Value.(*MemSessionStore).sid = sid
		mp.sessions[sid] = element
		delete(mp.sessions, oldSid)
		mp.lock.Unlock()
		return element.Value.(*MemSessionStore), nil
	}
	mp.lock.RUnlock()
	mp.lock.Lock()
	newSess := &MemSessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	element := mp.list.PushFront(newSess)
	mp.sessions[sid] = element
	mp.lock.Unlock()
	return newSess, nil
}

func (mp *MemProvider) SessionDestroy(sid string) error {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	if element, ok := mp.sessions[sid]; ok {
		delete(mp.sessions, sid)
		mp.list.Remove(element)
		return nil
	}
	return nil
}

func (mp *MemProvider) SessionGC() {
	mp.lock.RLock()
	for {
		element := mp.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*MemSessionStore).timeAccessed.Unix() + mp.maxLifeTime) < time.Now().Unix() {
			mp.lock.RUnlock()
			mp.lock.Lock()
			mp.list.Remove(element)
			delete(mp.sessions, element.Value.(*MemSessionStore).sid)
			mp.lock.Unlock()
			mp.lock.RLock()
		} else {
			break
		}
	}
	mp.lock.RUnlock()
}

func (mp *MemProvider) SessionAll() int {
	return mp.list.Len()
}

func (mp *MemProvider) SessionUpdate(sid string) error {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	if element, ok := mp.sessions[sid]; ok {
		element.Value.(*MemSessionStore).timeAccessed = time.Now()
		mp.list.MoveToFront(element)
		return nil
	}
	return nil
}

func init() {
	Register("memory", memProvider)
}
