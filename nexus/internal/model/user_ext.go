package model

// UserExt captures additions the new backend makes to the ported User model.
// When moving the full User struct from /home/user/new-api/model/user.go,
// merge these fields into it.
//
//   type User struct {
//       // ... existing new-api fields ...
//       DefaultOrgID       uint   `gorm:"index"`
//       Web3PrimaryAddress string `gorm:"type:varchar(128);index"`
//   }
//
// Likewise Token gains OrgID + X402Enabled + X402MaxPerReqUSD; Channel and
// Log gain a nullable OrgID (null = shared pool / legacy rows).
//
// This file exists as a placeholder so the migration checklist has an
// explicit anchor; it does not declare a GORM model itself.
