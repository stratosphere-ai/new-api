package service

import (
	"fmt"

	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	"github.com/QuantumNous/new-api/model"
)

// RegisterSeller creates a seller record for a user
func RegisterSeller(userId int) (*mpModel.Seller, error) {
	// Verify user exists
	user, err := model.GetUserById(userId, false)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if already a seller
	existing, _ := mpModel.GetSellerByUserId(userId)
	if existing != nil {
		return existing, nil // idempotent
	}

	seller := &mpModel.Seller{
		UserId: userId,
		Status: mpModel.SellerStatusActive,
	}
	if err := mpModel.CreateSeller(seller); err != nil {
		return nil, fmt.Errorf("failed to create seller: %w", err)
	}
	return seller, nil
}

// GetSellerDashboard returns seller stats
type SellerDashboard struct {
	Seller   *mpModel.Seller    `json:"seller"`
	Listings []*mpModel.Listing `json:"listings"`
}

func GetSellerDashboard(userId int) (*SellerDashboard, error) {
	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("seller not found")
	}

	listings, err := mpModel.GetListingsBySellerId(seller.Id)
	if err != nil {
		return nil, err
	}

	return &SellerDashboard{
		Seller:   seller,
		Listings: listings,
	}, nil
}
