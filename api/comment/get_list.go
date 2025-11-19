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
	Request  GetListApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response GetListApiResponse // API响应数据 (Body中的Data部分)
}

type GetListApiRequest struct {
	Body struct {
		ConfessionID int64 `json:"confession_id" binding:"required" desc:"表白ID"`
		PageNum      int   `json:"page_num" binding:"required" validate:"max=100" desc:"当前页码"`
		PageSize     int   `json:"page_size" binding:"required" validate:"max=10" desc:"页容量"`
	}
}

type GetListApiResponse struct {
	List []*ViewComment `json:"list"` // 评论视图对象列表
}

// ViewComment 是一个“评论视图对象”，用于将数据库模型转换为前端友好的结构
type ViewComment struct {
	ID           int64          `json:"id"`
	ConfessionID int64          `json:"confession_id"`
	ParentID     int64          `json:"parent_id"`
	UserID       int64          `json:"user_id"`
	Content      string         `json:"content"`
	CreatedAt    int64          `json:"created_at"` // 时间戳
	UpdatedAt    int64          `json:"updated_at"`
	User         *UserVO        `json:"user"`     // 评论用户的信息
	Children     []*ViewComment `json:"children"` // 子评论
}

// UserVO 封装了用户基本信息，避免泄露敏感数据
type UserVO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"` // 用户头像
}

// Run Api业务逻辑执行点
func (g *GetListApi) Run(ctx *gin.Context) kit.Code {
	r := repo.NewCommentRepo()
	u := repo.NewUserRepo()
	req := g.Request.Body

	// 鉴权
	_, err := jwt.GetUid(ctx)
	if err != nil {
		return comm.CodeNotLoggedIn
	}

	// 获取顶级评论
	tops, err := r.ListTopLevelByPost(ctx, req.ConfessionID, req.PageNum, req.PageSize)
	println(len(tops))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取顶级评论失败")
		return comm.CodeDatabaseError
	}

	// 如果顶级评论为空，直接返回空列表
	if len(tops) == 0 {
		g.Response.List = []*ViewComment{}
		return comm.CodeOK
	}

	var topIDs []int64
	for _, top := range tops {
		topIDs = append(topIDs, top.ID)
	}

	// 获取子评论
	children, err := r.GetChildrenByParentIDs(ctx, topIDs)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取子评论失败")
		return comm.CodeDatabaseError
	}

	childrenMap := make(map[int64][]*model.Comment)
	for _, ch := range children {
		childrenMap[ch.ParentID] = append(childrenMap[ch.ParentID], ch)
	}

	uniqueUserIDsMap := make(map[int64]struct{})

	// 收集顶级评论的用户ID
	for _, top := range tops {
		if top.UserID != 0 {
			uniqueUserIDsMap[top.UserID] = struct{}{}
		}
	}
	// 收集子评论的用户ID
	for _, ch := range children {
		if ch.UserID != 0 {
			uniqueUserIDsMap[ch.UserID] = struct{}{}
		}
	}

	var finalUserIDs []int64
	for userID := range uniqueUserIDsMap {
		finalUserIDs = append(finalUserIDs, userID)
	}

	if len(finalUserIDs) == 0 {
		g.Response.List = []*ViewComment{}
		nlog.Pick().WithContext(ctx).Warn("No valid user IDs found to fetch, returning empty list.")
		return comm.CodeOK
	}

	// 获取用户信息
	users, err := u.FindByIDs(ctx, finalUserIDs)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取用户列表失败")
		return comm.CodeDatabaseError
	}

	userMap := make(map[int64]*model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	var commentVOs []*ViewComment
	for _, top := range tops {
		topVO := &ViewComment{
			ID:           top.ID,
			ConfessionID: top.ConfessionID,
			ParentID:     top.ParentID,
			UserID:       top.UserID,
			Content:      top.Content,
			CreatedAt:    top.CreatedAt.Unix(),
			UpdatedAt:    top.UpdatedAt.Unix(),
		}

		// 填充用户信息
		if user, ok := userMap[top.UserID]; ok {
			topVO.User = &UserVO{
				ID:     user.ID,
				Name:   user.Name,
				Avatar: user.Avatar,
			}
		}

		// 填充子评论
		if childrenForTop, ok := childrenMap[top.ID]; ok {
			var childVOs []*ViewComment
			for _, child := range childrenForTop {
				childVO := &ViewComment{
					ID:           child.ID,
					ConfessionID: child.ConfessionID,
					ParentID:     child.ParentID,
					UserID:       child.UserID,
					Content:      child.Content,
					CreatedAt:    child.CreatedAt.Unix(),
					UpdatedAt:    child.UpdatedAt.Unix(),
				}
				// 填充子评论的用户信息
				if user, ok := userMap[child.UserID]; ok {
					childVO.User = &UserVO{
						ID:     user.ID,
						Name:   user.Name,
						Avatar: user.Avatar,
					}
				}

				childVOs = append(childVOs, childVO)
			}
			topVO.Children = childVOs
		}
		commentVOs = append(commentVOs, topVO)
	}

	g.Response = GetListApiResponse{
		List: commentVOs,
	}
	nlog.Pick().WithContext(ctx).Infof("Final response list contains %d ViewComment objects.", len(commentVOs))
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (g *GetListApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&g.Request.Body)
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
