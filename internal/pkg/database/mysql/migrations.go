package mysql

import (
	"QA-System/internal/model"
	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Survey{},
		&model.Question{},
		&model.Option{},
		&model.Manage{},
		&model.Pre{},
	)
}
