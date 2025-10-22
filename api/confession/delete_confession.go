package confession

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

// DeleteConfessionHandler API router注册点
func DeleteConfessionHandler() gin.HandlerFunc {
	api := DeleteConfessionApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfDeleteConfession).Pointer()).Name()] = api
	return hfDeleteConfession
}

type DeleteConfessionApi struct {
	Info     struct{}                    `name:"删除表白" desc:"删除表白"`
	Request  DeleteConfessionApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response DeleteConfessionApiResponse // API响应数据 (Body中的Data部分)
}

type DeleteConfessionApiRequest struct {
	Query struct {
		PostId int64 `form:"post_id" binding:"required" desc:"删除帖子编号"`
	}
}

type DeleteConfessionApiResponse struct{}

// Run Api业务逻辑执行点
func (d *DeleteConfessionApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewPostRepo()
	request := d.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	record, err := r.FindPostByID(ctx, request.PostId)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查找帖子失败")
		return comm.CodePostNotFound
	}
	if record.UserID != uid {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("非该帖子主人")
		return comm.CodePermissionDenied
	}
	err = r.DeletePost(ctx, request.PostId)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("删除失败")
		return comm.CodeDatabaseError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (d *DeleteConfessionApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&d.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfDeleteConfession API执行入口
func hfDeleteConfession(ctx *gin.Context) {
	api := &DeleteConfessionApi{}
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
