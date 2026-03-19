package services

import (
	"fmt"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func GetOrCreateEntity(log *logrus.Logger, database *sqlx.DB, entity *models.Entity, background *models.Background) (*int, error) {
	// Step 1: Check if db contains records with the name/display_name
	entities, err := db.FindEntitiesByNameOrDisplay(database, *entity.Name, *entity.OriName, *entity.BirthYear)
	if err != nil {
		return nil, fmt.Errorf("FindEntitiesByNameOrDisplay failed: %w", err)
	}

	var permID int
	toInsert := false
	if len(entities) > 0 {
		log.Infof("Found entities: %s", utils.StringValue(entities[0].Name))

		// Check if the stock code exists in any of the entities
		stockCodeFound := false
		for _, perm := range entities {
			if utils.StringValue(entity.StockCode) == utils.StringValue(perm.StockCode) {
				stockCodeFound = true
				break
			}
		}

		// Only insert if the stock code is not found in any entity
		if !stockCodeFound {
			toInsert = true
		}

		// Step 2 & 3 & 4: If records found (1 or more), update all their primary_perm_id
		// to be the same as the first record's secondary_perm_id
		permID = entities[0].SecondaryPermID

		err = db.UpdatePrimaryPermID(database, *entity.Name, *entity.OriName, *entity.BirthYear, permID)
		if err != nil {
			return nil, fmt.Errorf("UpdatePrimaryPermID failed: %w", err)
		}
	} else {
		toInsert = true
	}

	if toInsert {
		permID, err = db.InsertEntity(database, entity)
		if err != nil {
			return nil, fmt.Errorf("InsertEntity failed: %w", err)
		}
		log.Infof("Inserted new entity: %s", *entity.Name)
	}

	if background != nil {
		// Update background information
		if err = db.UpdateBackground(database, permID, background); err != nil {
			return nil, fmt.Errorf("Qualifications update failed: %w", err)
		}
	}

	return &permID, nil
}
