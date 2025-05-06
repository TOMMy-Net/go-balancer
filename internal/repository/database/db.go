package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

type ConfigDB struct {
	Host     string
	User     string
	Password string
	NameDB   string
	Port     string
	SSL      string
}

func NewPostgres(c *ConfigDB) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.NameDB, c.SSL,
	)
	return connect(postgres.Open(dsn))
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
