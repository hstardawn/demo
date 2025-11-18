package repo

import (
	"app/dao/model"
	"app/dao/query"
	"context"
	"errors"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type BlockRepo struct {
	query *query.Query
}

func NewBlockRepo() *BlockRepo {
	return &BlockRepo{
		query: query.Use(ndb.Pick()),
	}
}

func (r *BlockRepo) BlockUser(ctx context.Context, userID, blockedID int64) (err error) {
	db := r.query.Block

	err = db.WithContext(ctx).Create(&model.Block{
		UserID:    userID,
		BlockedID: blockedID,
		Status:    1,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *BlockRepo) UnBlockUser(ctx context.Context, userID, blockedID int64) error {
	db := r.query.Block

	_, err := db.WithContext(ctx).Where(db.UserID.Eq(userID), db.BlockedID.Eq(blockedID)).UpdateSimple(db.Status.Value(0))
	return err
}

func (r *BlockRepo) IsBlocked(ctx context.Context, userID, blockedID int64) (*model.Block, error) {
	db := r.query.Block
	record, err := db.WithContext(ctx).Where(db.UserID.Eq(userID), db.BlockedID.Eq(blockedID), db.Status.Eq(1)).First()
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return record, err

}

func (r *BlockRepo) GetBlockedList(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.Block, int64, error) {
	db := r.query.Block

	offset := (pageNum - 1) * pageSize

	// 查询分页数据
	list, err := db.WithContext(ctx).
		Where(db.UserID.Eq(userID)).
		Order(db.CreatedAt.Desc()). // 按发布时间倒序
		Limit(pageSize).
		Offset(offset).
		Find()
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	total, err := db.WithContext(ctx).
		Where(db.UserID.Eq(userID)).
		Count()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
