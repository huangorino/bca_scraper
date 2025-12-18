package main

import (
	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

func GetOrCreateEntity(change models.BoardroomChange) error {
	var entities []models.Entity

	// Company
	companyID, err := db.GetSCID(database, change.CompanyName, "COMPANY")
	if errors.Is(err, sql.ErrNoRows) {
		companyID, err = db.InsertEntityMaster(database, &models.EntityMaster{
			Type:      "COMPANY",
			Name:      change.CompanyName,
			CreatedAt: *change.DateAnnounced,
		})
		if err != nil {
			return fmt.Errorf("InsertEntityMaster failed: %w", err)
		}

		entities = append(entities, models.Entity{
			ScID:      companyID,
			Prefix:    "STOCK CODE",
			Value:     change.StockCode,
			CreatedAt: *change.DateAnnounced,
		})
	}
	if err != nil {
		return fmt.Errorf("Get Company SCID failed: %w", err)
	}

	// Person
	title, name := utils.SplitTitle(change.PersonName)
	personID, err := db.GetSCID(database, name, "PERSON")
	if errors.Is(err, sql.ErrNoRows) {
		personID, err = db.InsertEntityMaster(database, &models.EntityMaster{
			Type:      "PERSON",
			Name:      name,
			CreatedAt: *change.DateAnnounced,
		})
		if err != nil {
			return fmt.Errorf("InsertEntityMaster failed: %w", err)
		}

		data := []struct {
			prefix string
			value  string
		}{
			{"TITLE", title},
			{"BIRTH YEAR", strconv.Itoa(change.PersonBirthYear)},
			{"GENDER", change.PersonGender},
			{"NATIONALITY", change.PersonNationality},
		}

		for _, d := range data {
			entities = append(entities, models.Entity{
				ScID:      personID,
				Prefix:    d.prefix,
				Value:     d.value,
				CreatedAt: *change.DateAnnounced,
			})
		}
	}
	if err != nil {
		return fmt.Errorf("Get PersonSCID failed: %w", err)
	}

	for _, entity := range entities {
		if err = db.InsertEntity(database, &entity); err != nil {
			return fmt.Errorf("InsertEntity failed: %w", err)
		}
	}

	// Qualifications
	if err = db.UpdateBackground(database, personID, &change.Background); err != nil {
		return fmt.Errorf("Qualifications update failed: %w", err)
	}

	return nil
}
