package repo

import (
	"app/dao/model"
	"app/dao/query"
	"context"
	"errors"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type CommentRepo struct {
	query *query.Query
}

func NewCommentRepo() *CommentRepo {
	return &CommentRepo{
		query: query.Use(ndb.Pick()),
	}
}

func (r *CommentRepo) CreateConfession(ctx context.Context, comment *model.Comment) (err error) {
	db := r.query.Comment
	err = db.WithContext(ctx).Create(comment)
	if err != nil {
		return err
	}
	return nil
}

// ListTopLevelByPost 获取帖子的顶级评论
func (r *CommentRepo) ListTopLevelByPost(ctx context.Context, confessionID int64, pageNum, pageSize int) ([]*model.Comment, error) {
	db := r.query.Comment

	offset := (pageNum - 1) * pageSize

	var comments []*model.Comment
	comments, err := db.WithContext(ctx).
		Where(db.ConfessionID.Eq(confessionID)).
		Where(db.ParentID.Eq(0)). // <-- 修正：匹配 ParentID 为 NULL 或 0
		Limit(pageSize).
		Offset(offset).
		Find()

	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepo) GetChildrenByParentIDs(ctx context.Context, parentIDs []int64) ([]*model.Comment, error) {
	db := r.query.Comment
	if len(parentIDs) == 0 {
		return nil, nil
	}
	var list []*model.Comment
	list, err := db.WithContext(ctx).
		Where(db.ParentID.In(parentIDs...)).
		Order(db.ParentID, db.CreatedAt.Asc()).
		Find()
	return list, err
}

func (r *CommentRepo) GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	db := r.query.Comment
	record, err := db.WithContext(ctx).Where(db.ID.Eq(commentID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return record, nil
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, id int64) (err error) {
	db := r.query.Comment
	_, err = db.WithContext(ctx).Where(db.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}
