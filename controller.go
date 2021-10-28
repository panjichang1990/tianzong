package tianzong

type IController interface {
	Init()
	Prepare(ctx *Context)
	GetMap() map[routerInfo]HandlerFunc
	GetModel() string
	Finish(ctx *Context)
}

type ControllerBase struct {
	IController
	HandlerMap map[routerInfo]HandlerFunc
}

type routerInfo struct {
	Router string
	Name   string
	Desc   string
}

var ControllerBox = make([]IController, 0)

func (c *ControllerBase) RegisterController(controller IController) {
	ControllerBox = append(ControllerBox, controller)
}

func (c *ControllerBase) RegisterMethod(router, name, desc string, handler HandlerFunc) {
	if c.HandlerMap == nil {
		c.HandlerMap = make(map[routerInfo]HandlerFunc)
	}
	c.HandlerMap[routerInfo{
		Router: router,
		Name:   name,
		Desc:   desc,
	}] = handler
}

func (c *ControllerBase) GetMap() map[routerInfo]HandlerFunc {
	return c.HandlerMap
}

func (c *ControllerBase) Prepare(ctx *Context) {

}

func (c *ControllerBase) Finish(ctx *Context) {

}
