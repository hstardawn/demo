package comment

import (
	"app/dao/model"
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
	Info     struct{}          `name:"评论帖子" desc:"评论帖子"`
	Request  CreateApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response CreateApiResponse // API响应数据 (Body中的Data部分)
}

type CreateApiRequest struct {
	Body struct {
		ConfessionID int64  `json:"confession_id" binding:"required" desc:"表白帖子ID"`
		ParentID     *int64 `json:"parent_id" desc:"父评论ID"`
		Content      string `json:"content" binding:"required" desc:"评论内容"`
	}
}

type CreateApiResponse struct{}

// Run Api业务逻辑执行点
func (c *CreateApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewCommentRepo()
	request := c.Request.Body

	// 鉴权
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)

	// 校验帖子是否存在
	confession, err := repo.NewConfessionRepo().FindConfessionByID(ctx, request.ConfessionID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查找失败")
		return comm.CodeDatabaseError
	}
	if confession == nil {
		return comm.CodeConfessionNotFound
	}

	// 查找父评论是否存在
	if *request.ParentID != 0 {
		record, err := r.GetCommentByID(ctx, *request.ParentID)
		if err != nil {
			return comm.CodeDatabaseError
		}
		if record == nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("父评论不存在")
			return comm.CodeCommentNotFound
		}
	}

	// 创建评论
	newComment := &model.Comment{
		Content:      request.Content,
		UserID:       uid,
		ParentID:     *request.ParentID,
		ConfessionID: request.ConfessionID,
	}
	err = r.CreateConfession(ctx, newComment)
	if err != nil {
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (c *CreateApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&c.Request.Body)
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
