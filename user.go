package ppp

type Account struct {
}

//用户注册
func (U *Account) Add(request *User, resp *UserResult) error {
	resp.ReToken = "12345"
	return nil
}
