package gate

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panjichang1990/tianzong"
	"github.com/panjichang1990/tianzong/constant"
	"github.com/panjichang1990/tianzong/helper"
	"github.com/panjichang1990/tianzong/service"
	"github.com/panjichang1990/tianzong/session"
	"github.com/panjichang1990/tianzong/tzlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//实现网关服务器的通用功能
const (
	DefaultGrpcCacheLen              = 10
	DefaultMsgCacheLen               = 10
	DefaultClientActiveCheckInternal = 10
	DefaultClientOverTimeInternal    = 60
	DefaultTcpPort                   = 8888
	DefaultHttpPort                  = 8887
)

type mGate struct {
	service.GateServer
	//子服务链接池长度 DefaultGrpcCacheLen
	GrpcConnCacheLen int
	//队列缓存长度 DefaultMsgCacheLen
	MsgCacheLen int
	//监测子服务活跃间隔 DefaultClientActiveCheckInternal
	ClientActiveCheckInternal int64
	//子服务过期时间 DefaultClientOverTimeInternal
	ClientOverTimeInternal int64
	//项目ID
	ProjectId int32
	//确保每个子服务传递grpc 在单线程中进行
	clientQueue    chan *service.RegisterClientReq
	disClientQueue chan string
	menuQueue      chan *service.RegisterMenuReq
	zoneQueue      chan *service.RegisterZoneReq

	//内存缓存部分数据
	menuTree map[string][]string
	events   map[string][]string
	clients  map[string]*childInstance

	web *gin.Engine
	//tcp监听端口 默认 DefaultTcpPort
	TcpPort int
	//http监听端口 默认 DefaultHttpPort
	HttpPort    int
	Host        string
	handler     IGate
	adminCache  *sync.Map
	authAddress string
	authConn    *grpc.ClientConn
}

func (g *mGate) getAddress() string {
	return fmt.Sprintf("%v:%v", g.Host, g.TcpPort)
}

type IGate interface {
	OnClientReg(clientAddress string) //子服务注册触发
	OnMenuRegister(uri string)        //路由注册触发
	GetAdmin(adminId int32) tianzong.IAdmin
}

type RouteKey struct {
	Method string
	Uri    string
}

type childInstance struct {
	ch        *grpc.ClientConn
	address   string
	heartUnix int64
}

type childInstanceArr []*childInstance

func (a childInstanceArr) Len() int           { return len(a) }
func (a childInstanceArr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a childInstanceArr) Less(i, j int) bool { return a[i].heartUnix > a[j].heartUnix }

func (g *mGate) getAuthConn() service.AuthClient {
	if g.authConn == nil {
		tmp, err := grpc.Dial(g.authAddress, grpc.WithInsecure())
		if err != nil {
			return nil
		}
		g.authConn = tmp
	}
	return service.NewAuthClient(g.authConn)
}

