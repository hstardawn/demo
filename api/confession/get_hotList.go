package confession

import (
	"app/dao/model"
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

// GetHotListHandler API router注册点
func GetHotListHandler() gin.HandlerFunc {
	api := GetHotListApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetHotList).Pointer()).Name()] = api
	return hfGetHotList
}

type GetHotListApi struct {
	Info     struct{}              `name:"获取热度榜单" desc:"获取热度榜单"`
	Request  GetHotListApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetHotListApiResponse // API响应数据 (Body中的Data部分)
}

type GetHotListApiRequest struct {
	Query struct{}
}

type GetHotListApiResponse struct {
	List []*AnswerConfession `json:"list"`
}

// Run Api业务逻辑执行点
func (g *GetHotListApi) Run(ctx *gin.Context) kit.Code {
	rHot := repo.NewHotRepo()
	rDB := repo.NewConfessionRepo()
	rLike := repo.NewLikeRepo()
	b := repo.NewBlockRepo()

	id, _ := jwt.GetUid(ctx)
	uid := cast.ToInt64(id)

	// 1. 从 Redis ZSet 获取热度前 10 名的 ID
	ids, err := rHot.GetHotIDs(ctx, 1, 10) // 假设取Top10
	if err != nil {
		return comm.CodeDatabaseError
	}

	if len(ids) == 0 {
		g.Response.List = []*AnswerConfession{}
		return comm.CodeOK
	}

	// 2. 批量从 MySQL 查详情 (注意：MySQL查询结果是无序的)
	confessions, err := rDB.FindByIDs(ctx, ids)
	if err != nil {
		return comm.CodeDatabaseError
	}

	// 建立 ID -> Model 的映射，方便后续按顺序提取
	confMap := make(map[int64]*model.Confession)
	for _, c := range confessions {
		confMap[c.ID] = c
	}

	// 批量获取点赞数、浏览数、是否点赞
	likeInfos, _ := rLike.GetBatchLikeInfo(ctx, ids, uid)
	viewCounts, _ := rHot.GetBatchViewCounts(ctx, ids)

	uidSet := make(map[int64]struct{})
	for _, id := range ids {
		if item, ok := confMap[id]; ok {
			uidSet[item.UserID] = struct{}{}
		}
	}

	var blockedMap map[int64]bool
	if len(uidSet) > 0 {
		userIDs := make([]int64, 0, len(uidSet))
		for uid := range uidSet {
			userIDs = append(userIDs, uid)
		}
		blockedMap, err = b.GetBlockStatusBatch(ctx, userIDs, uid)
		if err != nil {
			return comm.CodeDatabaseError
		}
	}

	// 按 Redis 返回的 ids 顺序组装数据
	var viewList []*AnswerConfession
	for _, id := range ids {
		item, ok := confMap[id]
		if !ok {
			continue
		}
		isBlocked := blockedMap[item.UserID]
		vo := &AnswerConfession{
			Id:        item.ID,
			UserId:    item.UserID,
			Name:      item.Name,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			ImageUrl:  strings.Split(item.ImageUrls, ","),
			IsBlocked: isBlocked,
		}

		// 填充 Redis 数据
		if info, ok := likeInfos[id]; ok {
			vo.LikeCount = info.LikeCount
			vo.IsLiked = info.IsLiked
		}
		if views, ok := viewCounts[id]; ok {
			vo.ViewCount = views
		}

		viewList = append(viewList, vo)
	}

	g.Response.List = viewList
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetHotListApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfGetHotList API执行入口
func hfGetHotList(ctx *gin.Context) {
	api := &GetHotListApi{}
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
