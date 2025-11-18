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

// UpdateHandler API router注册点
func UpdateHandler() gin.HandlerFunc {
	api := UpdateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUpdate).Pointer()).Name()] = api
	return hfUpdate
}

type UpdateApi struct {
	Info     struct{}          `name:"更新表白" desc:"更新表白"`
	Request  UpdateApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response UpdateApiResponse // API响应数据 (Body中的Data部分)
}

type UpdateApiRequest struct {
	Body struct {
		ConfessionId int64    `json:"post_id" binding:"required" desc:"帖子ID"`
		Content      string   `json:"content" validate:"max=500, min=1" desc:"内容"`
		Image        []string `json:"image"  desc:"图片"`
		IsAnonymous  *int8    `json:"is_anonymous"  desc:"匿名"`
		IsVisible    *int8    `json:"is_visible"  desc:"可见性"`
	}
}

type UpdateApiResponse struct{}

// Run Api业务逻辑执行点
func (u *UpdateApi) Run(ctx *gin.Context) kit.Code {
	p := repo.NewConfessionRepo()
	r := repo.NewUserRepo()
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	request := u.Request.Body

	// 鉴权
	uid := cast.ToInt64(id)
	user, err := r.FindByID(ctx, uid)
	if err != nil {
		return comm.CodeUserNotFound
	}

	// 更新表白
	record, err := p.FindConfessionByID(ctx, request.ConfessionId)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if record == nil {
		return comm.CodeConfessionNotFound
	}

	updates := make(map[string]any)

	// 内容
	if request.Content != "" {
		updates["content"] = request.Content
	}

	// 匿名
	if request.IsAnonymous != nil {
		if *request.IsAnonymous == 1 {
			updates["is_anonymous"] = 1
			updates["name"] = "匿名用户"
		} else {
			updates["is_anonymous"] = 0
			updates["name"] = user.Name
		}
	}

	// 可见性
	if request.IsVisible != nil {
		if *request.IsVisible == 0 {
			updates["is_visible"] = 0
		} else {
			updates["is_visible"] = 1
		}
	}

	// 图片
	if len(request.Image) > 9 {
		nlog.Pick().WithContext(ctx).Warn("图片数量超过限制")
		return comm.CodeOutOfLimited
	}
	if request.Image != nil {
		updates["image_urls"] = strings.Join(request.Image, ",")
	}

	// 更新
	err = p.UpdateConfession(ctx, request.ConfessionId, updates)
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
