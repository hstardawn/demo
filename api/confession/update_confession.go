package confession

import (
	"app/dao/repo"
	"github.com/spf13/cast"
	"github.com/zjutjh/mygo/jwt"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// UpdateConfessionHandler API router注册点
func UpdateConfessionHandler() gin.HandlerFunc {
	api := UpdateConfessionApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUpdateConfession).Pointer()).Name()] = api
	return hfUpdateConfession
}

type UpdateConfessionApi struct {
	Info     struct{}                    `name:"更新表白" desc:"更新表白"`
	Request  UpdateConfessionApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response UpdateConfessionApiResponse // API响应数据 (Body中的Data部分)
}

type UpdateConfessionApiRequest struct {
	Body struct {
		PostId      int64    `json:"post_id" binding:"required" desc:"帖子ID"`
		Content     string   `json:"content"  desc:"内容"`
		Image       []string `json:"image"  desc:"图片"`
		IsAnonymous *bool    `json:"is_anonymous"  desc:"匿名"`
		IsVisible   *bool    `json:"is_visible"  desc:"可见性"`
	}
}

type UpdateConfessionApiResponse struct{}

// Run Api业务逻辑执行点
func (u *UpdateConfessionApi) Run(ctx *gin.Context) kit.Code {
	p := repo.NewPostRepo()
	r := repo.NewUserRepo()
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	request := u.Request.Body

	uid := cast.ToInt64(id)
	user, err := r.FindByID(ctx, uid)
	if err != nil {
		return comm.CodeUserNotFound
	}
	record, err := p.FindPostByID(ctx, request.PostId)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if record == nil {
		return comm.CodePostNotFound
	}

	if request.Content != "" {
		record.Content = request.Content
	}
	anon := false
	if request.IsAnonymous != nil {
		if *request.IsAnonymous {
			record.Name = "匿名用户"
			anon = true
		} else {
			record.Name = user.Name
		}
	}
	record.IsAnonymous = anon

	vis := true
	if request.IsVisible != nil && *request.IsVisible == false {
		vis = false
	}
	record.IsVisible = vis
	if len(request.Image) > 9 {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("图片数量超过限制")
		return comm.CodeOutOfLimited
	}
	record.ImageUrls = strings.Join(request.Image, ",")

	err = p.UpdatePost(ctx, record)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("更新失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (u *UpdateConfessionApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&u.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfUpdateConfession API执行入口
func hfUpdateConfession(ctx *gin.Context) {
	api := &UpdateConfessionApi{}
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
