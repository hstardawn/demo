package confession

import (
	"app/dao/model"
	"app/dao/repo"
	"github.com/spf13/cast"
	"github.com/zjutjh/mygo/jwt"
	"reflect"
	"runtime"
	"strings"
	"time"

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
	Info     struct{}          `name:"发布表白帖子" desc:"发布表白帖子"`
	Request  CreateApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response CreateApiResponse // API响应数据 (Body中的Data部分)
}

type CreateApiRequest struct {
	Body struct {
		Content     string    `json:"content" binding:"required" validate:"max=500, min=1" desc:"帖子内容"`
		IsAnonymous *int8     `json:"is_anonymous" binding:"required" desc:"是否匿名"`
		IsVisible   *int8     `json:"is_visible" binding:"required" desc:"是否可见"`
		Images      []string  `json:"images" desc:"图片"`
		PublishTime time.Time `json:"publish_time" binding:"required" desc:"预期发布时间"`
	}
}

type CreateApiResponse struct{}

// Run Api业务逻辑执行点
func (c *CreateApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewConfessionRepo()
	u := repo.NewUserRepo()
	now := time.Now()
	request := c.Request.Body
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)

	// 发布表白
	status := 0
	if !request.PublishTime.After(now) {
		status = 1
		request.PublishTime = now
	}
	user, err := u.FindByID(ctx, uid)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if len(request.Images) > 9 {
		nlog.Pick().WithContext(ctx).Warn("上传图片数量过多")
		return comm.CodeOutOfLimited
	}
	anon := 0
	if request.IsAnonymous != nil && *request.IsAnonymous == 1 {
		user.Name = "匿名用户"
		anon = 1
	}
	vis := 1
	if request.IsVisible != nil && *request.IsVisible == 0 {
		vis = 0
	}

	newPost := model.Confession{
		UserID:       uid,
		Name:         user.Name,
		Content:      request.Content,
		IsVisible:    int8(vis),
		IsAnonymous:  int8(anon),
		ImageUrls:    strings.Join(request.Images, ","),
		Status:       int32(status),
		ScheduleTime: request.PublishTime,
	}
	err = r.CreateConfession(ctx, &newPost)
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
