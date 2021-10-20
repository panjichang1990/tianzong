package gate

import "time"

const tokenName = `tz_token`

const idName = `tz_id`

type gateAuth struct {
	adminId        int32  //管理员ID
	adminName      string //管理员姓名
	activeIp       string //登录IP
	token          string //票据
	lastActiveTime int64
}

func (ga *gateAuth) checkIp(ip string) bool {
	return ga.activeIp == ip
}

func (ga *gateAuth) setActiveTime() {
	ga.lastActiveTime = time.Now().Unix()
}

func (ga *gateAuth) checkToken(token string) bool {
	return ga.token == token
}