func (cp *childInstance) Get() (*grpc.ClientConn, error) {

	if cp.ch == nil {
		tmp, err := grpc.Dial(cp.address, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		cp.ch = tmp
	}
	return cp.ch, nil
}

func (g *mGate) RegisterClient(_ context.Context, in *service.RegisterClientReq) (*service.RegisterRep, error) {
	if len(in.Address) == 0 {
		return &service.RegisterRep{
			Code: 500,
			Msg:  "address err",
		}, nil
	}
	g.clientQueue <- in
	return &service.RegisterRep{
		Code: 1,
		Msg:  "success",
	}, nil
}

func (g *mGate) DisRegisterClient(_ context.Context, in *service.DisRegisterClientReq) (*service.RegisterRep, error) {
	if len(in.Address) == 0 {
		return &service.RegisterRep{
			Code: 500,
			Msg:  "address err",
		}, nil
	}
	g.disClientQueue <- in.Address
	return &service.RegisterRep{
		Code: 1,
		Msg:  "success",
	}, nil
}

func (g *mGate) RegisterMenu(_ context.Context, in *service.RegisterMenuReq) (*service.RegisterRep, error) {
	tzlog.I(">>> %v", in.Address)
	if len(in.Address) == 0 {
		return &service.RegisterRep{
			Code: 500,
			Msg:  "address err",
		}, nil
	}
	g.menuQueue <- in
	return &service.RegisterRep{
		Code: 1,
		Msg:  "success",
	}, nil
}
func (g *mGate) Ping(_ context.Context, in *service.PingReq) (*service.PingRep, error) {
	return &service.PingRep{
		IsRegister: g.setChildHeart(in.Address),
	}, nil

}

func (g *mGate) ClearAuth(_ context.Context, in *service.ClearAuthReq) (*service.ClearAuthRep, error) {
	g.adminCache.Delete(in.GetAdminId())
	return &service.ClearAuthRep{Code: constant.SuccessCode}, nil
}

func (g *mGate) checkClientActive() {
	t := time.Now().Unix()
	for _, v := range g.clients {
		if v.heartUnix+g.ClientOverTimeInternal < t {
			g.disClientQueue <- v.address
		}
	}
	time.AfterFunc(time.Duration(g.ClientActiveCheckInternal)*time.Second, g.checkClientActive)
}

func (g *mGate) setChildHeart(address string) bool {
	v, ok := g.clients[address]
	if ok {
		v.heartUnix = time.Now().Unix()
	}
	return ok
}

func (g *mGate) getClient(address string) *childInstance {
	if v, ok := g.clients[address]; ok {
		return v
	}
	return nil
}

func (g *mGate) getClientByAddress(address string) *grpc.ClientConn {
	if v, ok := g.clients[address]; ok {
		conn, err := v.Get()
		if err != nil || conn == nil {
			return nil
		}
		return conn
	}
	return nil
}

func (g *mGate) getClientConn(uri string) *grpc.ClientConn {
	if v, ok := g.menuTree[uri]; ok {
		tmpSort := childInstanceArr{}
		for _, address := range v {
			cli := g.getClient(address)
			if cli != nil {
				tmpSort = append(tmpSort)
			}
		}
		if tmpSort.Len() > 0 {
			sort.Sort(tmpSort)
			return g.getClientByAddress(tmpSort[0].address)
		}
	}
	return nil
}

func (g *mGate) MPublish(topic string, header map[string]string, body string) {
	if v, ok := g.events[topic]; ok {
		for _, address := range v {
			conn := g.getClientByAddress(address)
			if conn != nil {
				child := service.NewChildClient(conn)
				rep, err := child.Publish(context.Background(), &service.PublishInfo{
					Topic:  topic,
					Header: header,
					Body:   body,
				})
				if err != nil {
					tzlog.E("%v event publish err to %v err: %v", topic, address, err)
					continue
				}
				if rep.Code != 1 {
					tzlog.W("%v event publish to %v code err", topic, address)
				} else {
					tzlog.I("%v event publish to %v success", topic, address)
				}
				//g.putClientConn(address, conn) //回收链接
			}
		}
	}
}

func (g *mGate) afterRun() {

}

func (g *mGate) beforeRun() {
	if g.GrpcConnCacheLen == 0 {
		g.GrpcConnCacheLen = DefaultGrpcCacheLen
	}
	if len(g.Host) == 0 {
		g.Host = "127.0.0.1"
	}
	if g.MsgCacheLen == 0 {
		g.MsgCacheLen = DefaultMsgCacheLen
	}
	if g.ClientOverTimeInternal == 0 {
		g.ClientOverTimeInternal = DefaultClientOverTimeInternal
	}
	if g.ClientActiveCheckInternal == 0 {
		g.ClientActiveCheckInternal = DefaultClientActiveCheckInternal
	}
	if g.TcpPort == 0 {
		g.TcpPort = DefaultTcpPort
	}
	if g.HttpPort == 0 {
		g.HttpPort = DefaultHttpPort
	}
	g.clientQueue = make(chan *service.RegisterClientReq, g.MsgCacheLen)
	g.menuQueue = make(chan *service.RegisterMenuReq, g.MsgCacheLen)
	g.disClientQueue = make(chan string, g.MsgCacheLen)
	g.clients = make(map[string]*childInstance)
	g.menuTree = make(map[string][]string)
	g.events = make(map[string][]string)
	g.registerToAuth()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				tzlog.E("channel err %v", err)
			}
		}()
		for {
			select {
			case c := <-g.clientQueue:
				if v, ok := g.clients[c.Address]; !ok || v == nil {
					g.clients[c.Address] = &childInstance{address: c.Address, heartUnix: time.Now().Unix()}
				}
				//事件注册
				for _, v := range c.Events {
					if _, ok := g.events[v]; !ok {
						g.events[v] = make([]string, 0)
					}
					for _, address := range g.events[v] {
						if address == c.Address {
							break
						}
						g.events[v] = append(g.events[v], c.Address)
					}
				}
			case addr := <-g.disClientQueue:
				//delete client
				if _, ok := g.clients[addr]; ok {
					delete(g.clients, addr)
				}
				//delete event
				for k, v := range g.events {
					for index, vv := range v {
						if vv == addr {
							g.events[k] = append(g.events[k][:index], g.events[k][index+1:]...)
						}
					}
					if len(g.events[k]) == 0 {
						delete(g.events, k)
					}
				}
				//delete route
				for k, v := range g.menuTree {
					for index, address := range v {
						if address == addr {
							g.menuTree[k] = append(g.menuTree[k][:index], g.menuTree[k][index+1:]...)
						}
					}
					if len(g.menuTree[k]) == 0 {
						delete(g.menuTree, k)
					}
				}
			case m := <-g.menuQueue:
				for _, v := range m.Data {
					tzlog.I("register route %v[%v]", v.Uri, m.GameId)
					if _, ok := g.menuTree[v.Uri]; !ok {
						g.menuTree[v.Uri] = make([]string, 0)
					}
					g.menuTree[v.Uri] = append(g.menuTree[v.Uri], m.Address)
				}

			}
		}
	}()

	session.Init()
	go g.ping()
}

