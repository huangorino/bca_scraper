package main

import (
	"fmt"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"
)

func GetOrCreateEntity(change *models.BoardroomChange) error {
	title, name := utils.SplitTitle(utils.StringValue(change.PersonName))

	// Step 1: Check if db contains records with the name/display_name
	entities, err := db.FindEntitiesByNameOrDisplay(database, name, utils.StringValue(change.PersonName), utils.IntValue(change.PersonBirthYear))
	if err != nil {
		return fmt.Errorf("FindEntitiesByNameOrDisplay failed: %w", err)
	}

	var permID int
	toInsert := false
	if len(entities) > 0 {
		log.Infof("Found entities: %s", utils.StringValue(entities[0].Name))

		// Check if the stock code exists in any of the entities
		stockCodeFound := false
		for i := range entities {
			entity := entities[i]

			if utils.StringValue(entity.StockCode) == utils.StringValue(change.StockCode) {
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

		err = db.UpdatePrimaryPermID(database, name, utils.StringValue(change.PersonName), utils.IntValue(change.PersonBirthYear), permID)
		if err != nil {
			return fmt.Errorf("UpdatePrimaryPermID failed: %w", err)
		}
	} else {
		toInsert = true
	}

	if toInsert {
		permID, err = db.InsertEntity(database, &models.Entity{
			DisplayName: change.PersonName,
			Name:        &name,
			Salutation:  &title,
			StockCode:   change.StockCode,
			BirthYear:   change.PersonBirthYear,
			Gender:      change.PersonGender,
			Nationality: change.PersonNationality,
			CreatedAt:   *change.DateAnnounced,
		})
		if err != nil {
			return fmt.Errorf("InsertEntity failed: %w", err)
		}
		log.Infof("Inserted new entity: %s", name)
	}

	// Update background information
	if err = db.UpdateBackground(database, permID, &change.Background); err != nil {
		return fmt.Errorf("Qualifications update failed: %w", err)
	}

	change.RelatedPerm = &permID
	change.PersonTitle = &title
	change.PersonName = &name

	return nil
}
