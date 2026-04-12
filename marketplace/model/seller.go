package model

import (
	"time"

	mainModel "github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type Seller struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int       `json:"user_id" gorm:"uniqueIndex"`
	Status      int       `json:"status" gorm:"default:1"`
	Balance     int64     `json:"balance" gorm:"default:0"`       // pending withdrawal balance (in quota units)
	TotalEarned int64     `json:"total_earned" gorm:"default:0"`  // cumulative earnings
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Seller) TableName() string {
	return "mp_sellers"
}

func CreateSeller(seller *Seller) error {
	return mainModel.DB.Create(seller).Error
}

func GetSellerByUserId(userId int) (*Seller, error) {
	var seller Seller
	err := mainModel.DB.Where("user_id = ?", userId).First(&seller).Error
	if err != nil {
		return nil, err
	}
	return &seller, nil
}

func GetSellerById(id int) (*Seller, error) {
	var seller Seller
	err := mainModel.DB.First(&seller, id).Error
	if err != nil {
		return nil, err
	}
	return &seller, nil
}

func IncrementSellerBalance(sellerId int, amount int64) error {
	return mainModel.DB.Model(&Seller{}).Where("id = ?", sellerId).
		UpdateColumns(map[string]interface{}{
			"balance":      gorm.Expr("balance + ?", amount),
			"total_earned": gorm.Expr("total_earned + ?", amount),
		}).Error
}

// GetSellersWithBalance returns all sellers with positive balance
func GetSellersWithBalance() ([]*Seller, error) {
	var sellers []*Seller
	err := mainModel.DB.Where("balance > 0 AND status = ?", SellerStatusActive).Find(&sellers).Error
	return sellers, err
}

// ResetSellerBalance atomically sets balance to 0 and returns the old balance
func ResetSellerBalance(sellerId int) (int64, error) {
	var seller Seller
	err := mainModel.DB.First(&seller, sellerId).Error
	if err != nil {
		return 0, err
	}
	oldBalance := seller.Balance
	if oldBalance <= 0 {
		return 0, nil
	}
	err = mainModel.DB.Model(&Seller{}).Where("id = ? AND balance = ?", sellerId, oldBalance).
		Update("balance", 0).Error
	if err != nil {
		return 0, err
	}
	return oldBalance, nil
}

// GetAllSellers returns all sellers (for admin)
func GetAllSellers() ([]*Seller, error) {
	var sellers []*Seller
	err := mainModel.DB.Order("created_at desc").Find(&sellers).Error
	return sellers, err
}

// UpdateSellerStatus updates seller status (active/suspended)
func UpdateSellerStatus(id int, status int) error {
	return mainModel.DB.Model(&Seller{}).Where("id = ?", id).Update("status", status).Error
}
