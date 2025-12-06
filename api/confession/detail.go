package confession

import (
	"app/dao/repo"
	"context"
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

// DetailHandler API router注册点
func DetailHandler() gin.HandlerFunc {
	api := DetailApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfDetail).Pointer()).Name()] = api
	return hfDetail
}

type DetailApi struct {
	Info     struct{}          `name:"表白详情" desc:"表白详情"`
	Request  DetailApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response DetailApiResponse // API响应数据 (Body中的Data部分)
}

type DetailApiRequest struct {
	Query struct {
		ConfessionID int `form:"confession_id" binding:"required" desc:"表白ID"`
	}
}

type DetailApiResponse struct {
	Data *ViewConfessionDetail `json:"data"`
}

// ViewConfessionDetail 详情页视图对象
type ViewConfessionDetail struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id"`
	CreatedAt int64  `json:"created_at"`

	// 动态数据
	LikeCount int64 `json:"like_count"` // 点赞数
	ViewCount int64 `json:"view_count"` // 浏览量
	IsLiked   bool  `json:"is_liked"`   // 当前用户是否点赞

	// 关联数据
	User *UserVO `json:"user"` // 作者信息
}

type UserVO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// Run Api业务逻辑执行点
func (d *DetailApi) Run(ctx *gin.Context) kit.Code {
	req := d.Request
	ID := req.Query.ConfessionID
	confessionID := cast.ToInt64(ID)

	// 1. 获取当前登录用户ID (未登录也可以看，uid为0)
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}
	uid := cast.ToInt64(id)

	rConfession := repo.NewConfessionRepo()
	rHot := repo.NewHotRepo()
	rLike := repo.NewLikeRepo()
	rUser := repo.NewUserRepo()

	if err := rHot.IncrView(context.Background(), confessionID); err != nil {
		nlog.Pick().Errorf("增加浏览量/热度失败 ID:%d, err:%v", confessionID, err)
		return comm.CodeDatabaseError
	}
	confession, err := rConfession.FindConfessionByID(ctx, confessionID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if confession == nil {
		return comm.CodeConfessionNotFound
	}

	// 查询 Redis 互动数据
	viewCount, _ := rHot.GetViewCount(ctx, confessionID)
	likeCount, _ := rLike.GetLikeCount(ctx, confessionID)
	isLiked := false
	isLiked, _ = rLike.IsUserLiked(ctx, confessionID, uid)

	// 5. 查询作者信息
	var userVO *UserVO
	if confession.UserID > 0 {
		user, err := rUser.FindByID(ctx, confession.UserID)
		if err == nil && user != nil {
			userVO = &UserVO{
				ID:     user.ID,
				Name:   user.Name,
				Avatar: user.Avatar,
			}
		}
	}

	// 组装返回
	d.Response.Data = &ViewConfessionDetail{
		ID:        confession.ID,
		Content:   confession.Content,
		UserID:    confession.UserID,
		CreatedAt: confession.CreatedAt.Unix(), // 假设DB里是time.Time或int64

		LikeCount: likeCount,
		ViewCount: viewCount,
		IsLiked:   isLiked,

		User: userVO,
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (d *DetailApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&d.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfDetail API执行入口
func hfDetail(ctx *gin.Context) {
	api := &DetailApi{}
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
