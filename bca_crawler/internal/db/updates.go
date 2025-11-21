package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"bca_crawler/internal/models"
)

func UpdateBoardroomChange(db *sqlx.DB, change *models.BoardroomChange) error {
	query := `
		INSERT INTO boardroom_changes (
			company_id, person_id, ann_id,
			category, date_announced, date_of_change,
			designation, previous_position, remarks,
			directorate, type_of_change
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ann_id) DO UPDATE SET
			company_id = excluded.company_id,
			person_id = excluded.person_id,
			category = excluded.category,
			date_announced = excluded.date_announced,
			date_of_change = excluded.date_of_change,
			designation = excluded.designation,
			previous_position = excluded.previous_position,
			remarks = excluded.remarks,
			directorate = excluded.directorate,
			type_of_change = excluded.type_of_change
	`
	_, err := db.Exec(db.Rebind(query),
		change.CompanyID, change.PersonID, change.AnnID,
		change.Category, change.DateAnnounced, change.DateOfChange,
		change.Designation, change.PreviousPosition, change.Remarks,
		change.Directorate, change.TypeOfChange)
	if err != nil {
		return fmt.Errorf("failed to insert or update boardroom change: %w", err)
	}
	return nil
}

func UpdateEntity(db *sqlx.DB, entity *models.Entity) (int64, error) {
	query := `
		INSERT INTO entities (type, name, stock_code, age, gender, nationality) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(type, name, IFNULL(stock_code, '')) DO UPDATE SET
			age = excluded.age,
			gender = excluded.gender,
			nationality = excluded.nationality,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(db.Rebind(query), entity.Type, entity.Name, entity.StockCode, entity.Age, entity.Gender, entity.Nationality)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert entity: %w", err)
	}

	// After the upsert, we need to fetch the ID. LastInsertId is not reliable
	// for updates on conflict, so we query for the row using the unique keys.
	var id int64
	querySelect := `
		SELECT id FROM entities WHERE type = ? AND name = ? AND IFNULL(stock_code, '') = ?
	`
	err = db.QueryRow(db.Rebind(querySelect), entity.Type, entity.Name, entity.StockCode).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve entity ID after upsert: %w", err)
	}
	return id, nil
}

func UpdateBackground(db *sqlx.DB, personID int64, bg *models.Background) error {
	queryInsert := `
		INSERT INTO backgrounds (
		entity_id, qualification, working_experience,
		directorships, family_relationship, conflict_of_interest, interest_in_securities) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(entity_id) DO UPDATE SET
			qualification = excluded.qualification,
			working_experience = excluded.working_experience,
			directorships = excluded.directorships,
			family_relationship = excluded.family_relationship,
			conflict_of_interest = excluded.conflict_of_interest,
			interest_in_securities = excluded.interest_in_securities
	`
	_, err := db.Exec(db.Rebind(queryInsert),
		personID, bg.Qualification, bg.WorkingExperience,
		bg.Directorships, bg.FamilyRelationship, bg.ConflictOfInterest, bg.InterestInSecurities)
	if err != nil {
		return fmt.Errorf("failed to insert qualification: %w", err)
	}

	return nil
}
