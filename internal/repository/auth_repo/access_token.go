package auth_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/auth"
)

type AccessTokenRepo struct{ DB *gorm.DB }

func NewAccessTokenRepo(db *gorm.DB) *AccessTokenRepo {
	return &AccessTokenRepo{DB: db}
}

func (r *AccessTokenRepo) GetByUsername(username string) (*auth.AccessToken, error) {
	var v auth.AccessToken
	if err := r.DB.Where("username = ?", username).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AccessTokenRepo) GetByAccessToken(token string) (*auth.AccessToken, error) {
	var v auth.AccessToken
	if err := r.DB.Where("access_token = ?", token).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AccessTokenRepo) GetByRefreshToken(token string) (*auth.AccessToken, error) {
	var v auth.AccessToken
	if err := r.DB.Where("refresh_token = ?", token).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AccessTokenRepo) Upsert(v *auth.AccessToken) error {
	return r.DB.
		Where(auth.AccessToken{Username: v.Username}).
		Assign(v).
		FirstOrCreate(v).Error
}

func (r *AccessTokenRepo) Save(v *auth.AccessToken) error {
	return r.DB.Save(v).Error
}

func (r *AccessTokenRepo) ClearByUsername(username string) error {
	return r.DB.Model(&auth.AccessToken{}).
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"access_token":             nil,
			"current_datetime_access":  nil,
			"refresh_token":            nil,
			"current_datetime_refresh": nil,
		}).Error
}

func (r *AccessTokenRepo) Delete(username string) error {
	return r.DB.Where("username = ?", username).Delete(&auth.AccessToken{}).Error
}
