package confession

import (
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

// GetConfessionHandler API router注册点
func GetConfessionHandler() gin.HandlerFunc {
	api := GetConfessionApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetConfession).Pointer()).Name()] = api
	return hfGetConfession
}

type GetConfessionApi struct {
	Info     struct{}                 `name:"获取表白" desc:"获取表白"`
	Request  GetConfessionApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetConfessionApiResponse // API响应数据 (Body中的Data部分)
}

type GetConfessionApiRequest struct {
	Query struct {
		PageSize int `form:"page_size" binding:"required" desc:"页容量"`
		PageNum  int `form:"page_num" binding:"required" desc:"当前页码"`
	}
}

type AnswerPost struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ImageUrl  []string  `json:"image_url"`
}
type GetConfessionApiResponse struct {
	TotalCount int          `json:"total_count" desc:"帖子数目"`
	Posts      []AnswerPost `json:"posts" desc:"帖子列表"`
}

// Run Api业务逻辑执行点
func (g *GetConfessionApi) Run(ctx *gin.Context) kit.Code {
	p := repo.NewPostRepo()
	b := repo.NewBlockRepo()
	request := g.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	list, _, err := p.GetAllPosts(ctx, request.PageNum, request.PageSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取表白列表失败")
		return comm.CodeListError
	}

	blockList, err := b.GetBlockedList(ctx, uid)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取拉黑列表失败")
		return comm.CodeListError
	}
	blockedIDs := make(map[int64]bool)
	for _, blk := range blockList {
		blockedIDs[blk.BlockedID] = true
	}

	filteredPosts := make([]AnswerPost, 0)
	for _, post := range list {
		if blockedIDs[post.UserID] {
			continue
		}
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

	g.Response = GetConfessionApiResponse{
		TotalCount: len(filteredPosts),
		Posts:      filteredPosts,
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetConfessionApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfGetConfession API执行入口
func hfGetConfession(ctx *gin.Context) {
	api := &GetConfessionApi{}
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
