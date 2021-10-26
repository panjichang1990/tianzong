package auth

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panjichang1990/tianzong/constant"
	"github.com/panjichang1990/tianzong/service"
	"github.com/panjichang1990/tianzong/tzlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"sync"
	"time"
)

//实现鉴权服务器通用的功能

const (
	DefaultGateOverTimeInternal = 60
	DefaultTcpPort              = 5055
	DefaultHttpPort             = 5066
)

type mAuth struct {
	service.AuthServer
	handler IAuth
	gateBox *sync.Map
	//网关心跳超时 DefaultGateOverTimeInternal
	GateOverTimeInternal int64
	ginEngine            *gin.Engine
	//鉴权服务TCP端口 DefaultTcpPort
	TcpPort int
	//Http端口 DefaultHttpPort
	HttpPort int
}

type gateClient struct {
	address   string
	heartUnix int64
	conn      *grpc.ClientConn
}

func (gc *gateClient) getConn() service.GateClient {
	if gc.conn == nil {
		conn, err := grpc.Dial(gc.address)
		if err != nil {
			tzlog.W("get gate conn err", err)
			return nil
		}
		gc.conn = conn
	}
	return service.NewGateClient(gc.conn)
}

type IAuthAdmin interface {
	GetAdminId() int32
	GetAdminName() string
	GetToken() string
	CheckToken(token string) bool
	IsActive() bool
	CheckProjectPower(projectId int32) bool
	Logout()
	GetHeader() string
}

type IAuth interface {
	GetAdmin(id int32) IAuthAdmin
	Logout(id int32)
}

type MAuth struct {
}

func (ma *MAuth) GetAdmin(id int32) IAuthAdmin {
	tzlog.I("获取角色信息 %v", id)
	return nil
}

func (ma *MAuth) Logout(id int32) {
	tzlog.I("账号退出 %v", id)
}

func (a *mAuth) clearGateCache(id int32, withOut string) {
	req := &service.ClearAuthReq{
		AdminId: id,
	}
	a.gateBox.Range(func(key, value interface{}) bool {
		if key.(string) != withOut {
			conn := value.(*gateClient).getConn()
			if conn != nil {
				_, _ = conn.ClearAuth(context.Background(), req)
			}
		}
		return true
	})
}

func (a *mAuth) Check(_ context.Context, in *service.CheckReq) (*service.CheckRep, error) {
	admin := a.handler.GetAdmin(in.AdminId)
	if admin == nil {
		return &service.CheckRep{
			Code: constant.AuthCodeErr,
			Msg:  "admin no exist",
		}, nil
	}
	if !admin.CheckToken(in.Token) {
		return &service.CheckRep{
			Code: constant.AuthCodeErr,
			Msg:  "token err",
		}, nil
	}
	if !admin.CheckProjectPower(in.ProjectId) {
		return &service.CheckRep{
			Code: constant.AuthCodeErr,
			Msg:  "no power",
		}, nil
	}
	a.clearGateCache(in.AdminId, in.Address)
	return &service.CheckRep{
		Code: constant.SuccessCode,
		Msg:  constant.SuccessMsg,
		Data: &service.AdminInfo{
			AdminId:   admin.GetAdminId(),
			AdminName: admin.GetAdminName(),
			Header:    admin.GetHeader(),
		},
	}, nil
}

func (a *mAuth) Logout(_ context.Context, in *service.LogoutReq) (*service.LogoutRep, error) {
	a.handler.Logout(in.AdminId)
	admin := a.handler.GetAdmin(in.AdminId)
	if admin == nil {
		return &service.LogoutRep{
			Code: constant.AuthCodeErr,
			Msg:  "admin no exist",
		}, nil
	}
	admin.Logout()
	a.clearGateCache(in.AdminId, "")
	return &service.LogoutRep{Code: constant.SuccessCode}, nil
}

func (a *mAuth) AuthRegister(_ context.Context, in *service.AuthRegisterReq) (*service.AuthRegisterRep, error) {
	tzlog.I("gate 注册 %v", in.Address)
	a.gateBox.Store(in.Address, &gateClient{
		address:   in.Address,
		heartUnix: time.Now().Unix(),
	})
	return &service.AuthRegisterRep{Code: constant.SuccessCode}, nil
}

func (a *mAuth) AuthDisRegister(_ context.Context, in *service.AuthRegisterReq) (*service.AuthRegisterRep, error) {
	tzlog.I("gate 注销 %v", in.Address)
	a.gateBox.Delete(in.Address)
	return &service.AuthRegisterRep{Code: constant.SuccessCode}, nil
}

func (a *mAuth) Ping(_ context.Context, in *service.AuthPingReq) (*service.AuthPingRep, error) {
	v, ok := a.gateBox.Load(in.Address)
	if ok {
		v.(*gateClient).heartUnix = time.Now().Unix()
	}
	return &service.AuthPingRep{IsRegister: ok}, nil
}

func (a *mAuth) checkGate() {
	t := time.Now().Unix()
	a.gateBox.Range(func(key, value interface{}) bool {
		if value.(*gateClient).heartUnix+a.GateOverTimeInternal < t {
			a.gateBox.Delete(key)
		}
		return true
	})
	time.AfterFunc(10*time.Second, a.checkGate)
}

func (a *mAuth) beforeRun() {
	if a.GateOverTimeInternal == 0 {
		a.GateOverTimeInternal = DefaultGateOverTimeInternal
	}
	if a.TcpPort == 0 {
		a.TcpPort = DefaultTcpPort
	}
	if a.HttpPort == 0 {
		a.HttpPort = DefaultHttpPort
	}
	go a.checkGate()
}

var defaultAuth = &mAuth{gateBox: &sync.Map{}}

func RegisterAuthWebHandler(handler *gin.Engine) {
	defaultAuth.ginEngine = handler
}

func RegisterAuthHandler(handler IAuth) {
	defaultAuth.handler = handler
}

func SetTcpPort(port int) {
	defaultAuth.TcpPort = port
}

func SetHttpPort(port int) {
	defaultAuth.HttpPort = port
}

func Run() {
	defaultAuth.Run()
}

func (a *mAuth) Run() {
	if a.handler == nil {
		tzlog.W("未注册auth handler")
		a.handler = new(MAuth)
	}
	if a.ginEngine == nil {
		tzlog.W("未注册http handler")
		a.ginEngine = gin.New()
	}
	a.beforeRun()
	go func() {
		tzlog.I("tcp run %v", a.TcpPort)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", a.TcpPort))
		if err != nil {
			fmt.Printf("监听端口失败: %s", err)
			return
		}
		s := grpc.NewServer()
		service.RegisterAuthServer(s, a)
		reflection.Register(s)
		err = s.Serve(lis)
		if err != nil {
			fmt.Printf("开启服务失败: %s", err)
			return
		}
	}()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", a.HttpPort),
		Handler: a.ginEngine,
	}
	tzlog.I("http run %v", a.HttpPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		tzlog.E("http服务启动异常 %v", err)
	}
	return
}
