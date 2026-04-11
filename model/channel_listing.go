package model

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/logger"
	"gorm.io/gorm"
)

// ChannelListing represents a user-submitted channel on the marketplace.
// It links a user (seller) to a channel they contributed, tracking
// revenue sharing and listing status independently from the channel itself.
type ChannelListing struct {
	Id              int            `json:"id" gorm:"primaryKey"`
	UserId          int            `json:"user_id" gorm:"index"`                              // The seller who submitted this channel
	ChannelId       int            `json:"channel_id" gorm:"uniqueIndex"`                     // The actual channel in the channel pool
	Status          int            `json:"status" gorm:"default:1;index"`                     // 1=active, 0=paused, 2=under_review, 3=rejected
	RevenueSharePct int            `json:"revenue_share_pct" gorm:"default:85"`               // Seller gets this %, platform keeps the rest
	MonthlyCapQuota int64          `json:"monthly_cap_quota" gorm:"default:0"`                // Max quota others can consume per month (0=unlimited)
	Models          string         `json:"models" gorm:"type:varchar(1024)"`                  // Copy of channel models for display
	ChannelType     int            `json:"channel_type"`                                      // Copy of channel type for display
	DisplayName     string         `json:"display_name" gorm:"type:varchar(100)"`             // Seller-facing name
	Description     string         `json:"description" gorm:"type:varchar(500)"`              // Seller's description
	PricePerMToken  float64        `json:"price_per_m_token" gorm:"default:0"`                // Seller's asking price per 1M tokens (0=platform default)
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

const (
	ListingStatusActive      = 1
	ListingStatusPaused      = 0
	ListingStatusUnderReview = 2
	ListingStatusRejected    = 3
)

func (l *ChannelListing) Insert() error {
	return DB.Create(l).Error
}

func (l *ChannelListing) Update() error {
	return DB.Model(l).Select(
		"status", "revenue_share_pct", "monthly_cap_quota",
		"display_name", "description", "price_per_m_token",
	).Updates(l).Error
}

func (l *ChannelListing) Delete() error {
	return DB.Delete(l).Error
}

func GetChannelListingsByUserId(userId int) ([]*ChannelListing, error) {
	var listings []*ChannelListing
	err := DB.Where("user_id = ?", userId).Order("id desc").Find(&listings).Error
	return listings, err
}

func GetChannelListingByIdAndUserId(id int, userId int) (*ChannelListing, error) {
	var listing ChannelListing
	err := DB.Where("id = ? AND user_id = ?", id, userId).First(&listing).Error
	return &listing, err
}

func GetChannelListingByChannelId(channelId int) (*ChannelListing, error) {
	var listing ChannelListing
	err := DB.Where("channel_id = ?", channelId).First(&listing).Error
	return &listing, err
}

func GetAllActiveListings() ([]*ChannelListing, error) {
	var listings []*ChannelListing
	err := DB.Where("status = ?", ListingStatusActive).Find(&listings).Error
	return listings, err
}

// ChannelEarning tracks monthly revenue for each listing.
type ChannelEarning struct {
	Id              int       `json:"id" gorm:"primaryKey"`
	ListingId       int       `json:"listing_id" gorm:"index"`
	UserId          int       `json:"user_id" gorm:"index"`
	ChannelId       int       `json:"channel_id" gorm:"index"`
	Month           string    `json:"month" gorm:"type:char(7);index"` // "2026-04"
	ConsumedQuota   int64     `json:"consumed_quota"`                  // Total quota consumed by buyers
	GrossRevenue    int64     `json:"gross_revenue"`                   // Revenue before platform cut (quota units)
	NetRevenue      int64     `json:"net_revenue"`                     // Revenue after platform cut (quota units)
	RevenueSharePct int       `json:"revenue_share_pct"`               // Snapshot of share % at time of earning
	SettledAt       *time.Time `json:"settled_at"`                     // When payout was made (nil = pending)
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (e *ChannelEarning) Insert() error {
	return DB.Create(e).Error
}

// GetOrCreateMonthlyEarning returns the earning record for a listing in the current month.
func GetOrCreateMonthlyEarning(listingId int, userId int, channelId int, revenueSharePct int) (*ChannelEarning, error) {
	month := time.Now().Format("2006-01")
	var earning ChannelEarning
	err := DB.Where("listing_id = ? AND month = ?", listingId, month).First(&earning).Error
	if err == nil {
		return &earning, nil
	}
	// Create new
	earning = ChannelEarning{
		ListingId:       listingId,
		UserId:          userId,
		ChannelId:       channelId,
		Month:           month,
		RevenueSharePct: revenueSharePct,
	}
	if err := earning.Insert(); err != nil {
		return nil, err
	}
	return &earning, nil
}

// AddEarning atomically adds consumed quota and calculates revenue.
func AddEarning(listingId int, userId int, channelId int, quotaConsumed int64, revenueSharePct int) error {
	month := time.Now().Format("2006-01")

	netQuota := quotaConsumed * int64(revenueSharePct) / 100

	result := DB.Model(&ChannelEarning{}).
		Where("listing_id = ? AND month = ?", listingId, month).
		Updates(map[string]interface{}{
			"consumed_quota": gorm.Expr("consumed_quota + ?", quotaConsumed),
			"gross_revenue":  gorm.Expr("gross_revenue + ?", quotaConsumed),
			"net_revenue":    gorm.Expr("net_revenue + ?", netQuota),
		})

	if result.RowsAffected == 0 {
		// Record doesn't exist yet, create it
		earning := ChannelEarning{
			ListingId:       listingId,
			UserId:          userId,
			ChannelId:       channelId,
			Month:           month,
			ConsumedQuota:   quotaConsumed,
			GrossRevenue:    quotaConsumed,
			NetRevenue:      netQuota,
			RevenueSharePct: revenueSharePct,
		}
		return earning.Insert()
	}
	return result.Error
}

// GetUserEarnings returns all earnings for a user, optionally filtered by month.
func GetUserEarnings(userId int, month string) ([]*ChannelEarning, error) {
	var earnings []*ChannelEarning
	query := DB.Where("user_id = ?", userId)
	if month != "" {
		query = query.Where("month = ?", month)
	}
	err := query.Order("month desc, id desc").Find(&earnings).Error
	return earnings, err
}

// GetUserTotalUnsettled returns the total unsettled net revenue for a user.
func GetUserTotalUnsettled(userId int) (int64, error) {
	var total int64
	err := DB.Model(&ChannelEarning{}).
		Where("user_id = ? AND settled_at IS NULL", userId).
		Select("COALESCE(SUM(net_revenue), 0)").
		Scan(&total).Error
	return total, err
}

// GetListingMonthlyConsumed returns the total quota consumed for a listing this month.
func GetListingMonthlyConsumed(listingId int) (int64, error) {
	month := time.Now().Format("2006-01")
	var total int64
	err := DB.Model(&ChannelEarning{}).
		Where("listing_id = ? AND month = ?", listingId, month).
		Select("COALESCE(SUM(consumed_quota), 0)").
		Scan(&total).Error
	return total, err
}

// SettleUserEarnings marks all unsettled earnings as settled and returns the total.
func SettleUserEarnings(userId int) (int64, error) {
	now := time.Now()
	var total int64
	err := DB.Model(&ChannelEarning{}).
		Where("user_id = ? AND settled_at IS NULL", userId).
		Select("COALESCE(SUM(net_revenue), 0)").
		Scan(&total).Error
	if err != nil || total == 0 {
		return 0, err
	}

	err = DB.Model(&ChannelEarning{}).
		Where("user_id = ? AND settled_at IS NULL", userId).
		Update("settled_at", &now).Error
	if err != nil {
		return 0, fmt.Errorf("failed to mark earnings as settled: %v", err)
	}

	// Add to user's quota balance
	err = IncreaseUserQuota(userId, int(total), true)
	if err != nil {
		return total, fmt.Errorf("earnings settled but failed to add quota: %v", err)
	}

	RecordLog(userId, LogTypeTopup,
		fmt.Sprintf("渠道代挂结算：%s quota 已转入余额", logger.FormatQuota(int(total))))

	return total, nil
}
