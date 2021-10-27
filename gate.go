package tianzong

import "github.com/gin-gonic/gin"

type IGate interface {
	OnMenuRegister(RouteInfo) //路由注册触发
	GetAuthInfo(ctx *gin.Context) IAdmin
	DiyBack(ctx *gin.Context)
}

type RouteInfo struct {
	Uri       string
	ParentUri string
	Name      string
	Desc      string
}
