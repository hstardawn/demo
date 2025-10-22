package repo

import (
	"app/dao/model"
	"context"
	"errors"
	"gorm.io/gorm"

	"github.com/zjutjh/mygo/ndb"

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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return record, nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	do := query.Use(ndb.Pick()).User
	record, err := do.WithContext(ctx).Where(do.Username.Eq(username)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *UserRepo) CreatUser(ctx context.Context, user *model.User) error {
	do := query.Use(ndb.Pick()).User
	err := do.WithContext(ctx).Create(user)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user *model.User) error {
	do := query.Use(ndb.Pick()).User
	err := do.WithContext(ctx).Save(user)
	if err != nil {
		return err
	}
	return nil
}
