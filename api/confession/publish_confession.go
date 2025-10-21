package confession

import (
	"app/dao/model"
	"app/dao/repo"
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"reflect"
	"runtime"
	"strings"

	"app/comm"
)

// PublishConfessionHandler API router注册点
func PublishConfessionHandler() gin.HandlerFunc {
	api := PublishConfessionApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfPublishConfession).Pointer()).Name()] = api
	return hfPublishConfession
}

type PublishConfessionApi struct {
	Info     struct{}                     `name:"发布表白帖子" desc:"发布表白帖子"`
	Request  PublishConfessionApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response PublishConfessionApiResponse // API响应数据 (Body中的Data部分)
}

type PublishConfessionApiRequest struct {
	Body struct {
		Content   string   `json:"content" binding:"required" desc:"帖子内容"`
		Anonymous bool     `json:"anonymous" binding:"required" desc:"是否匿名"`
		Visible   bool     `json:"visible" binding:"required" desc:"是否可见"`
		Images    []string `json:"images" desc:"图片"`
	}
}

type PublishConfessionApiResponse struct{}

// Run Api业务逻辑执行点
func (p *PublishConfessionApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewPostRepo()
	request := p.Request.Body

	if len(request.Images) != 0 && len(request.Images) > 9 {
		nlog.Pick().WithContext(ctx).Warn("上传图片数量过多")
		return comm.CodeOutOfLimited
	}
	newPost := model.Post{
		Content:   request.Content,
		IsVisible: request.Visible,
		ImageUrls: strings.Join(request.Images, ","),
	}
	err := r.CreatePost(ctx, &newPost)
	if err != nil {
		return comm.CodeDatabaseError
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (p *PublishConfessionApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&p.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfPublishConfession API执行入口
func hfPublishConfession(ctx *gin.Context) {
	api := &PublishConfessionApi{}
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
