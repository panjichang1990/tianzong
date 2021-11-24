package tianzong

type IController interface {
	Init()
	Prepare(ctx *Context)
	GetMap() map[*routerInfo]HandlerFunc
	GetModel() string
	Finish(ctx *Context)
}

type ControllerBase struct {
	IController
	HandlerMap map[*routerInfo]HandlerFunc
}

type routerInfo struct {
	Router    string
	Name      string
	Desc      string
	ParentUri string
	Ext       map[string]string
}

var ControllerBox = make([]IController, 0)

func (c *ControllerBase) RegisterController(controller IController) {
	ControllerBox = append(ControllerBox, controller)
}

func (c *ControllerBase) RegisterMethod(uri, parentUri, name, desc string, ext map[string]string, handler HandlerFunc) {
	if c.HandlerMap == nil {
		c.HandlerMap = make(map[*routerInfo]HandlerFunc)
	}
	c.HandlerMap[c.SetExt(&routerInfo{
		Router:    uri,
		Name:      name,
		Desc:      desc,
		ParentUri: parentUri,
		Ext:       ext,
	})] = handler
}

func (c *ControllerBase) SetExt(route *routerInfo) *routerInfo {
	return route
}

func (c *ControllerBase) GetMap() map[*routerInfo]HandlerFunc {
	return c.HandlerMap
}

func (c *ControllerBase) Prepare(ctx *Context) {

}

func (c *ControllerBase) Finish(ctx *Context) {

}
