package comment

import (
	"app/dao/model"
	"app/dao/repo"
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
	Info     struct{}           `name:"获取评论列表" desc:"获取评论列表"`
	Request  GetListApiRequest  // API请求参数
	Response GetListApiResponse // API响应数据
}

type GetListApiRequest struct {
	Body struct {
		ConfessionID int64 `json:"confession_id" binding:"required" desc:"表白ID"`
		PageNum      int   `json:"page_num" binding:"required" validate:"max=100" desc:"当前页码"`
		PageSize     int   `json:"page_size" binding:"required" validate:"max=10" desc:"页容量"`
	}
}

type GetListApiResponse struct {
	List []*ViewComment `json:"list"` // 扁平化的评论列表
}

// ViewComment 评论视图对象 - 已移除 Children 嵌套结构
type ViewComment struct {
	ID           int64   `json:"id"`
	ConfessionID int64   `json:"confession_id"`
	ParentID     int64   `json:"parent_id"` // 如果是0代表是一级评论，非0代表是子评论
	UserID       int64   `json:"user_id"`
	Content      string  `json:"content"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
	User         *UserVO `json:"user"` // 评论用户的信息
}

type UserVO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// Run Api业务逻辑执行点
func (g *GetListApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewCommentRepo()
	u := repo.NewUserRepo()
	req := g.Request.Body

	// 1. 鉴权
	_, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	// 2. 获取顶级评论（按分页）
	tops, err := r.ListTopLevelByPost(ctx, req.ConfessionID, req.PageNum, req.PageSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取顶级评论失败")
		return comm.CodeDatabaseError
	}
	if len(tops) == 0 {
		g.Response.List = []*ViewComment{}
		return comm.CodeOK
	}

	// 3. 收集顶级评论ID，批量获取子评论
	var topIDs []int64
	for _, top := range tops {
		topIDs = append(topIDs, top.ID)
	}

	children, err := r.GetChildrenByParentIDs(ctx, topIDs)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取子评论失败")
		return comm.CodeDatabaseError
	}

	// 4. 将子评论按 ParentID 分组，方便后续组装
	childrenMap := make(map[int64][]*model.Comment)
	for _, ch := range children {
		childrenMap[ch.ParentID] = append(childrenMap[ch.ParentID], ch)
	}

	// 5. 收集所有涉及到的 UserID (包括父评论和子评论)
	uniqueUserIDsMap := make(map[int64]struct{})
	for _, top := range tops {
		if top.UserID != 0 {
			uniqueUserIDsMap[top.UserID] = struct{}{}
		}
	}
	for _, ch := range children {
		if ch.UserID != 0 {
			uniqueUserIDsMap[ch.UserID] = struct{}{}
		}
	}

	var finalUserIDs []int64
	for userID := range uniqueUserIDsMap {
		finalUserIDs = append(finalUserIDs, userID)
	}

	// 6. 批量获取用户信息
	userMap := make(map[int64]*model.User)
	if len(finalUserIDs) > 0 {
		users, err := u.FindByIDs(ctx, finalUserIDs)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("获取用户列表失败")
			return comm.CodeDatabaseError
		}
		for _, user := range users {
			userMap[user.ID] = user
		}
	}

	var flatComments []*ViewComment

	// 辅助函数：将 model 转换为 view object
	convertToVO := func(m *model.Comment) *ViewComment {
		vo := &ViewComment{
			ID:           m.ID,
			ConfessionID: m.ConfessionID,
			ParentID:     m.ParentID,
			UserID:       m.UserID,
			Content:      m.Content,
			CreatedAt:    m.CreatedAt.Unix(),
			UpdatedAt:    m.UpdatedAt.Unix(),
		}
		if user, ok := userMap[m.UserID]; ok {
			vo.User = &UserVO{
				ID:     user.ID,
				Name:   user.Name,
				Avatar: user.Avatar,
			}
		}
		return vo
	}

	for _, top := range tops {
		// 7.1 加入顶级评论
		flatComments = append(flatComments, convertToVO(top))

		// 7.2 如果该顶级评论有子评论，紧随其后加入
		if childrenForTop, ok := childrenMap[top.ID]; ok {
			for _, child := range childrenForTop {
				flatComments = append(flatComments, convertToVO(child))
			}
		}
	}

	g.Response = GetListApiResponse{
		List: flatComments,
	}

	return comm.CodeOK
}

// Init Api初始化
func (g *GetListApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&g.Request.Body)
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
