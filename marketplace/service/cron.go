package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
)

var settlementOnce sync.Once

// StartMonthlySettlementCron starts a background goroutine that runs
// monthly settlement on the 1st of each month at 00:05 UTC.
func StartMonthlySettlementCron() {
	settlementOnce.Do(func() {
		go func() {
			for {
				now := time.Now().UTC()
				// Calculate next 1st of month at 00:05
				next := time.Date(now.Year(), now.Month()+1, 1, 0, 5, 0, 0, time.UTC)
				sleepDuration := next.Sub(now)
				common.SysLog(fmt.Sprintf("marketplace: next settlement scheduled at %s (in %s)", next.Format(time.RFC3339), sleepDuration))
				time.Sleep(sleepDuration)

				RunMonthlySettlement()
			}
		}()
	})
}

// RunMonthlySettlement processes all sellers with positive balance,
// creates withdrawal records, and resets their balances.
func RunMonthlySettlement() {
	now := time.Now().UTC()
	periodEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodStart := periodEnd.AddDate(0, -1, 0)

	common.SysLog(fmt.Sprintf("marketplace: running monthly settlement for period %s to %s",
		periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")))

	sellers, err := mpModel.GetSellersWithBalance()
	if err != nil {
		common.SysError(fmt.Sprintf("marketplace: failed to get sellers with balance: %v", err))
		return
	}

	if len(sellers) == 0 {
		common.SysLog("marketplace: no sellers with balance to settle")
		return
	}

	settled := 0
	for _, seller := range sellers {
		amount, err := mpModel.ResetSellerBalance(seller.Id)
		if err != nil {
			common.SysError(fmt.Sprintf("marketplace: failed to reset balance for seller %d: %v", seller.Id, err))
			continue
		}
		if amount <= 0 {
			continue
		}

		withdrawal := &mpModel.Withdrawal{
			SellerId:    seller.Id,
			Amount:      amount,
			Status:      mpModel.WithdrawalStatusPending,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
		}
		if err := mpModel.CreateWithdrawal(withdrawal); err != nil {
			common.SysError(fmt.Sprintf("marketplace: failed to create withdrawal for seller %d: %v", seller.Id, err))
			// Try to restore the balance
			_ = mpModel.IncrementSellerBalance(seller.Id, amount)
			continue
		}

		settled++
		common.SysLog(fmt.Sprintf("marketplace: created withdrawal #%d for seller %d, amount %d", withdrawal.Id, seller.Id, amount))
	}

	common.SysLog(fmt.Sprintf("marketplace: settlement complete, %d/%d sellers settled", settled, len(sellers)))
}
