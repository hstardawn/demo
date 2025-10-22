package repo

import (
	"app/dao/model"
	"app/dao/query"
	"context"
	"github.com/zjutjh/mygo/ndb"
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
