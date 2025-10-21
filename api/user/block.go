package user

import (
	"app/dao/repo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/jwt"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"reflect"
	"runtime"

	"app/comm"
)

// BlockHandler API router注册点
func BlockHandler() gin.HandlerFunc {
	api := BlockApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfBlock).Pointer()).Name()] = api
	return hfBlock
}

type BlockApi struct {
	Info     struct{}         `name:"拉黑用户" desc:"拉黑用户"`
	Request  BlockApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response BlockApiResponse // API响应数据 (Body中的Data部分)
}

type BlockApiRequest struct {
	Query struct {
		BlockedId int64 `form:"blocked_id" binding:"required" desc:"拉黑用户ID"`
	}
}

type BlockApiResponse struct{}

// Run Api业务逻辑执行点
func (b *BlockApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewBlockRepo()
	request := b.Request.Query

	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	if uid == request.BlockedId {
		return comm.CodeParameterInvalid // 自己不能拉黑自己
	}

	isBlocked, err := r.IsBlocked(ctx, uid, request.BlockedId)
	if err != nil {
		return comm.CodeSearchError
	}
	if isBlocked {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("拉黑关系已存在")
		return comm.CodeBlockExisted
	}
	err = r.BlockUser(ctx, uid, request.BlockedId)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("拉黑失败")
		return comm.CodeBlockError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (b *BlockApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&b.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfBlock API执行入口
func hfBlock(ctx *gin.Context) {
	api := &BlockApi{}
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
