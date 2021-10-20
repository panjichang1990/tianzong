package driver

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
)

var cookieProvider = &CookieProvider{}

type CookieSessionStore struct {
	sid    string
	values map[interface{}]interface{}
	lock   sync.RWMutex
}

func (st *CookieSessionStore) Set(key, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values[key] = value
	return nil
}

func (st *CookieSessionStore) Get(key interface{}) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.values[key]; ok {
		return v
	}
	return nil
}

func (st *CookieSessionStore) Delete(key interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.values, key)
	return nil
}

func (st *CookieSessionStore) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values = make(map[interface{}]interface{})
	return nil
}

func (st *CookieSessionStore) SessionID() string {
	return st.sid
}

func (st *CookieSessionStore) SessionRelease(w http.ResponseWriter) {
	encodedCookie, err := encodeCookie(cookieProvider.block, cookieProvider.config.SecurityKey, cookieProvider.config.SecurityName, st.values)
	if err == nil {
		cookie := &http.Cookie{Name: cookieProvider.config.CookieName,
			Value:    url.QueryEscape(encodedCookie),
			Path:     "/",
			HttpOnly: true,
			Secure:   cookieProvider.config.Secure,
			MaxAge:   cookieProvider.config.MaxAge}
		http.SetCookie(w, cookie)
	}
}

type cookieConfig struct {
	SecurityKey  string `json:"securityKey"`
	BlockKey     string `json:"blockKey"`
	SecurityName string `json:"securityName"`
	CookieName   string `json:"cookieName"`
	Secure       bool   `json:"secure"`
	MaxAge       int    `json:"maxage"`
}

type CookieProvider struct {
	maxLifeTime int64
	config      *cookieConfig
	block       cipher.Block
}

func (cp *CookieProvider) SessionInit(maxlifetime int64, config string) error {
	cp.config = &cookieConfig{}
	err := json.Unmarshal([]byte(config), cp.config)
	if err != nil {
		return err
	}
	if cp.config.BlockKey == "" {
		cp.config.BlockKey = string(generateRandomKey(16))
	}
	if cp.config.SecurityName == "" {
		cp.config.SecurityName = string(generateRandomKey(20))
	}
	cp.block, err = aes.NewCipher([]byte(cp.config.BlockKey))
	if err != nil {
		return err
	}
	cp.maxLifeTime = maxlifetime
	return nil
}

func (cp *CookieProvider) SessionRead(sid string) (Store, error) {
	maps, _ := decodeCookie(cp.block,
		cp.config.SecurityKey,
		cp.config.SecurityName,
		sid, cp.maxLifeTime)
	if maps == nil {
		maps = make(map[interface{}]interface{})
	}
	rs := &CookieSessionStore{sid: sid, values: maps}
	return rs, nil
}

func (cp *CookieProvider) SessionExist(_ string) bool {
	return true
}

func (cp *CookieProvider) SessionRegenerate(_, _ string) (Store, error) {
	return nil, nil
}

func (cp *CookieProvider) SessionDestroy(_ string) error {
	return nil
}

func (cp *CookieProvider) SessionGC() {
}

func (cp *CookieProvider) SessionAll() int {
	return 0
}

func (cp *CookieProvider) SessionUpdate(_ string) error {
	return nil
}

func init() {
	Register("cookie", cookieProvider)
}
