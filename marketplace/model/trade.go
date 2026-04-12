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

// BuyerStats holds aggregated usage statistics for a buyer
type BuyerStats struct {
	TotalTrades  int64 `json:"total_trades"`
	TotalTokens  int64 `json:"total_tokens"`
	TotalSpent   int64 `json:"total_spent"`   // buyer_amount sum
	TotalSaved   int64 `json:"total_saved"`   // retail_amount - buyer_amount
}

// GetBuyerStats returns aggregated statistics for a buyer
func GetBuyerStats(buyerUserId int) (*BuyerStats, error) {
	var stats BuyerStats
	err := mainModel.DB.Model(&Trade{}).
		Where("buyer_user_id = ?", buyerUserId).
		Select("COUNT(*) as total_trades, COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(SUM(buyer_amount), 0) as total_spent, COALESCE(SUM(retail_amount - buyer_amount), 0) as total_saved").
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// BuyerModelUsage holds per-model usage for a buyer
type BuyerModelUsage struct {
	ModelName   string `json:"model_name"`
	TradeCount  int64  `json:"trade_count"`
	TotalTokens int64  `json:"total_tokens"`
	TotalSpent  int64  `json:"total_spent"`
	AvgDiscount int    `json:"avg_discount"`
}

// GetBuyerModelUsage returns per-model usage breakdown
func GetBuyerModelUsage(buyerUserId int) ([]*BuyerModelUsage, error) {
	var usage []*BuyerModelUsage
	err := mainModel.DB.Model(&Trade{}).
		Where("buyer_user_id = ?", buyerUserId).
		Select("model_name, COUNT(*) as trade_count, SUM(total_tokens) as total_tokens, SUM(buyer_amount) as total_spent, AVG(discount) as avg_discount").
		Group("model_name").
		Order("total_spent desc").
		Scan(&usage).Error
	return usage, err
}
