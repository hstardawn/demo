package repo

import (
	"app/dao/model"
	"app/dao/query"
	"context"
	"errors"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type BlockRepo struct{}

func NewBlockRepo() *BlockRepo {
	return &BlockRepo{}
}

func (r *BlockRepo) BlockUser(ctx context.Context, userID, blockedID int64) (err error) {
	db := query.Use(ndb.Pick()).Block
	record, err := db.WithContext(ctx).Where(db.BlockedID.Eq(blockedID), db.UserID.Eq(userID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		block := &model.Block{UserID: userID, BlockedID: blockedID, Status: true}
		err = db.WithContext(ctx).Create(block)
		if err != nil {
			return err
		}
	}
	_, err = db.WithContext(ctx).Where(db.ID.Eq(record.ID)).UpdateSimple(db.Status.Value(true))
	if err != nil {
		return err
	}
	return nil
}

func (r *BlockRepo) UnBlockUser(ctx context.Context, userID, blockedID int64) error {
	db := query.Use(ndb.Pick()).Block
	// 查找拉黑记录
	record, err := db.WithContext(ctx).Where(
		db.BlockedID.Eq(blockedID),
		db.UserID.Eq(userID),
	).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	_, err = db.WithContext(ctx).Where(db.ID.Eq(record.ID)).UpdateSimple(db.Status.Value(false))
	return err
}

func (r *BlockRepo) IsBlocked(ctx context.Context, userID, blockedID int64) (bool, error) {
	var count int64
	db := query.Use(ndb.Pick()).Block
	count, err := db.WithContext(ctx).Where(db.UserID.Eq(userID), db.BlockedID.Eq(blockedID), db.Status.Is(true)).Count()
	return count > 0, err
}

func (r *BlockRepo) GetBlockedList(ctx context.Context, userID int64) ([]*model.Block, error) {
	db := query.Use(ndb.Pick()).Block

	// 查询该用户所有已屏蔽的记录
	list, err := db.WithContext(ctx).
		Where(db.UserID.Eq(userID), db.Status.Is(true)).
		Order(db.CreatedAt.Desc()).
		Find()
	if err != nil {
		return nil, err
	}

	return list, nil
}
