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

// GetListHandler API router注册点
func GetListHandler() gin.HandlerFunc {
	api := GetListApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetList).Pointer()).Name()] = api
	return hfGetList
}

type GetListApi struct {
	Info     struct{}           `name:"获取表白" desc:"获取表白"`
	Request  GetListApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetListApiResponse // API响应数据 (Body中的Data部分)
}

type GetListApiRequest struct {
	Query struct {
		PageSize int `form:"page_size" binding:"required" validate:"max=10, min=1" desc:"页容量"`
		PageNum  int `form:"page_num" binding:"required" validate:"max=100, min=1" desc:"当前页码"`
	}
}

type AnswerConfession struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ImageUrl  []string  `json:"image_url"`
	IsBlocked bool      `json:"is_blocked"`
	LikeCount int64     `json:"like_count"` // 总赞数
	IsLiked   bool      `json:"is_liked"`   // 我是否赞过
}
type GetListApiResponse struct {
	TotalCount  int64              `json:"total_count" desc:"帖子数目"`
	Confessions []AnswerConfession `json:"posts" desc:"帖子列表"`
}

// Run Api业务逻辑执行点
func (g *GetListApi) Run(ctx *gin.Context) kit.Code {
	p := repo.NewConfessionRepo()
	b := repo.NewBlockRepo()
	l := repo.NewLikeRepo()
	request := g.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	// 查找用户
	uid := cast.ToInt64(id)
	list, _, err := p.GetAllConfessions(ctx, request.PageNum, request.PageSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取表白列表失败")
		return comm.CodeListError
	}
	confessionIDs := make([]int64, 0)
	for _, v := range list {
		confessionIDs = append(confessionIDs, v.ID)
	}

	// 获取拉黑关系
	blockList, total, err := b.GetBlockedList(ctx, uid, request.PageNum, request.PageSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取拉黑列表失败")
		return comm.CodeDatabaseError
	}
	blockedIDs := make(map[int64]bool)
	for _, blk := range blockList {
		blockedIDs[blk.BlockedID] = true
	}

	// 批量去 Redis 查询点赞信息
	likeMap, err := l.GetBatchLikeInfo(ctx, confessionIDs, uid)
	if err != nil {
		nlog.Pick().Errorf("获取点赞信息失败: %v", err)
		return comm.CodeDatabaseError
	}

	filteredConfessions := make([]AnswerConfession, 0)
	for _, item := range list {
		isBlocked := blockedIDs[uid]
		newConfession := AnswerConfession{
			Id:        item.ID,
			UserId:    item.UserID,
			Name:      item.Name,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			ImageUrl:  strings.Split(item.ImageUrls, ","),
			IsBlocked: isBlocked,
		}
		// 填充点赞信息
		if info, ok := likeMap[item.ID]; ok {
			newConfession.LikeCount = info.LikeCount
			newConfession.IsLiked = info.IsLiked
		} else {
			newConfession.LikeCount = 0
			newConfession.IsLiked = false
		}

		filteredConfessions = append(filteredConfessions, newConfession)
	}

	g.Response = GetListApiResponse{
		TotalCount:  total,
		Confessions: filteredConfessions,
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetListApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfGetList API执行入口
func hfGetList(ctx *gin.Context) {
	api := &GetListApi{}
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
