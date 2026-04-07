package main

import (
	"fmt"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"
	"sort"
	"github.com/jmoiron/sqlx"
)

// parser for change in boardroom announcements

var database *sqlx.DB
var log = utils.Logger

func main() {
	// -------------------------------------------------------------------------
	// 1️⃣ Load Configuration
	// -------------------------------------------------------------------------
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(fmt.Sprintf("❌ Config load failed: %v", err))
	}

	// Initialize logger
	utils.InitLogger()
	log.Infof("🔧 Configuration loaded: %+v", *cfg)

	// -------------------------------------------------------------------------
	// 2️⃣ Connect to Database
	// -------------------------------------------------------------------------
	database, err = db.Setup(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("❌ Failed to setup DB: %v", err)
	}
	defer database.Close()

	data, err := db.FetchEntity(database)
	if err != nil {
		log.Fatalf("❌ Failed to fetch entities: %v", err)
	}
	if len(data) == 0 {
		log.Info("⚠️ No entities found. Exiting.")
		return
	}

	// -------------------------------------------------------------------------
	// 3️⃣ Filter entities with null primary_perm_id, then group by primary_perm_id
	// -------------------------------------------------------------------------
	grouped := make(map[int][]models.Entity)
	for _, entity := range data {
		if entity.PrimaryPermID == nil {
			continue
		}
		grouped[*entity.PrimaryPermID] = append(grouped[*entity.PrimaryPermID], entity)
	}
	log.Infof("Grouped %d entities into %d primary perm ID groups", len(data), len(grouped))

	// -------------------------------------------------------------------------
	// 4️⃣ Fetch board changes and index by related_perm
	// -------------------------------------------------------------------------
	boardChanges, err := db.FetchBoardChanges(database)
	if err != nil {
		log.Fatalf("❌ Failed to fetch board changes: %v", err)
	}

	changesByPerm := make(map[int][]models.BoardroomChange)
	for _, bc := range boardChanges {
		if bc.RelatedPerm != nil {
			changesByPerm[*bc.RelatedPerm] = append(changesByPerm[*bc.RelatedPerm], bc)
		}
	}

	// -------------------------------------------------------------------------
	// 5️⃣ For each primary perm id group, build entity roles
	// -------------------------------------------------------------------------
	for primaryPermID, entities := range grouped {
		// Collect all board changes linked to any entity in the group
		var allChanges []models.BoardroomChange
		for _, entity := range entities {
			allChanges = append(allChanges, changesByPerm[entity.SecondaryPermID]...)
		}

		// Group changes by stock_code to produce one EntityRole per company
		// Convert map to slice
		// Sort changes by date to ensure chronological processing
		sort.Slice(allChanges, func(i, j int) bool {
			ti := utils.TimeValue(allChanges[i].DateOfChange)
			tj := utils.TimeValue(allChanges[j].DateOfChange)
			if !ti.Equal(tj) {
				return ti.Before(tj)
			}
			// Tie-breaker by announcement ID if dates are same
			return utils.IntValue(allChanges[i].AnnID) < utils.IntValue(allChanges[j].AnnID)
		})

		var rolePtrs []*models.EntityRole
		currentRoleTracker := make(map[string]*models.EntityRole)

		for _, bc := range allChanges {
			input := models.RoleChangeInput{
				PermID:           primaryPermID,
				StockCode:        utils.StringValue(bc.StockCode),
				CompanyName:      utils.StringValue(bc.CompanyName),
				Designation:      utils.StringValue(bc.Designation),
				PreviousPosition: utils.StringValue(bc.PreviousPosition),
				TypeOfChange:     utils.StringValue(bc.TypeOfChange),
				DateOfChange:     bc.DateOfChange,
				Category:         "DIRECTOR",
			}
			newRoles := services.ProcessSingleRoleChange(input, currentRoleTracker)
			rolePtrs = append(rolePtrs, newRoles...)
		}

		var roles []models.EntityRole
		for _, r := range rolePtrs {
			roles = append(roles, *r)
		}

		if len(roles) == 0 {
			log.Warnf("⚠️ No roles found for primary_perm_id=%d, skipping", primaryPermID)
			continue
		}

		// Insert roles into database
		if err := db.InsertEntityRoles(database, roles); err != nil {
			log.Fatalf("❌ Failed to insert entity roles: %v", err)
		}
	}
}
