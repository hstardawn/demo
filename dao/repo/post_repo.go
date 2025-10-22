package repo

import (
	"app/dao/model"
	"app/dao/query"
	"context"
	"errors"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type PostRepo struct{}

func NewPostRepo() *PostRepo {
	return &PostRepo{}
}

func (r *PostRepo) CreatePost(ctx context.Context, post *model.Post) (err error) {
	db := query.Use(ndb.Pick()).Post
	err = db.WithContext(ctx).Create(post)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepo) UpdatePost(ctx context.Context, post *model.Post) (err error) {
	db := query.Use(ndb.Pick()).Post
	err = db.WithContext(ctx).Save(post)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepo) FindPostByID(ctx context.Context, id int64) (post *model.Post, err error) {
	db := query.Use(ndb.Pick()).Post
	record, err := db.WithContext(ctx).Where(db.ID.Eq(id)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return record, nil
}

func (r *PostRepo) GetAllPosts(ctx context.Context, pageNum, pageSize int) ([]*model.Post, int64, error) {
	db := query.Use(ndb.Pick()).Post

	offset := (pageNum - 1) * pageSize

	// 查询分页数据
	list, err := db.WithContext(ctx).
		Where(db.IsVisible.Is(true)).
		Order(db.CreatedAt.Desc()). // 按发布时间倒序
		Limit(pageSize).
		Offset(offset).
		Find()
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	total, err := db.WithContext(ctx).
		Where(db.IsVisible.Is(true)).
		Count()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *PostRepo) GetMyPosts(ctx context.Context, pageNum, pageSize int, uid int64) ([]*model.Post, int64, error) {
	db := query.Use(ndb.Pick()).Post

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

func (r *PostRepo) DeletePost(ctx context.Context, id int64) (err error) {
	db := query.Use(ndb.Pick()).Post
	_, err = db.WithContext(ctx).Where(db.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}
