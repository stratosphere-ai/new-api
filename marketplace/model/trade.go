package model

import (
	"time"

	mainModel "github.com/QuantumNous/new-api/model"
)

type Trade struct {
	Id               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ListingId        int       `json:"listing_id" gorm:"index"`
	BuyerUserId      int       `json:"buyer_user_id" gorm:"index"`
	SellerId         int       `json:"seller_id" gorm:"index"`
	ChannelId        int       `json:"channel_id"`
	ModelName        string    `json:"model_name"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	Discount         int       `json:"discount"`                    // seller discount at time of trade
	RetailAmount     int64     `json:"retail_amount"`               // official price (quota units)
	BuyerAmount      int64     `json:"buyer_amount"`                // what buyer paid
	SellerAmount     int64     `json:"seller_amount"`               // what seller earns
	PlatformAmount   int64     `json:"platform_amount"`             // platform commission
	CreatedAt        time.Time `json:"created_at" gorm:"index"`
}

func (Trade) TableName() string {
	return "mp_trades"
}

func CreateTrade(trade *Trade) error {
	return mainModel.DB.Create(trade).Error
}

func GetTradesBySellerId(sellerId int, startTime, endTime time.Time) ([]*Trade, error) {
	var trades []*Trade
	err := mainModel.DB.Where("seller_id = ? AND created_at >= ? AND created_at < ?", sellerId, startTime, endTime).
		Order("created_at desc").Find(&trades).Error
	return trades, err
}

func GetTradesByBuyerUserId(buyerUserId int, limit int, offset int) ([]*Trade, error) {
	var trades []*Trade
	err := mainModel.DB.Where("buyer_user_id = ?", buyerUserId).
		Order("created_at desc").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, err
}

// GetSellerEarningsForPeriod sums seller_amount for a seller in a given period
func GetSellerEarningsForPeriod(sellerId int, startTime, endTime time.Time) (int64, error) {
	var total int64
	err := mainModel.DB.Model(&Trade{}).
		Where("seller_id = ? AND created_at >= ? AND created_at < ?", sellerId, startTime, endTime).
		Select("COALESCE(SUM(seller_amount), 0)").Scan(&total).Error
	return total, err
}
