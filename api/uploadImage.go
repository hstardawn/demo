package api

import (
	"mime/multipart"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// UploadImageHandler API router注册点
func UploadImageHandler() gin.HandlerFunc {
	api := &UploadImageApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUploadImage).Pointer()).Name()] = api
	return hfUploadImage
}

type UploadImageApi struct {
	Info     struct{}               `name:"上传图片" desc:"用户上传图片接口"`
	Request  UploadImageApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response UploadImageApiResponse // API响应数据 (Body中的Data部分)
}

// UploadImageApiRequest 请求参数
type UploadImageApiRequest struct {
	Body struct {
		File *multipart.FileHeader `form:"file" binding:"required"` // multipart/form-data 表单字段 file
	}
}

// UploadImageApiResponse 响应数据
type UploadImageApiResponse struct {
	URL string `json:"url"` // 上传后的图片URL
}

// Run Api业务逻辑执行点
func (u *UploadImageApi) Run(ctx *gin.Context) kit.Code {
	file := u.Request.Body.File
	if file == nil {
		return comm.CodeParameterInvalid
	}

	uploadedURL := "/static/uploads/" + file.Filename

	u.Response = UploadImageApiResponse{
		URL: uploadedURL,
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (u *UploadImageApi) Init(ctx *gin.Context) (err error) {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	u.Request.Body.File = file
	return nil
}

// hfUploadImage API执行入口
func hfUploadImage(ctx *gin.Context) {
	api := &UploadImageApi{}
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
