package session

import (
	"github.com/gin-gonic/gin"
	"path/filepath"
	"tianzong/session/driver"
	"tianzong/tzlog"
)

var globalSessions *driver.Manager

// SessionConfig holds session related config
type Config struct {
	SessionOn                    bool
	SessionProvider              string
	SessionName                  string
	SessionGCMaxLifetime         int64
	SessionProviderConfig        string
	SessionCookieLifeTime        int
	SessionAutoSetCookie         bool
	SessionDomain                string
	SessionDisableHTTPOnly       bool // used to allow for cross domain cookies/javascript cookies.
	SessionEnableSidInHTTPHeader bool // enable store/get the sessionId into/from http headers
	SessionNameInHTTPHeader      string
	SessionEnableSidInURLQuery   bool // enable get the sessionId from Url Query params
}

var DefaultConf = Config{
	SessionOn:                    false,
	SessionProvider:              "memory",
	SessionName:                  "mySessionID",
	SessionGCMaxLifetime:         3600,
	SessionProviderConfig:        "",
	SessionDisableHTTPOnly:       false,
	SessionCookieLifeTime:        0, //set cookie default is the browser life
	SessionAutoSetCookie:         true,
	SessionDomain:                "",
	SessionEnableSidInHTTPHeader: false, // enable store/get the sessionId into/from http headers
	SessionNameInHTTPHeader:      "mySessionID",
	SessionEnableSidInURLQuery:   false, // enable get the sessionId from Url Query params
}

func Init() {
	//	sessionConf := conf.AppConf.GetString("session", "SessionProviderConfig")
	cof := new(driver.ManagerConfig)
	cof.CookieName = DefaultConf.SessionName
	cof.EnableSetCookie = DefaultConf.SessionAutoSetCookie
	cof.Gclifetime = DefaultConf.SessionGCMaxLifetime
	cof.CookieLifeTime = DefaultConf.SessionCookieLifeTime
	cof.ProviderConfig = filepath.ToSlash("")
	cof.DisableHTTPOnly = DefaultConf.SessionDisableHTTPOnly
	cof.Domain = DefaultConf.SessionDomain
	cof.EnableSidInHTTPHeader = DefaultConf.SessionEnableSidInHTTPHeader
	cof.SessionNameInHTTPHeader = DefaultConf.SessionNameInHTTPHeader
	cof.EnableSidInURLQuery = DefaultConf.SessionEnableSidInURLQuery
	var err error
	globalSessions, err = driver.NewManager("memory", cof)
	if err != nil {
		tzlog.W("启动session异常 %v", err)
		return
	}
	go globalSessions.GC()
}

func Get(ctx *gin.Context) driver.Store {
	sess, err := globalSessions.SessionStart(ctx.Writer, ctx.Request)
	if err != nil {
		tzlog.W("开启session异常 %v", err)
		return nil
	}
	return sess
}

func DestorySession(ctx *gin.Context) {
	globalSessions.SessionDestroy(ctx.Writer, ctx.Request)
}
