package comm

import "github.com/zjutjh/mygo/kit"

var CodeOK = kit.NewCode(0, "成功")

// 系统错误码
var (
	CodeUnknownError           = kit.NewCode(10000, "未知错误")
	CodeThirdServiceError      = kit.NewCode(10001, "三方服务错误")
	CodeDatabaseError          = kit.NewCode(10002, "数据库错误")
	CodeRedisError             = kit.NewCode(10003, "Redis错误")
	CodeMiddlewareServiceError = kit.NewCode(10004, "中间件服务错误")
)

// 业务通用错误码
var (
	CodeNotLoggedIn        = kit.NewCode(20000, "用户未登录")
	CodeLoginExpired       = kit.NewCode(20001, "登录过期，请重新登录")
	CodePermissionDenied   = kit.NewCode(20002, "用户无权限")
	CodeParameterInvalid   = kit.NewCode(20003, "参数非法")
	CodeDataParseError     = kit.NewCode(20004, "数据解析异常")
	CodeDataNotFound       = kit.NewCode(20005, "数据不存在")
	CodeDataConflict       = kit.NewCode(20006, "数据冲突")
	CodeServiceMaintenance = kit.NewCode(20007, "系统维护中")
	CodeTooFrequently      = kit.NewCode(20008, "操作过于频繁/未获得锁")
)

// 业务错误码 从 30000 开始
var (
	CodeUserExisted        = kit.NewCode(30000, "用户已存在")
	CodeSaveError          = kit.NewCode(30001, "保存失败")
	CodeHashError          = kit.NewCode(30002, "加密失败")
	CodeUserNotFound       = kit.NewCode(30003, "用户不存在")
	CodePasswordError      = kit.NewCode(30004, "密码错误")
	CodeOutOfLimited       = kit.NewCode(30005, "图片数量超出限制")
	CodeBlockExisted       = kit.NewCode(30009, "已拉黑")
	CodeBlockNotExisted    = kit.NewCode(30009, "未拉黑")
	CodeUnblockError       = kit.NewCode(30010, "解除拉黑失败")
	CodeSearchError        = kit.NewCode(30011, "查询拉黑关系失败")
	CodeBlockError         = kit.NewCode(30012, "拉黑失败")
	CodeListError          = kit.NewCode(30013, "获取列表失败")
	CodeConfessionNotFound = kit.NewCode(30014, "帖子不存在")
	CodeCommentNotFound    = kit.NewCode(30015, "评论不存在")
	CodeRepeatAction       = kit.NewCode(30016, "重复无效操作")
)
