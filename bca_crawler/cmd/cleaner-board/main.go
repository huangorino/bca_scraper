package main

import (
	"fmt"
	"strings"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"

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

	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	data_exhist, err := db.FetchExHist(database)
	data_entity, err := db.FetchEntity(database)
	data_board, err := db.FetchBoardChanges(database)

	toInsertList := []models.BoardroomChange{}

	// Loop through exhist records
	for h := range data_exhist {
		current_h := data_exhist[h]
		var matchedEntity *models.Entity

		// Find matching entity
		for e := range data_entity {
			current_e := data_entity[e]

			if utils.StringValue(current_h.DirectorName) == utils.StringValue(current_e.Name) ||
				utils.StringValue(current_h.DirectorName) == utils.StringValue(current_e.DisplayName) ||
				utils.StringValue(current_h.DirectorName) == utils.StringValue(current_e.OriName) ||
				strings.TrimSpace(utils.StringValue(current_h.Title)+" "+utils.StringValue(current_h.DirectorName)) == utils.StringValue(current_e.Name) ||
				strings.TrimSpace(utils.StringValue(current_h.Title)+" "+utils.StringValue(current_h.DirectorName)) == utils.StringValue(current_e.DisplayName) ||
				strings.TrimSpace(utils.StringValue(current_h.Title)+" "+utils.StringValue(current_h.DirectorName)) == utils.StringValue(current_e.OriName) {

				matchedEntity = &current_e
				break
			}
		}

		// Skip if no entity matched
		if matchedEntity == nil {
			log.Warnf("⚠️ No entity found for %s", utils.StringValue(current_h.DirectorName))
			continue
		}

		var id int

		if matchedEntity.PrimaryPermID != nil {
			id = *matchedEntity.PrimaryPermID
		} else {
			id = 0
		}

		// Track if we found matching records in boardchanges
		foundAppointmentMatch := false
		foundResignationMatch := false
		foundStockCodeMatch := false

		// Loop through boardchanges to find matching records
		for b := range data_board {
			current_b := data_board[b]

			// Check if board change matches the entity name
			nameMatches := utils.StringValue(current_b.PersonName) == utils.StringValue(matchedEntity.Name) ||
				utils.StringValue(current_b.PersonName) == utils.StringValue(matchedEntity.DisplayName) ||
				utils.StringValue(current_b.PersonName) == utils.StringValue(matchedEntity.OriName) ||
				strings.TrimSpace(utils.StringValue(current_b.PersonTitle)+" "+utils.StringValue(current_b.PersonName)) == utils.StringValue(matchedEntity.Name) ||
				strings.TrimSpace(utils.StringValue(current_b.PersonTitle)+" "+utils.StringValue(current_b.PersonName)) == utils.StringValue(matchedEntity.DisplayName) ||
				strings.TrimSpace(utils.StringValue(current_b.PersonTitle)+" "+utils.StringValue(current_b.PersonName)) == utils.StringValue(matchedEntity.OriName)

			if !nameMatches {
				continue
			}

			// Check stock code matches
			if utils.StringValue(current_b.StockCode) != utils.StringValue(current_h.StockCode) {
				continue
			}

			// Mark that we found at least one record with matching stock code
			foundStockCodeMatch = true

			// Check TypeOfChange and compare dates
			if utils.StringValue(current_b.TypeOfChange) == "Appointment" {
				// Compare with AppDate
				if (*current_b.DateOfChange).Equal(utils.TimeValue(current_h.AppDate)) {
					foundAppointmentMatch = true
				}
			} else if utils.StringValue(current_b.TypeOfChange) == "Resignation" ||
				utils.StringValue(current_b.TypeOfChange) == "Retirement" ||
				utils.StringValue(current_b.TypeOfChange) == "Cessation of Office" ||
				utils.StringValue(current_b.TypeOfChange) == "Vacation Of Office" ||
				utils.StringValue(current_b.TypeOfChange) == "Others" {
				// Compare with ResDate
				if (*current_b.DateOfChange).Equal(utils.TimeValue(current_h.ResDate)) {
					foundResignationMatch = true
				}
			}
		}

		// If no stock code match found, insert both records
		if !foundStockCodeMatch {
			foundAppointmentMatch = false
			foundResignationMatch = false
		}

		// Insert Appointment record if AppDate doesn't match
		appDateValue := utils.TimeValue(current_h.AppDate)
		if !foundAppointmentMatch && !appDateValue.IsZero() {
			// Skip if AppDate is 1900-01-01 (invalid/placeholder date)
			if appDateValue.Year() == 1900 && appDateValue.Month() == 1 && appDateValue.Day() == 1 {
			} else {
				log.Infof("Inserting Appointment record for %s - %s (AppDate: %s)", utils.StringValue(current_h.StockCode), utils.StringValue(current_h.DirectorName), appDateValue.Format("2006-01-02"))

				birthYear := utils.TimeValue(current_h.BirthDate).Year()
				toInsertList = append(toInsertList, models.BoardroomChange{
					CompanyName:       current_h.CompanyName,
					StockCode:         current_h.StockCode,
					PersonName:        current_h.DirectorName,
					PersonTitle:       current_h.Title,
					PersonBirthYear:   &birthYear,
					PersonGender:      current_h.Gender,
					PersonNationality: current_h.Nationality,
					DateAnnounced:     current_h.AppDate,
					DateOfChange:      current_h.AppDate,
					Designation:       current_h.Designation,
					TypeOfChange:      utils.PtrString("Appointment"),
					RelatedPerm:       &id,
				})
			}
		}

		// Insert Resignation record if ResDate doesn't match
		resDateValue := utils.TimeValue(current_h.ResDate)
		if !foundResignationMatch && !resDateValue.IsZero() {
			// Skip if ResDate is 1900-01-01 (invalid/placeholder date)
			if resDateValue.Year() == 1900 && resDateValue.Month() == 1 && resDateValue.Day() == 1 {
			} else {
				log.Infof("Inserting Resignation record for %s - %s (ResDate: %s)", utils.StringValue(current_h.StockCode), utils.StringValue(current_h.DirectorName), resDateValue.Format("2006-01-02"))

				birthYear := utils.TimeValue(current_h.BirthDate).Year()
				toInsertList = append(toInsertList, models.BoardroomChange{
					CompanyName:       current_h.CompanyName,
					StockCode:         current_h.StockCode,
					PersonName:        current_h.DirectorName,
					PersonTitle:       current_h.Title,
					PersonBirthYear:   &birthYear,
					PersonGender:      current_h.Gender,
					PersonNationality: current_h.Nationality,
					DateAnnounced:     current_h.ResDate,
					DateOfChange:      current_h.ResDate,
					PreviousPosition:  current_h.Designation,
					TypeOfChange:      utils.PtrString("Resignation"),
					RelatedPerm:       &id,
				})
			}
		}
	}

	updated := 0
	for i := range toInsertList {
		err := db.UpdateBoardroomChange(database, &toInsertList[i])
		if err != nil {
			log.Errorf("❌ Entity insert failed for ann_id %s: %v", utils.StringValue(toInsertList[i].PersonName), err)
			continue
		}

		updated++

		log.Infof("🏁 Done. Updated %d records.", updated)
	}
}
