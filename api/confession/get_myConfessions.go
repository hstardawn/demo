package confession

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
	"strings"

	"app/comm"
)

// GetMyConfessionsHandler API router注册点
func GetMyConfessionsHandler() gin.HandlerFunc {
	api := GetMyConfessionsApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetMyConfessions).Pointer()).Name()] = api
	return hfGetMyConfessions
}

type GetMyConfessionsApi struct {
	Info     struct{}                    `name:"获取个人帖子" desc:"获取个人帖子"`
	Request  GetMyConfessionsApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetMyConfessionsApiResponse // API响应数据 (Body中的Data部分)
}

type GetMyConfessionsApiRequest struct {
	Query struct {
		PageSize int `form:"page_size" binding:"required" desc:"页容量"`
		PageNum  int `form:"page_num" binding:"required" desc:"当前页码"`
	}
}

type GetMyConfessionsApiResponse struct {
	TotalCount int          `json:"total_count" desc:"帖子数目"`
	Posts      []AnswerPost `json:"posts" desc:"帖子列表"`
}

// Run Api业务逻辑执行点
func (g *GetMyConfessionsApi) Run(ctx *gin.Context) kit.Code {
	p := repo.NewPostRepo()
	request := g.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	list, _, err := p.GetMyPosts(ctx, request.PageNum, request.PageSize, uid)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取表白列表失败")
		return comm.CodeListError
	}

	filteredPosts := make([]AnswerPost, 0)
	for _, post := range list {
		newPost := AnswerPost{
			Id:        post.ID,
			UserId:    post.UserID,
			Name:      post.Name,
			Content:   post.Content,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
			ImageUrl:  strings.Split(post.ImageUrls, ","),
		}
		filteredPosts = append(filteredPosts, newPost)
	}

	g.Response = GetMyConfessionsApiResponse{
		TotalCount: len(list),
		Posts:      filteredPosts,
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetMyConfessionsApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfGetMyConfessions API执行入口
func hfGetMyConfessions(ctx *gin.Context) {
	api := &GetMyConfessionsApi{}
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
