package confession

import (
	"app/dao/repo"
	"context"
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

// LikeHandler API router注册点
func LikeHandler() gin.HandlerFunc {
	api := LikeApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfLike).Pointer()).Name()] = api
	return hfLike
}

type LikeApi struct {
	Info     struct{}        `name:"点赞" desc:"点赞"`
	Request  LikeApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response LikeApiResponse // API响应数据 (Body中的Data部分)
}

type LikeApiRequest struct {
	Query struct {
		ConfessionID int64 `form:"confession_id" binding:"required" desc:"表白ID"`
		Action       int   `form:"action" desc:"点赞行为"`
	}
}

type LikeApiResponse struct{}

// Run Api业务逻辑执行点
func (l *LikeApi) Run(ctx *gin.Context) kit.Code {
	req := l.Request.Query
	r := repo.NewLikeRepo()
	h := repo.NewHotRepo()

	// 获取当前用户ID
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)

	isLike, err := r.IsUserLiked(ctx, req.ConfessionID, uid)
	nowStatus := cast.ToBool(req.Action)
	if nowStatus == isLike {
		return comm.CodeRepeatAction
	}
	err = r.LikeAction(ctx, req.ConfessionID, uid, req.Action)
	if err != nil {
		return comm.CodeDatabaseError
	}
	err = h.UpdateLikeScore(context.Background(), req.ConfessionID, req.Action)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("更新热度失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (l *LikeApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&l.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfLike API执行入口
func hfLike(ctx *gin.Context) {
	api := &LikeApi{}
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
