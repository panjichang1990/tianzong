package core

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Base struct {
	Controller
	Model  string
	GinMap map[*Method]gin.HandlerFunc
}

// RegisterController  注册控制器
func (c *Base) RegisterController(controller Controller) {
	box = append(box, controller)
}

// Init 控制器初始化方法
func (c *Base) Init() {

}

// Prepare 路由前置方法
func (c *Base) Prepare(*gin.Context) {

}

// Finish 后置方法
func (c *Base) Finish(ctx *gin.Context) {

	ctx.Abort()
}

func (c *Base) GetModel() string {
	return ""
}

// GetMap 获取控制器的路由组
func (c *Base) GetMap() map[*Method]gin.HandlerFunc {
	return c.GinMap
}

// Fetch 渲染
func (c *Base) Fetch(ctx *gin.Context, fileName string) {
	ctx.HTML(http.StatusOK, fileName, gin.H(ctx.Keys))
}

//String 输出字符串
func (c *Base) String(ctx *gin.Context, msg string) {
	ctx.String(http.StatusOK, msg)
}

// Json 输出json
func (c *Base) Json(ctx *gin.Context, param interface{}) {
	ctx.JSON(http.StatusOK, param)
}

//IsAjax 判断请求是ajax
func (c *Base) IsAjax(ctx *gin.Context) bool {
	return ctx.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

//IsPost 判断请求是post
func (c *Base) IsPost(ctx *gin.Context) bool {
	return ctx.Request.Method == Post
}

//IsGet 判断请求是get
func (c *Base) IsGet(ctx *gin.Context) bool {
	return ctx.Request.Method == Get
}

//RegisterTemplateFunc 声明用到的模板函数  最好不要和通用的重名 可放在初始化函数中声明
func (c *Base) RegisterTemplateFunc(name string, f interface{}) {
	//此处不做判断 同名函数按后注册为主
	myTemplateFunc[name] = f
}

//RegisterRouter 注册路由
func (c *Base) RegisterRouter(m *Method, f gin.HandlerFunc) {
	if c.GinMap == nil {
		c.GinMap = make(map[*Method]gin.HandlerFunc)
	}
	c.GinMap[m] = f
}

//MethodGet get
func (c *Base) MethodGet(path string) *Method {
	return &Method{
		Method: Get,
		Path:   path,
	}
}

//MethodNoRoute err404
func (c *Base) MethodNoRoute() *Method {
	return &Method{
		Method: NoRoute,
		Path:   "",
	}
}

//MethodPost Post
func (c *Base) MethodPost(path string) *Method {
	return &Method{
		Method: Post,
		Path:   path,
	}
}

//MethodAny any
func (c *Base) MethodAny(path string) *Method {
	return &Method{
		Method: Any,
		Path:   path,
	}
}

//MethodGetAndPost post and get
func (c *Base) MethodGetAndPost(path string) *Method {
	return &Method{
		Method: GetAndPost,
		Path:   path,
	}
}

//Assign 绑定模板变量
func (c *Base) Assign(ctx *gin.Context, name string, val interface{}) {
	ctx.Set(name, val)
}
