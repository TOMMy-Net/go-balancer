package database

import (
	"context"

	"github.com/TOMMy-Net/go-balancer/internal/domain/models"
	"github.com/jinzhu/copier"
)

func (d *Database) AddClient(ctx context.Context, m *models.Client) error {
	var client Clients
	err := copier.Copy(&client, m)
	if err != nil {
		return err
	}
	err = d.db.WithContext(ctx).Create(&client).Error
	return err
}
