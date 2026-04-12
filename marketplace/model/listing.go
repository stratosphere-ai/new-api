package model

import (
	"fmt"
	"time"

	mainModel "github.com/QuantumNous/new-api/model"
)

type Listing struct {
	Id           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	SellerId     int       `json:"seller_id" gorm:"index"`
	ChannelId    int       `json:"channel_id" gorm:"index"`       // 1:1 mapping to channels table
	ProviderType string    `json:"provider_type"`                  // anthropic, openai, google
	Discount     int       `json:"discount"`                      // percentage off retail, e.g. 30
	TotalCap     int64     `json:"total_cap" gorm:"default:0"`    // lifetime token cap, 0 = unlimited
	HourlyCap    int64     `json:"hourly_cap" gorm:"default:0"`   // per-hour token cap, 0 = unlimited
	DailyCap     int64     `json:"daily_cap" gorm:"default:0"`    // per-day token cap, 0 = unlimited
	TokensUsed   int64     `json:"tokens_used" gorm:"default:0"`  // cumulative tokens consumed
	Label        string    `json:"label"`                         // seller-defined label
	Status       int       `json:"status" gorm:"default:1"`       // active, paused, exhausted
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Listing) TableName() string {
	return "mp_listings"
}

func CreateListing(listing *Listing) error {
	return mainModel.DB.Create(listing).Error
}

func GetListingById(id int) (*Listing, error) {
	var listing Listing
	err := mainModel.DB.First(&listing, id).Error
	if err != nil {
		return nil, err
	}
	return &listing, nil
}

func GetListingByChannelId(channelId int) (*Listing, error) {
	var listing Listing
	err := mainModel.DB.Where("channel_id = ?", channelId).First(&listing).Error
	if err != nil {
		return nil, err
	}
	return &listing, nil
}

func GetListingsBySellerId(sellerId int) ([]*Listing, error) {
	var listings []*Listing
	err := mainModel.DB.Where("seller_id = ?", sellerId).Order("created_at desc").Find(&listings).Error
	return listings, err
}

func GetActiveListings() ([]*Listing, error) {
	var listings []*Listing
	err := mainModel.DB.Where("status = ?", ListingStatusActive).Find(&listings).Error
	return listings, err
}

func UpdateListingStatus(id int, status int) error {
	return mainModel.DB.Model(&Listing{}).Where("id = ?", id).Update("status", status).Error
}

func IncrementListingTokensUsed(id int, tokens int64) (int64, error) {
	listing := &Listing{}
	result := mainModel.DB.Model(listing).Where("id = ?", id).
		Update("tokens_used", mainModel.DB.Raw("tokens_used + ?", tokens))
	if result.Error != nil {
		return 0, result.Error
	}
	// Return the new total
	err := mainModel.DB.Select("tokens_used").First(listing, id).Error
	if err != nil {
		return 0, err
	}
	return listing.TokensUsed, nil
}

func DeleteListing(id int) error {
	return mainModel.DB.Delete(&Listing{}, id).Error
}

// GetMarketDiscountRange returns min and max discounts of active listings for a model
func GetMarketDiscountRange() (int, int, error) {
	var result struct {
		MinDiscount int
		MaxDiscount int
	}
	err := mainModel.DB.Model(&Listing{}).
		Where("status = ?", ListingStatusActive).
		Select("MIN(discount) as min_discount, MAX(discount) as max_discount").
		Scan(&result).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get discount range: %w", err)
	}
	return result.MinDiscount, result.MaxDiscount, nil
}
