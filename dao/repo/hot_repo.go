package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
	"strconv"
)

const (
	KeyHotRank    = "rank:confession:hot"
	KeyPrefixView = "view:confession:"
	ScorePerLike  = 3.0 // 点赞权重
	ScorePerView  = 2.0 // 浏览权重
)

type HotRepo struct {
	redis redis.UniversalClient
}

func NewHotRepo() *HotRepo {
	return &HotRepo{redis: nedis.Pick()}
}

// IncrView 增加浏览量并更新热度
func (r *HotRepo) IncrView(ctx context.Context, confessionID int64) error {
	rd := r.redis
	pipe := rd.Pipeline()

	// 1. 浏览量计数器 +1 (用于详情页展示)
	viewKey := fmt.Sprintf("%s%d", KeyPrefixView, confessionID)
	pipe.Incr(ctx, viewKey)

	// 2. 热度榜单分数 +2
	pipe.ZIncrBy(ctx, KeyHotRank, ScorePerView, strconv.FormatInt(confessionID, 10))

	_, err := pipe.Exec(ctx)
	return err
}

// UpdateLikeScore 点赞/取消赞 引起的热度变化
// action: 1=点赞, 2=取消
func (r *HotRepo) UpdateLikeScore(ctx context.Context, confessionID int64, action int) error {
	rd := r.redis
	delta := ScorePerLike
	if action == 0 {
		delta = -ScorePerLike // 取消赞，扣分
	}
	// ZINCRBY 支持负数，自动处理加减
	return rd.ZIncrBy(ctx, KeyHotRank, delta, strconv.FormatInt(confessionID, 10)).Err()
}

// GetHotIDs 获取热度榜单前N名的ID
func (r *HotRepo) GetHotIDs(ctx context.Context, page, size int) ([]int64, error) {
	rd := r.redis
	start := int64((page - 1) * size)
	stop := start + int64(size) - 1

	// ZREVRANGE: 按分数从大到小排序
	result, err := rd.ZRevRange(ctx, KeyHotRank, start, stop).Result()
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, s := range result {
		id, _ := strconv.ParseInt(s, 10, 64)
		ids = append(ids, id)
	}
	return ids, nil
}

// GetBatchViewCounts 批量获取浏览量 (用于列表展示)
func (r *HotRepo) GetBatchViewCounts(ctx context.Context, ids []int64) (map[int64]int64, error) {
	rd := r.redis
	pipe := rd.Pipeline()
	cmds := make(map[int64]*redis.StringCmd)

	for _, id := range ids {
		key := fmt.Sprintf("%s%d", KeyPrefixView, id)
		cmds[id] = pipe.Get(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[int64]int64)
	for id, cmd := range cmds {
		val, _ := cmd.Int64() // 如果key不存在默认返回0
		res[id] = val
	}
	return res, nil
}

// GetViewCount 获取单个浏览量
func (r *HotRepo) GetViewCount(ctx context.Context, id int64) (int64, error) {
	rd := r.redis
	key := fmt.Sprintf("view:confession:%d", id)
	val, err := rd.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return val, err
}
