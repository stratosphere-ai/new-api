package model

import (
	mainModel "github.com/QuantumNous/new-api/model"
)

// MigrateMarketplaceTables creates marketplace tables via GORM AutoMigrate
func MigrateMarketplaceTables() error {
	return mainModel.DB.AutoMigrate(
		&Seller{},
		&Listing{},
		&Trade{},
		&Withdrawal{},
	)
}
