package main

import (
	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"
	"fmt"
)

func GetOrCreateEntity(change models.BoardroomChange) error {
	title, name := utils.SplitTitle(change.PersonName)

	// Step 1: Check if db contains records with the name/display_name
	entities, err := db.FindEntitiesByNameOrDisplay(database, name, change.PersonName)
	if err != nil {
		return fmt.Errorf("FindEntitiesByNameOrDisplay failed: %w", err)
	}

	var permID int
	toInsert := false
	if len(entities) > 0 {
		// Step 2 & 3 & 4: If records found (1 or more), update all their primary_perm_id
		// to be the same as the first record's secondary_perm_id
		permID = entities[0].SecondaryPermID

		err = db.UpdatePrimaryPermID(database, name, change.PersonName, permID)
		if err != nil {
			return fmt.Errorf("UpdatePrimaryPermID failed: %w", err)
		}
	} else {
		toInsert = true
	}

	for i := range entities {
		entity := entities[i]

		if entity.StockCode != &change.StockCode {
			toInsert = true
		}
	}

	if toInsert {
		permID, err = db.InsertEntity(database, &models.Entity{
			DisplayName: change.PersonName,
			Name:        &name,
			Salutation:  &title,
			StockCode:   &change.StockCode,
			BirthYear:   &change.PersonBirthYear,
			Gender:      &change.PersonGender,
			Nationality: &change.PersonNationality,
			CreatedAt:   *change.DateAnnounced,
		})
		if err != nil {
			return fmt.Errorf("InsertEntity failed: %w", err)
		}
	}

	// Update background information
	if err = db.UpdateBackground(database, permID, &change.Background); err != nil {
		return fmt.Errorf("Qualifications update failed: %w", err)
	}

	return nil
}
