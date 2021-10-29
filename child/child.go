package child

import (
	"context"
	"fmt"
	"github.com/panjichang1990/tianzong"
	"github.com/panjichang1990/tianzong/constant"
	"github.com/panjichang1990/tianzong/service"
	"github.com/panjichang1990/tianzong/tzlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type ChildServer struct {
	service.ChildServer
	GateAddress   []*gateAddress
	TcpPort       int
	Host          string
	CheckInternal int64
	address       string
	tcpConn       *grpc.ClientConn
	events        map[string]func(header map[string]string, body string)
	handlers      map[string][]tianzong.HandlerFunc
}

func (c *ChildServer) registerEvents(topic string, handler func(header map[string]string, body string)) {
	if c.events == nil {
		c.events = make(map[string]func(header map[string]string, body string))
	}
	if _, ok := c.events[topic]; ok {
		panic("事件" + topic + "已注册")
		return
	}
	c.events[topic] = handler
}

type gateAddress struct {
	address string
	conn    *grpc.ClientConn
}

func (gc *gateAddress) getConn() service.GateClient {
	if gc.conn == nil {
		conn, err := grpc.Dial(gc.address, grpc.WithInsecure())
		if err != nil {
			return nil
		}
		gc.conn = conn
	}
	return service.NewGateClient(gc.conn)
}

func (c *ChildServer) getAddress() string {
	if len(c.address) == 0 {
		c.address = fmt.Sprintf("%v:%v", c.Host, c.TcpPort)
	}
	return c.address
}

func (c *ChildServer) Do(ctx context.Context, in *service.DoReq) (*service.DoRep, error) {
	if v, ok := c.handlers[in.Uri]; ok {
		ctx := &tianzong.Context{Request: in, Response: &service.DoRep{}, Context: ctx}
		for _, f := range v {
			f(ctx)
		}
		return ctx.Response, nil
	}
	return &service.DoRep{Code: constant.ErrNoFondCode, Msg: constant.ErrNoFondMsg, Body: constant.ErrNoFondMsg, ContentType: constant.TypeText}, nil
}

func (c *ChildServer) ReloadChannel(context.Context, *service.ReloadChannelReq) (*service.ReloadRep, error) {

	return &service.ReloadRep{Code: constant.SuccessCode, Msg: constant.SuccessMsg}, nil
}

func (c *ChildServer) Publish(_ context.Context, in *service.PublishInfo) (*service.PublishRep, error) {
	if v, ok := c.events[in.Topic]; ok {
		v(in.Header, in.Body)
	}
	return &service.PublishRep{Code: constant.SuccessCode, Msg: constant.SuccessMsg}, nil
}

func (c *ChildServer) register() {
	for _, v := range c.GateAddress {

		gateConn := v.getConn()
		if gateConn == nil {
			continue
		}
		_, err := gateConn.RegisterClient(context.Background(), &service.RegisterClientReq{
			Address: c.getAddress(),
			GameId:  0,
		})
		if err != nil {
			tzlog.E("register to gate err: %v", err)
			return
		}
		_, err = gateConn.RegisterMenu(context.Background(), &service.RegisterMenuReq{
			GameId:  0,
			Address: c.getAddress(),
			Data:    c.getRouters(),
		})
	}

}

var defaultChild = &ChildServer{}

func (c *ChildServer) getRouters() []*service.MenuInfo {
	res := make([]*service.MenuInfo, 0)
	handlers := make(map[string][]tianzong.HandlerFunc)
	for _, v := range tianzong.ControllerBox {
		for k, h := range v.GetMap() {
			router := &service.MenuInfo{
				Name: k.Name,
				Desc: k.Desc,
				Uri:  "/" + v.GetModel() + k.Router,
			}
			res = append(res, router)
			if _, ok := handlers[router.Uri]; !ok {
				handlers[router.Uri] = []tianzong.HandlerFunc{v.Prepare, h, v.Finish}
			} else {
				tzlog.W("router is register %v", router.Uri)
			}
		}
	}
	c.handlers = handlers
	return res
}

func (c *ChildServer) check() {
	for _, address := range c.GateAddress {
		gateConn := address.getConn()
		if gateConn == nil {
			continue
		}
		rep, err := gateConn.Ping(context.Background(), &service.PingReq{Address: c.getAddress()})
		if err != nil {
			tzlog.W("gate ping err %v", err)
		} else {
			if !rep.IsRegister {
				c.register()
			}
		}
	}
	time.AfterFunc(time.Duration(c.CheckInternal)*time.Second, c.check)
}

func (c *ChildServer) Run() {
	if c.TcpPort == 0 {
		panic("需要设置TCP端口")
	}
	if len(c.Host) == 0 {
		c.Host = "127.0.0.1"
	}
	if len(c.GateAddress) == 0 {
		panic("需要设置网关地址")
	}
	c.register()
	lis, err := net.Listen("tcp", c.getAddress())
	if err != nil {
		tzlog.E("listen err :%v %v", c.getAddress(), err)
		return
	}
	go c.check()
	s := grpc.NewServer()
	service.RegisterChildServer(s, c)
	reflection.Register(s)
	err = s.Serve(lis)
	if err != nil {
		tzlog.E("open server err: %v", err)
		return
	}
	return
}

func SetPort(port int) {
	defaultChild.TcpPort = port
}

func SetGateAddress(address ...string) {
	tmp := make([]*gateAddress, 0)
	for _, v := range address {
		tmp = append(tmp, &gateAddress{
			address: v,
		})
	}
	defaultChild.GateAddress = tmp
}

func Run() {
	defaultChild.Run()
}
