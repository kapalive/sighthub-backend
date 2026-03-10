package assessment

import (
	"errors"

	"gorm.io/gorm"
	a "sighthub-backend/internal/models/medical/vision_exam/assessment"
)

type MyTopDiseaseRepo struct{ DB *gorm.DB }

func NewMyTopDiseaseRepo(db *gorm.DB) *MyTopDiseaseRepo {
	return &MyTopDiseaseRepo{DB: db}
}

func (r *MyTopDiseaseRepo) GetByID(id int64) (*a.MyTopDisease, error) {
	var v a.MyTopDisease
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *MyTopDiseaseRepo) GetAll() ([]a.MyTopDisease, error) {
	var items []a.MyTopDisease
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MyTopDiseaseRepo) Create(v *a.MyTopDisease) error {
	return r.DB.Create(v).Error
}

func (r *MyTopDiseaseRepo) Save(v *a.MyTopDisease) error {
	return r.DB.Save(v).Error
}

func (r *MyTopDiseaseRepo) Delete(id int64) error {
	return r.DB.Delete(&a.MyTopDisease{}, id).Error
}
