package repo

import (
	"app/dao/model"
	"app/dao/query"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
	"time"
)

type ConfessionRepo struct {
	query *query.Query
}

func NewConfessionRepo() *ConfessionRepo {
	return &ConfessionRepo{
		query: query.Use(ndb.Pick()),
	}
}

func (r *ConfessionRepo) CreateConfession(ctx context.Context, confession *model.Confession) (err error) {
	db := r.query.Confession
	err = db.WithContext(ctx).
		Select(
			db.Content,
			db.IsAnonymous,
			db.IsVisible,
			db.ImageUrls,
			db.UserID,
			db.Name,
			db.Status,
			db.ScheduleTime,
		).
		Create(confession)
	if err != nil {
		return err
	}

	return nil
}

func (r *ConfessionRepo) UpdateConfession(ctx context.Context, confessionID int64, updates map[string]any) error {
	db := r.query.Confession
	_, err := db.WithContext(ctx).
		Where(db.ID.Eq(confessionID)).
		Updates(updates)
	return err
}

func (r *ConfessionRepo) FindConfessionByID(ctx context.Context, id int64) (confession *model.Confession, err error) {
	db := r.query.Confession
	record, err := db.WithContext(ctx).Where(db.ID.Eq(id), db.Status.Eq(1)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return record, nil
}

func (r *ConfessionRepo) GetAllConfessions(ctx context.Context, pageNum, pageSize int) ([]*model.Confession, int64, error) {
	db := r.query.Confession

	offset := (pageNum - 1) * pageSize

	// 查询分页数据
	list, err := db.WithContext(ctx).
		Where(db.IsVisible.Eq(1), db.Status.Eq(1)).
		Order(db.CreatedAt.Desc()). // 按发布时间倒序
		Limit(pageSize).
		Offset(offset).
		Find()
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	total, err := db.WithContext(ctx).
		Where(db.IsVisible.Eq(1)).
		Count()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *ConfessionRepo) GetMyConfessions(ctx context.Context, pageNum, pageSize int, uid int64) ([]*model.Confession, int64, error) {
	db := r.query.Confession

	offset := (pageNum - 1) * pageSize

	// 查询分页数据
	list, err := db.WithContext(ctx).
		Where(db.UserID.Eq(uid)).
		Order(db.CreatedAt.Desc()). // 按发布时间倒序
		Limit(pageSize).
		Offset(offset).
		Find()
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	total, err := db.WithContext(ctx).
		Where(db.UserID.Eq(uid)).
		Count()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *ConfessionRepo) DeleteConfession(ctx context.Context, id int64) (err error) {
	db := r.query.Confession
	_, err = db.WithContext(ctx).Where(db.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (r *ConfessionRepo) FindDueToPublish(ctx context.Context) ([]*model.Confession, error) {
	list := make([]*model.Confession, 0)
	db := r.query.Confession
	now := time.Now()

	// 查询条件
	list, err := db.WithContext(ctx).
		Where(db.Status.Eq(0), db.ScheduleTime.Lte(now)).
		Find()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ConfessionRepo) PublishDue(ctx context.Context, id int64) error {
	db := r.query.Confession
	_, err := db.WithContext(ctx).Where(db.ID.Eq(id), db.Status.Eq(0)).Update(db.Status, 1)
	if err != nil {
		return err
	}
	return nil
}

func (r *ConfessionRepo) FindByIDs(ctx *gin.Context, ids []int64) ([]*model.Confession, error) {
	db := r.query.Confession
	list := make([]*model.Confession, 0)
	for _, id := range ids {
		confession, err := db.WithContext(ctx).Where(db.ID.Eq(id), db.Status.Eq(1)).First()
		if err != nil {
			return nil, err
		}
		list = append(list, confession)
	}
	return list, nil
}
