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

// GetMyListHandler API router注册点
func GetMyListHandler() gin.HandlerFunc {
	api := GetMyListApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetMyList).Pointer()).Name()] = api
	return hfGetMyList
}

type GetMyListApi struct {
	Info     struct{}             `name:"获取个人帖子" desc:"获取个人帖子"`
	Request  GetMyListApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetMyListApiResponse // API响应数据 (Body中的Data部分)
}

type GetMyListApiRequest struct {
	Query struct {
		PageSize int `form:"page_size" binding:"required" validate:"max=10, min=1" desc:"页容量"`
		PageNum  int `form:"page_num" binding:"required" validate:"max=100, min=1" desc:"当前页码"`
	}
}

type MyConfession struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ImageUrl  []string  `json:"image_url"`
	LikeCount int64     `json:"like_count"` // 总赞数
	ViewCount int64     `json:"view_count"`
	IsLiked   bool      `json:"is_liked"` // 我是否赞过
}
type GetMyListApiResponse struct {
	TotalCount int64          `json:"total_count" desc:"帖子数目"`
	Posts      []MyConfession `json:"posts" desc:"帖子列表"`
}

// Run Api业务逻辑执行点
func (g *GetMyListApi) Run(ctx *gin.Context) kit.Code {
	p := repo.NewConfessionRepo()
	l := repo.NewLikeRepo()
	request := g.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)

	// 获取用户帖子
	list, total, err := p.GetMyConfessions(ctx, request.PageNum, request.PageSize, uid)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取表白列表失败")
		return comm.CodeDatabaseError
	}
	confessionIDs := make([]int64, 0)
	for _, v := range list {
		confessionIDs = append(confessionIDs, v.ID)
	}

	likeMap, err := l.GetBatchLikeInfo(ctx, confessionIDs, uid)
	if err != nil {
		nlog.Pick().Errorf("获取点赞信息失败: %v", err)
		return comm.CodeDatabaseError
	}

	filteredPosts := make([]MyConfession, 0)
	for _, item := range list {
		newConfession := MyConfession{
			Id:        item.ID,
			UserId:    item.UserID,
			Name:      item.Name,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			ImageUrl:  strings.Split(item.ImageUrls, ","),
		}
		// 填充点赞信息
		if info, ok := likeMap[item.ID]; ok {
			newConfession.LikeCount = info.LikeCount
			newConfession.IsLiked = info.IsLiked
		} else {
			newConfession.LikeCount = 0
			newConfession.IsLiked = false
		}

		filteredPosts = append(filteredPosts, newConfession)
	}

	g.Response = GetMyListApiResponse{
		TotalCount: total,
		Posts:      filteredPosts,
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetMyListApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfGetMyList API执行入口
func hfGetMyList(ctx *gin.Context) {
	api := &GetMyListApi{}
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
