package comment

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

// DeleteHandler API router注册点
func DeleteHandler() gin.HandlerFunc {
	api := DeleteApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfDelete).Pointer()).Name()] = api
	return hfDelete
}

type DeleteApi struct {
	Info     struct{}          `name:"删除评论" desc:"删除评论"`
	Request  DeleteApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response DeleteApiResponse // API响应数据 (Body中的Data部分)
}

type DeleteApiRequest struct {
	Query struct {
		CommentID int64 `form:"comment_id" binding:"required" desc:"评论id"`
	}
}

type DeleteApiResponse struct{}

// Run Api业务逻辑执行点
func (d *DeleteApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewCommentRepo()
	request := d.Request.Query

	// 鉴权
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)

	// 查找评论
	comment, err := r.GetCommentByID(ctx, request.CommentID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查找评论失败")
		return comm.CodeDatabaseError
	}
	if comment == nil {
		return comm.CodeCommentNotFound
	}

	// 判断权限
	if comment.UserID != uid {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("无权删除帖子")
		return comm.CodePermissionDenied
	}

	// 删除评论
	err = r.DeleteCommentByID(ctx, request.CommentID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("删除失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (d *DeleteApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&d.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfDelete API执行入口
func hfDelete(ctx *gin.Context) {
	api := &DeleteApi{}
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
