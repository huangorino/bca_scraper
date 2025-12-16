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

func UpdateEntity(db *sqlx.DB, e *models.Entity) (int64, error) {
	var (
		conflictCols  string
		conflictWhere string
		stockCode     any = nil
		birthYear     any = nil
	)

	if e.Type == "COMPANY" {
		conflictCols = "(type, name, stock_code)"
		conflictWhere = "type = 'COMPANY'"
		stockCode = e.StockCode
	} else {
		conflictCols = "(type, name, birth_year)"
		conflictWhere = "type = 'PERSON'"
		birthYear = e.BirthYear
	}

	query := fmt.Sprintf(`
	INSERT INTO entities (
		type,
		name,
		title,
		stock_code,
		birth_year,
		gender,
		nationality,
		created_at
	)
	VALUES (
		?, ?, ?, ?, ?, ?, ?, ?
	)
	ON CONFLICT %s
	WHERE %s
	DO UPDATE SET
		title       = EXCLUDED.title,
		gender      = EXCLUDED.gender,
		nationality = EXCLUDED.nationality,
		updated_at  = CURRENT_TIMESTAMP
	RETURNING id
	`, conflictCols, conflictWhere)

	var id int64
	err := db.QueryRowx(
		db.Rebind(query),
		e.Type,
		e.Name,
		e.Title,
		stockCode,
		birthYear,
		e.Gender,
		e.Nationality,
		e.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert entity: %w", err)
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
