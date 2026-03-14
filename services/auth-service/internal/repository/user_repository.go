package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kumparan/cacher"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/bagasss3/toko/packages/database"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

type userRepository struct {
	db          *database.DB
	cacheKeeper cacher.Keeper
	cacheTTL    time.Duration
}

func NewUserRepository(
	db *database.DB,
	cacheKeeper cacher.Keeper,
	cacheTTL time.Duration,
) model.UserRepository {
	return &userRepository{
		db:          db,
		cacheKeeper: cacheKeeper,
		cacheTTL:    cacheTTL,
	}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	key := fmt.Sprintf("user:id:%s", id.String())

	res, mu, err := r.cacheKeeper.GetOrLock(key)
	if err != nil {
		log.WithError(err).Warn("cache get failed, falling back to db")
		return r.findByIDFromDB(ctx, id)
	}

	if res != nil {
		var user model.User
		if err := json.Unmarshal([]byte(res.(string)), &user); err == nil {
			return &user, nil
		}
	}

	defer mu.Unlock()

	user, err := r.findByIDFromDB(ctx, id)
	if err != nil || user == nil {
		return user, err
	}

	r.storeToCache(key, user)
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	key := fmt.Sprintf("user:email:%s", email)

	res, mu, err := r.cacheKeeper.GetOrLock(key)
	if err != nil {
		log.WithError(err).Warn("cache get failed, falling back to db")
		return r.findByEmailFromDB(ctx, email)
	}

	if res != nil {
		var user model.User
		if err := json.Unmarshal([]byte(res.(string)), &user); err == nil {
			return &user, nil
		}
	}

	defer mu.Unlock()

	user, err := r.findByEmailFromDB(ctx, email)
	if err != nil || user == nil {
		return user, err
	}

	r.storeToCache(key, user)
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.Conn.WithContext(ctx).Create(user).Error
}

func (r *userRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, user *model.User) error {
	return tx.WithContext(ctx).Create(user).Error
}

func (r *userRepository) CountAdmins(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.Conn.WithContext(ctx).
		Model(&model.User{}).
		Where("role = ?", model.RoleAdmin).
		Count(&count).Error
	return count, err
}

func (r *userRepository) InvalidateUser(id uuid.UUID, email string) {
	keys := []string{
		fmt.Sprintf("user:id:%s", id.String()),
		fmt.Sprintf("user:email:%s", email),
	}
	if err := r.cacheKeeper.DeleteByKeys(keys); err != nil {
		log.WithError(err).Warn("failed to invalidate user cache")
	}
}

func (r *userRepository) findByIDFromDB(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Conn.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) findByEmailFromDB(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.Conn.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) storeToCache(key string, user *model.User) {
	b, err := json.Marshal(user)
	if err != nil {
		log.WithError(err).Warn("failed to marshal user for cache")
		return
	}
	r.cacheKeeper.StoreWithoutBlocking(
		cacher.NewItemWithCustomTTL(key, string(b), r.cacheTTL),
	)
}
