package service

import (
	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

// EnsureMarketplaceGroup registers the "marketplace" group in both
// UserUsableGroups and GroupRatio if not already present.
// Called once at startup.
func EnsureMarketplaceGroup() {
	groupName := mpModel.MarketplaceGroup

	// Add to user usable groups
	groups := setting.GetUserUsableGroupsCopy()
	if _, exists := groups[groupName]; !exists {
		groups[groupName] = "Marketplace"
		jsonStr, err := common.Marshal(groups)
		if err == nil {
			_ = setting.UpdateUserUsableGroupsByJSONString(string(jsonStr))
			common.SysLog("marketplace: registered 'marketplace' in user usable groups")
		}
	}

	// Add to group ratio (1.0 = no markup, price same as official)
	if !ratio_setting.ContainsGroupRatio(groupName) {
		ratios := ratio_setting.GetGroupRatioCopy()
		ratios[groupName] = 1.0
		jsonStr, err := common.Marshal(ratios)
		if err == nil {
			_ = ratio_setting.UpdateGroupRatioByJSONString(string(jsonStr))
			common.SysLog("marketplace: registered 'marketplace' in group ratios with ratio 1.0")
		}
	}
}
