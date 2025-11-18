package user

import (
	"app/dao/model"
	"app/dao/repo"
	"github.com/zjutjh/mygo/jwt"
	"reflect"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// RegisterHandler API router注册点
func RegisterHandler() gin.HandlerFunc {
	api := RegisterApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegister).Pointer()).Name()] = api
	return hfRegister
}

type RegisterApi struct {
	Info     struct{}            `name:"注册" desc:"用户注册"`
	Request  RegisterApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response RegisterApiResponse // API响应数据 (Body中的Data部分)
}

type RegisterApiRequest struct {
	Body struct {
		Username string `json:"username" binding:"required" desc:"用户名"`
		Password string `json:"password" binding:"required" validate:"max=12, min=6" desc:"密码"`
		Name     string `json:"name" binding:"required" validate:"max=10, min=1" desc:"昵称"`
		Avatar   string `json:"avatar" binding:"required" desc:"头像"`
	}
}

type RegisterApiResponse struct {
	Token string `json:"token"`
}

// Run Api业务逻辑执行点
func (r *RegisterApi) Run(ctx *gin.Context) kit.Code {
	u := repo.NewUserRepo()
	register := r.Request.Body

	// 查找是否存在用户
	user, err := u.FindByUsername(ctx, register.Username)
	if err != nil {
		return comm.CodeSaveError
	}
	if user != nil {
		return comm.CodeUserExisted
	}
	password, err := comm.HashPassword(register.Password)
	if err != nil {
		return comm.CodeHashError
	}

	// 创建用户
	newUser := model.User{
		Username: register.Username,
		Password: password,
		Avatar:   register.Avatar,
		Name:     register.Name,
	}
	err = u.CreatUser(ctx, &newUser)
	if err != nil {
		return comm.CodeDatabaseError
	}

	// 生成token
	token, err := jwt.Pick().GenerateToken(strconv.FormatInt(newUser.ID, 10))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("token生成失败")
		return comm.CodeMiddlewareServiceError
	}
	r.Response.Token = token

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (r *RegisterApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBind(&r.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfRegister API执行入口
func hfRegister(ctx *gin.Context) {
	api := &RegisterApi{}
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
