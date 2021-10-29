package tianzong

import (
	"context"
	"encoding/json"
	"github.com/panjichang1990/tianzong/constant"
	"github.com/panjichang1990/tianzong/helper"
	"github.com/panjichang1990/tianzong/service"
	"github.com/panjichang1990/tianzong/tzlog"
)

type Context struct {
	Request  *service.DoReq
	Response *service.DoRep
	close    chan int
	context.Context
	stop bool
}

func (c *Context) Stop() {
	c.stop = true
}

func (c *Context) IsStop() bool {
	return c.stop
}

func (c *Context) Query(key string) string {
	value, ok := c.Request.Query[key]
	if ok && len(value.V) > 0 {
		return value.V[0]
	}
	return ""
}

func (c *Context) QueryArray(key string) []string {
	value, ok := c.Request.Query[key]
	if ok {
		return value.V
	}
	return []string{}
}

func (c *Context) PostForm(key string) string {
	if v, ok := c.Request.PostForm[key]; ok && len(v.V) > 0 {
		return v.V[0]
	}
	return ""
}

func (c *Context) PostFormArray(key string) []string {
	value, ok := c.Request.PostForm[key]
	if ok {
		return value.V
	}
	return []string{}
}

func (c *Context) JSON(data interface{}) {
	c.Response.ContentType = constant.TypeJSON
	body, err := json.Marshal(data)
	if err != nil {
		tzlog.W("json marshal err %v", err)
	}
	c.Response.Body = string(body)
}

func (c *Context) String(data interface{}) {
	c.Response.ContentType = constant.TypeText
	c.Response.Body = helper.GetString(data)
}
