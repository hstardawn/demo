package cron

import (
	"app/dao/repo"
	"context"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/nlog"
)

type ScheduleSendJob struct{}

func (ScheduleSendJob) Run() {
	ctx := context.Background()
	r := repo.NewConfessionRepo()

	// 1. 查询到期的任务
	list, err := r.FindDueToPublish(ctx)
	if err != nil {
		nlog.Pick().Errorf("Cron扫描数据库失败: %v", err)
		return
	}

	if len(list) == 0 {
		return // 没有需要发布的
	}

	nlog.Pick().Infof("Cron扫描到 %d 条待发布表白", len(list))

	// 2. 遍历并发布
	for _, item := range list {
		err := r.PublishDue(ctx, item.ID)
		if err != nil {
			nlog.Pick().Errorf("表白发布失败 ID: %d, err: %v", item.ID, err)
			continue
		}

		nlog.Pick().Infof("表白 ID: %d 已自动发布", item.ID)
	}
	nlog.Pick().WithField("app", config.AppName()).Debug("定时任务运行")
}
