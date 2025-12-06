package repo

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

const KeyPrefixLike = "like:confession:"

type LikeRepo struct {
	redis redis.UniversalClient
}

func NewLikeRepo() *LikeRepo {
	return &LikeRepo{redis: nedis.Pick()}
}

func (r *LikeRepo) LikeAction(ctx context.Context, confessionID, userID int64, action int) error {
	rd := r.redis
	key := fmt.Sprintf("%s%d", KeyPrefixLike, confessionID)
	if action == 1 {
		err := rd.SAdd(ctx, key, userID).Err()
		if err != nil {
			return err
		}
	} else {
		err := rd.SRem(ctx, key, userID).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

type LikeInfo struct {
	LikeCount int64
	IsLiked   bool
}

func (r *LikeRepo) GetBatchLikeInfo(ctx context.Context, confessionIDs []int64, currentUserID int64) (map[int64]*LikeInfo, error) {
	rd := r.redis
	pipe := rd.Pipeline()
	resultMap := make(map[int64]*LikeInfo)

	// 定义接收结果的 Future 对象
	countCmds := make(map[int64]*redis.IntCmd)
	isLikedCmds := make(map[int64]*redis.BoolCmd)

	for _, id := range confessionIDs {
		key := fmt.Sprintf("%s%d", KeyPrefixLike, id)
		// 1. 获取点赞总数
		countCmds[id] = pipe.SCard(ctx, key)

		// 2. 如果用户已登录，检查是否点过赞
		if currentUserID > 0 {
			isLikedCmds[id] = pipe.SIsMember(ctx, key, currentUserID)
		}
	}

	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// 解析结果
	for _, id := range confessionIDs {
		info := &LikeInfo{LikeCount: 0, IsLiked: false}

		// 读取数量
		if cmd, ok := countCmds[id]; ok {
			info.LikeCount = cmd.Val()
		}

		// 读取状态
		if cmd, ok := isLikedCmds[id]; ok {
			info.IsLiked = cmd.Val()
		}

		resultMap[id] = info
	}

	return resultMap, nil
}
