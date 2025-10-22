package user

import (
	"app/dao/model"
	"app/dao/repo"
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/jwt"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"reflect"
	"runtime"
	"strconv"

	"app/comm"
)

// UpdateHandler API router注册点
func UpdateHandler() gin.HandlerFunc {
	api := UpdateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUpdate).Pointer()).Name()] = api
	return hfUpdate
}

type UpdateApi struct {
	Info     struct{}          `name:"更新用户信息" desc:"更新用户信息"`
	Request  UpdateApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response UpdateApiResponse // API响应数据 (Body中的Data部分)
}

type UpdateApiRequest struct {
	Body struct {
		Username string `json:"username" binding:"required" desc:"用户名"`
		Password string `json:"password" desc:"密码"`
		Name     string `json:"name" desc:"昵称"`
		Avatar   string `json:"avatar" desc:"头像"`
	}
}

type UpdateApiResponse struct{}

// Run Api业务逻辑执行点
func (u *UpdateApi) Run(ctx *gin.Context) kit.Code {
	s := repo.NewUserRepo()
	updateRequest := u.Request.Body
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid, err := strconv.ParseInt(id, 10, 64)
	user, err := s.FindByID(ctx, uid)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if user == nil {
		return comm.CodeUserNotFound
	}

	newUser := &model.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		Username:  user.Username,
		Password:  user.Password,
		Name:      user.Name,
		Avatar:    user.Avatar,
	}
	if updateRequest.Password != "" {
		password, err := comm.HashPassword(updateRequest.Password)
		if err != nil {
			return comm.CodeHashError
		}
		newUser.Password = password
	}
	if updateRequest.Name != "" {
		newUser.Name = updateRequest.Name
	}
	if updateRequest.Avatar != "" {
		newUser.Avatar = updateRequest.Avatar
	}

	err = s.UpdateUser(ctx, newUser)
	if err != nil {
		return comm.CodeDatabaseError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (u *UpdateApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&u.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfUpdate API执行入口
func hfUpdate(ctx *gin.Context) {
	api := &UpdateApi{}
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
