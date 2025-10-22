package block

import (
	"app/dao/repo"
	"github.com/spf13/cast"
	"github.com/zjutjh/mygo/jwt"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// UnblockHandler API router注册点
func UnblockHandler() gin.HandlerFunc {
	api := UnblockApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUnblock).Pointer()).Name()] = api
	return hfUnblock
}

type UnblockApi struct {
	Info     struct{}           `name:"取消拉黑" desc:"取消拉黑"`
	Request  UnblockApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response UnblockApiResponse // API响应数据 (Body中的Data部分)
}

type UnblockApiRequest struct {
	Query struct {
		BlockedId int64 `form:"blocked_id" binding:"required" desc:"被拉黑用户的Id"`
	}
}

type UnblockApiResponse struct{}

// Run Api业务逻辑执行点
func (u *UnblockApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewBlockRepo()
	request := u.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	isBlocked, err := r.IsBlocked(ctx, uid, request.BlockedId)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询拉黑关系失败")
		return comm.CodeDatabaseError
	}
	if !isBlocked {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("未拉黑，无法解除拉黑")
		return comm.CodeBlockNotExisted
	}
	err = r.UnBlockUser(ctx, uid, request.BlockedId)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("解除拉黑失败")
		return comm.CodeUnblockError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (u *UnblockApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&u.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfUnblock API执行入口
func hfUnblock(ctx *gin.Context) {
	api := &UnblockApi{}
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
