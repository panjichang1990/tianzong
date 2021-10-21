package tianzong

type IAdmin interface {
	ToJson() string           //当前管理员的数据 转为json字符串
	CheckUri(uri string) bool //监测路由是否有权限访问
}