func (g *mGate) registerToAuth() {
	tzlog.I("注册至auth", g.authAddress)
	auth := g.getAuthConn()
	if auth == nil {
		return
	}
	_, _ = auth.AuthRegister(context.Background(), &service.AuthRegisterReq{
		Address: g.getAddress(),
	})
}

func (g *mGate) ping() {
	conn := g.getAuthConn()
	if conn == nil {
		return
	}
	rep, err := conn.Ping(context.Background(), &service.AuthPingReq{
		Address: g.getAddress(),
	})
	if err != nil {
		tzlog.I("ping auth fail ", err)
	} else {
		if !rep.IsRegister {
			g.registerToAuth()
		}
	}
	time.AfterFunc(10*time.Second, g.ping)
}

func (g *mGate) Run() {
	g.beforeRun()
	go func() {
		tzlog.I("tcp run ", g.TcpPort)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", g.TcpPort))
		if err != nil {
			fmt.Printf("监听端口失败: %s", err)
			return
		}
		s := grpc.NewServer()
		service.RegisterGateServer(s, g)
		reflection.Register(s)
		err = s.Serve(lis)
		if err != nil {
			fmt.Printf("开启服务失败: %s", err)
			return
		}
	}()
	g.web = gin.New()
	g.web.NoRoute(g.Center)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", g.HttpPort),
		Handler: g.web,
	}
	tzlog.I("http run ", g.HttpPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		tzlog.E("http服务启动异常 %v", err)
	}
	return
}

func (g *mGate) context() context.Context {
	//r, _ := context.WithTimeout(context.Background(), 60*time.Second)
	return context.Background()
}

