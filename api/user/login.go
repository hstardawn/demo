package user

import (
	"app/dao/repo"
	"errors"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// LoginHandler API router注册点
func LoginHandler() gin.HandlerFunc {
	api := LoginApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfLogin).Pointer()).Name()] = api
	return hfLogin
}

type LoginApi struct {
	Info     struct{}         `name:"登陆" desc:"用户登录"`
	Request  LoginApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response LoginApiResponse // API响应数据 (Body中的Data部分)
}

type LoginApiRequest struct {
	Body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
}

type LoginApiResponse struct{}

// Run Api业务逻辑执行点
func (l *LoginApi) Run(ctx *gin.Context) kit.Code {
	u := repo.NewUserRepo()
	loginRequest := l.Request.Body
	db := ndb.Pick()
	if db == nil {
		return comm.CodeDatabaseError
	}

	user, err := u.FindByUsername(ctx, loginRequest.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comm.CodeUserNotFound
	}
	if !comm.CheckPassword(user.Password, loginRequest.Password) {
		return comm.CodePasswordError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (l *LoginApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&l.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfLogin API执行入口
func hfLogin(ctx *gin.Context) {
	api := &LoginApi{}
	err := api.Init(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
