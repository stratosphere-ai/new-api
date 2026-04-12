package model

import (
	"time"

	mainModel "github.com/QuantumNous/new-api/model"
)

type Withdrawal struct {
	Id          int        `json:"id" gorm:"primaryKey;autoIncrement"`
	SellerId    int        `json:"seller_id" gorm:"index"`
	Amount      int64      `json:"amount"`                          // withdrawal amount (quota units)
	Status      int        `json:"status" gorm:"default:1"`         // pending, approved, paid, rejected
	PeriodStart time.Time  `json:"period_start"`
	PeriodEnd   time.Time  `json:"period_end"`
	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `json:"paid_at"`
}

func (Withdrawal) TableName() string {
	return "mp_withdrawals"
}

func CreateWithdrawal(w *Withdrawal) error {
	return mainModel.DB.Create(w).Error
}

func GetWithdrawalsBySellerId(sellerId int) ([]*Withdrawal, error) {
	var withdrawals []*Withdrawal
	err := mainModel.DB.Where("seller_id = ?", sellerId).Order("created_at desc").Find(&withdrawals).Error
	return withdrawals, err
}

func UpdateWithdrawalStatus(id int, status int) error {
	updates := map[string]interface{}{"status": status}
	if status == WithdrawalStatusPaid {
		now := time.Now()
		updates["paid_at"] = &now
	}
	return mainModel.DB.Model(&Withdrawal{}).Where("id = ?", id).Updates(updates).Error
}

func GetPendingWithdrawals() ([]*Withdrawal, error) {
	var withdrawals []*Withdrawal
	err := mainModel.DB.Where("status = ?", WithdrawalStatusPending).Order("created_at asc").Find(&withdrawals).Error
	return withdrawals, err
}

// GetAllWithdrawals returns all withdrawals optionally filtered by status
func GetAllWithdrawals(status int) ([]*Withdrawal, error) {
	var withdrawals []*Withdrawal
	q := mainModel.DB.Order("created_at desc")
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&withdrawals).Error
	return withdrawals, err
}

// GetWithdrawalById returns a single withdrawal
func GetWithdrawalById(id int) (*Withdrawal, error) {
	var w Withdrawal
	err := mainModel.DB.First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}
