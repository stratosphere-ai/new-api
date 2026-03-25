package model

import (
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// AgentKey is a credential that users create and give to their AI agents.
// It carries pre-configured preferences (models, budget, expiry policy)
// so the agent only needs this one key to interact with the platform.
type AgentKey struct {
	Id          int            `json:"id" gorm:"primaryKey"`
	UserId      int            `json:"user_id" gorm:"index"`
	Name        string         `json:"name" gorm:"type:varchar(50);index"`                      // Human-readable name, e.g. "Lobster Bot"
	Key         string         `json:"key" gorm:"type:char(48);uniqueIndex"`                    // The ak-xxx key
	Status      int            `json:"status" gorm:"default:1"`                                 // 1=active, 0=disabled
	Models      string         `json:"models" gorm:"type:varchar(1024);default:''"`             // Comma-separated allowed models (empty = all)
	MonthlyBudget int          `json:"monthly_budget" gorm:"default:0"`                         // Monthly auto-approve budget (quota units), 0=unlimited
	MaxQuotaPerRequest int     `json:"max_quota_per_request" gorm:"default:0"`                  // Max quota per single provision (0=no limit)
	DefaultExpire string       `json:"default_expire" gorm:"type:varchar(10);default:'30d'"`    // Default token expiry
	AllowIps    *string        `json:"allow_ips" gorm:"type:text;default:''"`                   // IP whitelist (empty = all)
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (ak *AgentKey) Insert() error {
	return DB.Create(ak).Error
}

func (ak *AgentKey) Update() error {
	return DB.Model(ak).Select(
		"name", "status", "models", "monthly_budget",
		"max_quota_per_request", "default_expire", "allow_ips",
	).Updates(ak).Error
}

func (ak *AgentKey) Delete() error {
	return DB.Delete(ak).Error
}

func (ak *AgentKey) TouchLastUsed() {
	now := time.Now()
	DB.Model(ak).Update("last_used_at", &now)
}

func GetAgentKeyByKey(key string) (*AgentKey, error) {
	if key == "" {
		return nil, errors.New("agent key is empty")
	}
	var ak AgentKey
	err := DB.Where(commonKeyCol+" = ? AND status = 1", key).First(&ak).Error
	return &ak, err
}

func GetAgentKeysByUserId(userId int) ([]*AgentKey, error) {
	var keys []*AgentKey
	err := DB.Where("user_id = ?", userId).Order("id desc").Find(&keys).Error
	return keys, err
}

func GetAgentKeyByIdAndUserId(id int, userId int) (*AgentKey, error) {
	var ak AgentKey
	err := DB.Where("id = ? AND user_id = ?", id, userId).First(&ak).Error
	return &ak, err
}

// GetAgentKeyMonthlySpending returns total quota spent via this agent key this month.
func GetAgentKeyMonthlySpending(agentKeyId int) (int64, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var total int64
	err := DB.Model(&MarketplacePurchase{}).
		Where("agent_key_id = ? AND status = ? AND created_at >= ?",
			agentKeyId, MarketplacePurchaseStatusCompleted, monthStart).
		Select("COALESCE(SUM(quota), 0)").
		Scan(&total).Error
	return total, err
}

func (ak *AgentKey) GetAllowedModels() []string {
	if ak.Models == "" {
		return nil
	}
	models := strings.Split(ak.Models, ",")
	result := make([]string, 0, len(models))
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}

func (ak *AgentKey) GetIpWhitelist() []string {
	if ak.AllowIps == nil || *ak.AllowIps == "" {
		return nil
	}
	ips := strings.Split(strings.ReplaceAll(*ak.AllowIps, " ", ""), "\n")
	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			result = append(result, ip)
		}
	}
	return result
}
