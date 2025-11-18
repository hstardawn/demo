package block

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

// GetListHandler API router注册点
func GetListHandler() gin.HandlerFunc {
	api := GetListApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfGetList).Pointer()).Name()] = api
	return hfGetList
}

type GetListApi struct {
	Info     struct{}           `name:"查看拉黑名单" desc:"查看拉黑名单"`
	Request  GetListApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetListApiResponse // API响应数据 (Body中的Data部分)
}

type GetListApiRequest struct {
	Query struct {
		PageNum  int `form:"page_num" validate:"max=100, min=1" desc:"当前页码"`
		PageSize int `form:"page_size" validate:"max=10, min=1" desc:"页容量"`
	}
}

type List struct {
	BlockedID int64 `json:"blocked_id"`
}
type GetListApiResponse struct {
	Total int    `json:"total" desc:"列表长度"`
	List  []List `json:"list" desc:"拉黑列表"`
}

// Run Api业务逻辑执行点
func (g *GetListApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewBlockRepo()
	request := g.Request.Query
	id, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	uid := cast.ToInt64(id)
	list, _, err := r.GetBlockedList(ctx, uid, request.PageNum, request.PageSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取拉黑列表失败")
		return comm.CodeDatabaseError
	}

	respList := make([]List, 0, len(list))
	for _, v := range list {
		if v.Status == 0 {
			continue
		}
		respList = append(respList, List{
			BlockedID: v.BlockedID,
		})
	}
	g.Response = GetListApiResponse{
		Total: len(respList),
		List:  respList,
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
