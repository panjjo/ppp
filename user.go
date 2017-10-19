package ppp

import "gopkg.in/mgo.v2/bson"

const (
	FC_ACCOUNT_REGIST string = "Account.Regist"
	FC_ACCOUNT_AUTH   string = "Account.Auth"
)

type Account struct {
}

//账户注册
//如果传入mchid表示直接绑定授权帐号
func (A *Account) Regist(request *User, resp *AccountResult) error {
	user := getUser(request.UserId, request.Type)
	if user.MchId == request.MchId {
		resp.Code = UserErrRegisted
		return nil
	}
	switch request.Type {
	case PAYTYPE_ALIPAY, PAYTYPE_WXPAY, PAYTYPE_PPP:
	default:
		resp.Code = SysErrParams
		return nil
	}
	if request.MchId != "" {
		//验证授权是否存在
		auth := getToken(request.MchId, request.Type)
		if auth.Id == "" {
			resp.Code = AuthErr
			return nil
		}
		request.Status = 1
	}
	if user.UserId != "" {
		//更新授权绑定
		updateUser(user.UserId, user.Type, bson.M{"$set": bson.M{"mchid": request.MchId}})
	} else {
		//新增
		request.Id = randomString(32)
		saveUser(*request)
	}
	resp.Data = *request
	return nil
}

//账户授权
//将账户和授权绑定
func (A *Account) Auth(request *AccountAuth, resp *Response) error {
	//查询用户
	var user User
	if user = getUser(request.UserId, request.Type); user.Id == "" {
		resp.Code = UserErrNotFount
		return nil
	}
	var auth authBase
	if auth = getToken(request.MchId, request.Type); auth.Id == "" {
		resp.Code = AuthErr
		return nil
	}
	user.MchId = request.MchId
	user.Status = auth.Status
	updateUser(user.UserId, user.Type, bson.M{"$set": user})
	return nil
}

//账户解绑
func (A *Account) UnBind(request *User, resp *Response) error {
	user := getUser(request.UserId, request.Type)
	if user.UserId == "" {
		resp.Code = UserErrNotFount
		return nil
	}
	updateUser(request.UserId, request.Type, bson.M{"$set": bson.M{"mchid": "", "status": UserWaitVerify}})
	return nil
}
