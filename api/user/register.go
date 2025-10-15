package user

import (
	"app/dao/model"
	"app/dao/query"
	"app/service"
	"github.com/zjutjh/mygo/ndb"
	"mime/multipart"
	"reflect"
	"runtime"

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
		Username string                `form:"username" binding:"required"`
		Password string                `form:"password" binding:"required"`
		Name     string                `form:"name" binding:"required"`
		Avatar   *multipart.FileHeader `form:"avatar" binding:"required"`
	}
}

type RegisterApiResponse struct {
	ID int64 `json:"id"`
}

// Run Api业务逻辑执行点
func (r *RegisterApi) Run(ctx *gin.Context) kit.Code {
	register := r.Request.Body
	db := ndb.Pick()
	if db == nil {
		return comm.CodeDatabaseError
	}
	userQuery := query.Use(db).User

	user, err := userQuery.Where(userQuery.Username.Eq(register.Username)).First()
	if err == nil && user != nil {
		return comm.CodeUserExisted
	}

	url, err := service.SaveUploadedImage(ctx, register.Avatar)
	if err != nil {
		return comm.CodeSaveError
	}
	password, err := comm.HashPassword(register.Password)
	if err != nil {
		return comm.CodeHashError
	}
	newUser := model.User{
		Username: register.Username,
		Password: password,
		Avatar:   url,
		Name:     register.Name,
	}

	err = userQuery.Create(&newUser)
	if err != nil {
		return comm.CodeDatabaseError
	}
	r.Response.ID = newUser.ID
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
