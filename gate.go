package tianzong

import "github.com/gin-gonic/gin"

type IGate interface {
	OnMenuRegister(RouteInfo) //路由注册触发
	GetAdmin(adminId int32) IAdmin
	GetAuthInfo(ctx *gin.Context) *AuthInfo
}

type RouteInfo struct {
	Uri       string
	ParentUri string
	Name      string
	Desc      string
}

type AuthInfo struct {
	AdminId    int32
	AdminToken string
}
