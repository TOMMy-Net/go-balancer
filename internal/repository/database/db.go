package database


import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewPostgres(url string) (*Database, error) {

	return connect(mysql.Open(url))
}

func connect(d gorm.Dialector) (*Database, error) {
	db, err := gorm.Open(d, &gorm.Config{})
	if err != nil {
		return &Database{}, err
	}

	migrate(db)
	return &Database{
		db: db,
	}, nil
}

func migrate(g *gorm.DB) error {
	return g.AutoMigrate(Clients{})
}
