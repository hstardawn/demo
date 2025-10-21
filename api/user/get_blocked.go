package user

import (
	"app/dao/repo"
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

// GetBlockedHandler API router注册点
func GetBlockedHandler() gin.HandlerFunc {
	api := GetBlockedApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetBlocked).Pointer()).Name()] = api
	return hfGetBlocked
}

type GetBlockedApi struct {
	Info     struct{}              `name:"查看拉黑名单" desc:"查看拉黑名单"`
	Request  GetBlockedApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetBlockedApiResponse // API响应数据 (Body中的Data部分)
}

type GetBlockedApiRequest struct {
	Query struct {
		PageSize int `form:"page_size" binding:"required" desc:"最大容量"`
		PageNum  int `form:"page_num" binding:"required" desc:"页码数目"`
	}
}

type BlockList struct {
	BlockedID int64 `json:"blocked_id"`
}
type GetBlockedApiResponse struct {
	Total int64       `json:"total" desc:"总拉黑人数"`
	List  []BlockList `json:"list" desc:"拉黑列表"`
}

// Run Api业务逻辑执行点
func (g *GetBlockedApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewBlockRepo()
	request := g.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	list, total, err := r.GetBlockedList(ctx, uid, request.PageNum, request.PageSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取拉黑列表失败")
		return comm.CodeListError
	}

	respList := make([]BlockList, 0, len(list))
	for _, v := range list {
		respList = append(respList, BlockList{
			BlockedID: v.BlockedID,
		})
	}
	g.Response = GetBlockedApiResponse{
		Total: total,
		List:  respList,
	}
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetBlockedApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfGetBlocked API执行入口
func hfGetBlocked(ctx *gin.Context) {
	api := &GetBlockedApi{}
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
