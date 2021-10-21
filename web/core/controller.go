package core

import (
	"github.com/gin-gonic/gin"
	"github.com/panjichang1990/tianzong/web/middleware"
	"html/template"
)

type Controller interface {
	Init()
	GetMap() map[*Method]gin.HandlerFunc
	Prepare(*gin.Context)
	GetModel() string
	Finish(*gin.Context)
}

type Method struct {
	Method string
	Path   string
}

var box = make([]Controller, 0)

var myTemplateFunc = template.FuncMap{}

var defultDebug = false

//Model 返回http引擎
func SetDebug(debug bool) {
	defultDebug = debug
}

func Model(mid ...gin.HandlerFunc) *gin.Engine {
	if defultDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.SetFuncMap(myTemplateFunc)
	for _, v := range mid {
		r.Use(v)
	}
	if len(box) > 0 {
		for _, v := range box {
			v.Init()
			tmp := v.GetMap()
			if len(tmp) > 0 {
				for key, val := range tmp {
					if v.GetModel() != "" {
						key.Path = "/" + v.GetModel() + key.Path
					}
					switch key.Method {
					case Get:
						r.GET(key.Path, v.Prepare, val, v.Finish)
					case Post:
						r.POST(key.Path, v.Prepare, val, v.Finish)
					case Delete:
						r.DELETE(key.Path, v.Prepare, val, v.Finish)
					case Head:
						r.HEAD(key.Path, v.Prepare, val, v.Finish)
					case Put:
						r.PUT(key.Path, v.Prepare, val, v.Finish)
					case NoRoute:
						r.NoRoute(v.Prepare, val, v.Finish)
					case Any:
						r.Any(key.Path, v.Prepare, val, v.Finish)
					case GetAndPost:
						r.GET(key.Path, v.Prepare, val, v.Finish)
						r.POST(key.Path, v.Prepare, val, v.Finish)
					default:
						//TODO 方便路由注册 可直接 continue
						r.Any(key.Path, v.Prepare, val, v.Finish)
					}
				}
			}
		}
	}
	return r
}
