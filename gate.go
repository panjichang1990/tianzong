package tianzong

type IGate interface {
	OnMenuRegister(RouteInfo) //路由注册触发
	GetAdmin(adminId int32) IAdmin
}

type RouteInfo struct {
	Uri       string
	ParentUri string
	Name      string
	Desc      string
}