func (g *mGate) checkAuth(ctx *gin.Context) (int, *gateAuth, string) {
	sess := session.Get(ctx)
	token := helper.GetString(sess.Get(tokenName))
	id := helper.GetInt32(sess.Get(idName))
	//token := ctx.Request.Header.Get(tokenName)
	if v, ok := g.adminCache.Load(id); ok {
		if info, ok1 := v.(*gateAuth); ok1 {
			if !info.checkIp(ctx.ClientIP()) {
				return constant.NeedLoginCode, nil, constant.NeedLoginMsg
			}
			if !info.checkToken(token) {
				return constant.NeedLoginCode, nil, constant.NeedLoginMsg
			}
			return constant.SuccessCode, info, ""
		}
	}
	conn := g.getAuthConn()
	if conn == nil {
		return constant.SysAuthErr, nil, constant.SysErrMsg
	}
	rep, err := conn.Check(context.Background(), &service.CheckReq{
		Token:     token,
		ProjectId: g.ProjectId,
		AdminId:   id,
		Address:   fmt.Sprintf("%v:%v", g.Host, g.TcpPort),
	})
	if err != nil {
		tzlog.W("鉴权异常 %v", err)
		return constant.SysAuthErr, nil, constant.SysErrMsg
	}
	if rep.Code != constant.SuccessCode {
		return constant.AuthCodeErr, nil, rep.Msg
	}
	info := &gateAuth{
		adminId:        rep.Data.AdminId,
		adminName:      rep.Data.AdminName,
		activeIp:       ctx.ClientIP(),
		token:          token,
		lastActiveTime: time.Now().Unix(),
	}
	_ = sess.Set(idName, id)
	_ = sess.Set(tokenName, token)
	return constant.SuccessCode, info, ""
}

func (g *mGate) Center(ctx *gin.Context) {
	checkCode, authInfo, errMsg := g.checkAuth(ctx)
	if checkCode != constant.SuccessCode {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": checkCode,
			"msg":  errMsg,
		})
	}
	admin := g.handler.GetAdmin(authInfo.adminId)
	if admin == nil {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 300,
			"msg":  "auth err",
		})
		return
	}
	pth := routerPath(ctx.Request.RequestURI)
	conn := g.getClientConn(pth)
	if conn == nil {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 404,
			"msg":  "no router",
		})
		return
	}

	client := service.NewChildClient(conn)
	req := g.buildDoReq(ctx)
	req.AdminId = authInfo.adminId
	req.AdminName = authInfo.adminName
	req.AdminJson = admin.ToJson()
	rep, err := client.Do(g.context(), req)
	if err != nil {
		tzlog.W("grpc err %v", err)
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 500,
			"msg":  "server err",
		})
		return
	}
	ctx.Writer.Header().Set("Content-Type", rep.ContentType)
	ctx.String(http.StatusOK, rep.GetBody())
}

func (g *mGate) buildDoReq(ctx *gin.Context) *service.DoReq {
	r := &service.DoReq{
		Header:   make(map[string]*service.Value, 0),
		Query:    make(map[string]*service.Value, 0),
		PostForm: make(map[string]*service.Value, 0),
	}
	for k, v := range ctx.Request.Header {
		r.Header[k] = &service.Value{V: v}
	}
	for k, v := range ctx.Request.URL.Query() {
		r.Query[k] = &service.Value{V: v}
	}
	_ = ctx.Request.ParseForm()
	for k, v := range ctx.Request.PostForm {
		r.PostForm[k] = &service.Value{V: v}
	}
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		tzlog.W("read body err %v", err)
	}
	r.Body = string(body)
	return r
}

func routerPath(pth string) string {
	index := strings.Index(pth, "?")
	if index > 0 {
		pth = pth[:index]
	}
	index = strings.Index(pth, ".")
	if index > 0 {
		pth = pth[:index]
	}
	return pth
}

var defaultGate = new(mGate)

func RegisterGateHandler(handler IGate) {
	defaultGate.handler = handler
}

func RegisterGateWebHandler(handler *gin.Engine) {
	defaultGate.web = handler
}

func Run() {
	defaultGate.Run()
}

func SetHttpPort(port int) {
	defaultGate.HttpPort = port
}

func SetTcpPort(port int) {
	defaultGate.TcpPort = port
}

func SetAuthAddress(address string) {
	defaultGate.authAddress = address
}
