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

// CreateHandler API router注册点
func CreateHandler() gin.HandlerFunc {
	api := CreateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfCreate).Pointer()).Name()] = api
	return hfCreate
}

type CreateApi struct {
	Info     struct{}          `name:"拉黑用户" desc:"拉黑用户"`
	Request  CreateApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response CreateApiResponse // API响应数据 (Body中的Data部分)
}

type CreateApiRequest struct {
	Query struct {
		BlockedId int64 `form:"blocked_id" binding:"required" desc:"拉黑用户ID"`
	}
}

type CreateApiResponse struct{}

// Run Api业务逻辑执行点
func (c *CreateApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewBlockRepo()
	request := c.Request.Query

	// 鉴权
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)
	if uid == request.BlockedId {
		return comm.CodeParameterInvalid // 自己不能拉黑自己
	}

	// 创建拉黑关系
	record, err := r.IsBlocked(ctx, uid, request.BlockedId)
	if err != nil {
		return comm.CodeSearchError
	}
	if record != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("拉黑关系已存在")
		return comm.CodeDatabaseError
	}
	err = r.BlockUser(ctx, uid, request.BlockedId)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("拉黑失败")
		return comm.CodeDatabaseError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (c *CreateApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&c.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfCreate API执行入口
func hfCreate(ctx *gin.Context) {
	api := &CreateApi{}
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
