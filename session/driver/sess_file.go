package driver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	fileProvider  = &FileProvider{}
	gcMaxLifeTime int64
)

type FileSessionStore struct {
	sid    string
	lock   sync.RWMutex
	values map[interface{}]interface{}
}

func (fs *FileSessionStore) Set(key, value interface{}) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values[key] = value
	return nil
}

func (fs *FileSessionStore) Get(key interface{}) interface{} {
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	if v, ok := fs.values[key]; ok {
		return v
	}
	return nil
}

func (fs *FileSessionStore) Delete(key interface{}) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	delete(fs.values, key)
	return nil
}

func (fs *FileSessionStore) Flush() error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values = make(map[interface{}]interface{})
	return nil
}

func (fs *FileSessionStore) SessionID() string {
	return fs.sid
}

func (fs *FileSessionStore) SessionRelease(_ http.ResponseWriter) {
	fileProvider.lock.Lock()
	defer fileProvider.lock.Unlock()
	b, err := EncodeGob(fs.values)
	if err != nil {
		SLogger.Println(err)
		return
	}
	_, err = os.Stat(path.Join(fileProvider.savePath, string(fs.sid[0]), string(fs.sid[1]), fs.sid))
	var f *os.File
	if err == nil {
		f, err = os.OpenFile(path.Join(fileProvider.savePath, string(fs.sid[0]), string(fs.sid[1]), fs.sid), os.O_RDWR, 0777)
		if err != nil {
			SLogger.Println(err)
			return
		}
	} else if os.IsNotExist(err) {
		f, err = os.Create(path.Join(fileProvider.savePath, string(fs.sid[0]), string(fs.sid[1]), fs.sid))
		if err != nil {
			SLogger.Println(err)
			return
		}
	} else {
		return
	}
	_ = f.Truncate(0)
	_, _ = f.Seek(0, 0)
	_, _ = f.Write(b)
	_ = f.Close()
}

// FileProvider File session provider
type FileProvider struct {
	lock        sync.RWMutex
	maxLifeTime int64
	savePath    string
}

func (fp *FileProvider) SessionInit(maxlifetime int64, savePath string) error {
	fp.maxLifeTime = maxlifetime
	fp.savePath = savePath
	return nil
}

func (fp *FileProvider) SessionRead(sid string) (Store, error) {
	if strings.ContainsAny(sid, "./") {
		return nil, nil
	}
	if len(sid) < 2 {
		return nil, errors.New("length of the sid is less than 2")
	}
	fileProvider.lock.Lock()
	defer fileProvider.lock.Unlock()

	err := os.MkdirAll(path.Join(fp.savePath, string(sid[0]), string(sid[1])), 0777)
	if err != nil {
		SLogger.Println(err.Error())
	}
	_, err = os.Stat(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	var f *os.File
	if err == nil {
		f, err = os.OpenFile(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid), os.O_RDWR, 0777)
	} else if os.IsNotExist(err) {
		f, err = os.Create(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	} else {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	_ = os.Chtimes(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid), time.Now(), time.Now())
	var kv map[interface{}]interface{}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = DecodeGob(b)
		if err != nil {
			return nil, err
		}
	}

	ss := &FileSessionStore{sid: sid, values: kv}
	return ss, nil
}

func (fp *FileProvider) SessionExist(sid string) bool {
	fileProvider.lock.Lock()
	defer fileProvider.lock.Unlock()

	_, err := os.Stat(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	return err == nil
}

func (fp *FileProvider) SessionDestroy(sid string) error {
	fileProvider.lock.Lock()
	defer fileProvider.lock.Unlock()
	_ = os.Remove(path.Join(fp.savePath, string(sid[0]), string(sid[1]), sid))
	return nil
}

func (fp *FileProvider) SessionGC() {
	fileProvider.lock.Lock()
	defer fileProvider.lock.Unlock()

	gcMaxLifeTime = fp.maxLifeTime
	_ = filepath.Walk(fp.savePath, gcPath)
}

func (fp *FileProvider) SessionAll() int {
	a := &activeSession{}
	err := filepath.Walk(fp.savePath, func(path string, f os.FileInfo, err error) error {
		return a.visit(path, f, err)
	})
	if err != nil {
		SLogger.Printf("filepath.Walk() returned %v\n", err)
		return 0
	}
	return a.total
}

func (fp *FileProvider) SessionRegenerate(oldsid, sid string) (Store, error) {
	fileProvider.lock.Lock()
	defer fileProvider.lock.Unlock()

	oldPath := path.Join(fp.savePath, string(oldsid[0]), string(oldsid[1]))
	oldSidFile := path.Join(oldPath, oldsid)
	newPath := path.Join(fp.savePath, string(sid[0]), string(sid[1]))
	newSidFile := path.Join(newPath, sid)

	_, err := os.Stat(newSidFile)
	if err == nil {
		return nil, fmt.Errorf("newsid %s exist", newSidFile)
	}

	err = os.MkdirAll(newPath, 0777)
	if err != nil {
		SLogger.Println(err.Error())
	}

	_, err = os.Stat(oldSidFile)
	if err == nil {
		b, err := ioutil.ReadFile(oldSidFile)
		if err != nil {
			return nil, err
		}

		var kv map[interface{}]interface{}
		if len(b) == 0 {
			kv = make(map[interface{}]interface{})
		} else {
			kv, err = DecodeGob(b)
			if err != nil {
				return nil, err
			}
		}

		_ = ioutil.WriteFile(newSidFile, b, 0777)
		_ = os.Remove(oldSidFile)
		_ = os.Chtimes(newSidFile, time.Now(), time.Now())
		ss := &FileSessionStore{sid: sid, values: kv}
		return ss, nil
	}

	newF, err := os.Create(newSidFile)
	if err != nil {
		return nil, err
	}
	_ = newF.Close()
	ss := &FileSessionStore{sid: sid, values: make(map[interface{}]interface{})}
	return ss, nil
}

func gcPath(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if (info.ModTime().Unix() + gcMaxLifeTime) < time.Now().Unix() {
		_ = os.Remove(path)
	}
	return nil
}

type activeSession struct {
	total int
}

func (as *activeSession) visit(_ string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f.IsDir() {
		return nil
	}
	as.total = as.total + 1
	return nil
}

func init() {
	Register("file", fileProvider)
}
