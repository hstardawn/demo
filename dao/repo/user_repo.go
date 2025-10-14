package repo

import (
	"context"

	"github.com/zjutjh/mygo/ndb"

	"app/dao/model"
	"app/dao/query"
)

type UserRepo struct {
}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	do := query.Use(ndb.Pick()).User
	record, err := do.WithContext(ctx).Where(do.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return record, nil
}
