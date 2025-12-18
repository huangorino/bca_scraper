package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"bca_crawler/internal/models"
)

// SaveAnnouncement inserts or updates a full announcement
func SaveAnnouncement(db *sqlx.DB, a *models.Announcement) error {
	now := time.Now().UTC()

	attachmentsJSON, err := json.Marshal(a.Attachments)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO announcements(
		ann_id, title, link, company_name, stock_name, date_posted, category, ref_number, content, attachments)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT(ann_id)
	DO UPDATE SET
		title = EXCLUDED.title,
		link = EXCLUDED.link,
		company_name = EXCLUDED.company_name,
		stock_name = EXCLUDED.stock_name,
		category = EXCLUDED.category,
		ref_number = EXCLUDED.ref_number,
		content = EXCLUDED.content,
		attachments = EXCLUDED.attachments;`,
		a.AnnID, a.Title, a.Link, a.CompanyName, a.StockName, now, a.Category, a.RefNumber, a.Content, attachmentsJSON)

	return err
}

func UpdateAnnouncement(db *sqlx.DB, a *models.Announcement) error {
	attachmentsJSON, err := json.Marshal(a.Attachments)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO announcements (
		ann_id, company_name, stock_name, date_posted, category, ref_number, attachments, content
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT(ann_id)
	DO UPDATE SET
		company_name = EXCLUDED.company_name,
		stock_name = EXCLUDED.stock_name,
		date_posted = EXCLUDED.date_posted,
		category = EXCLUDED.category,
		ref_number = EXCLUDED.ref_number,
		attachments = EXCLUDED.attachments,
		content = EXCLUDED.content;`,
		a.AnnID, a.CompanyName, a.StockName, a.DatePosted, a.Category, a.RefNumber, attachmentsJSON, a.Content)
	return err
}

func UpdateBoardroomChange(db *sqlx.DB, change *models.BoardroomChange) error {
	query := `
		INSERT INTO boardroom_changes (
			ann_id, company_name, stock_code,
			person_name, person_title, person_birth_year,
			person_gender, person_nationality,
			date_announced, date_of_change,
			designation, previous_position, remarks,
			directorate, type_of_change
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT(ann_id) DO UPDATE SET
			company_name = excluded.company_name,
			stock_code = excluded.stock_code,
			person_name = excluded.person_name,
			person_title = excluded.person_title,
			person_birth_year = excluded.person_birth_year,
			person_gender = excluded.person_gender,
			person_nationality = excluded.person_nationality,
			date_announced = excluded.date_announced,
			date_of_change = excluded.date_of_change,
			designation = excluded.designation,
			previous_position = excluded.previous_position,
			remarks = excluded.remarks,
			directorate = excluded.directorate,
			type_of_change = excluded.type_of_change
	`
	_, err := db.Exec(db.Rebind(query),
		change.AnnID, change.CompanyName, change.StockCode,
		change.PersonName, change.PersonTitle, change.PersonBirthYear,
		change.PersonGender, change.PersonNationality,
		change.DateAnnounced, change.DateOfChange,
		change.Designation, change.PreviousPosition, change.Remarks,
		change.Directorate, change.TypeOfChange)
	if err != nil {
		return fmt.Errorf("failed to insert or update boardroom change: %w", err)
	}
	return nil
}

func InsertEntityMaster(db *sqlx.DB, e *models.EntityMaster) error {
	query := `
		INSERT INTO entities_master (
			sc_id, type, name,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(sc_id) DO UPDATE SET
			type = excluded.type,
			name = excluded.name,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(
		db.Rebind(query),
		e.ScID,
		e.Type,
		e.Name,
		e.CreatedAt,
		e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert entity master: %w", err)
	}
	return nil
}

func UpdateEntity(db *sqlx.DB, e *models.Entity) error {
	query := `
		INSERT INTO entities (
			sc_id, prefix, name,
			title, birth_year, gender,
			nationality, created_at
		)
		VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)
		ON CONFLICT (sc_id) DO UPDATE SET
			title       = EXCLUDED.title,
			gender      = EXCLUDED.gender,
			nationality = EXCLUDED.nationality,
			updated_at  = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(
		db.Rebind(query),
		e.ScID,
		e.Prefix,
		e.Name,
		e.Title,
		e.BirthYear,
		e.Gender,
		e.Nationality,
		e.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert entity: %w", err)
	}

	return nil
}

func UpdateBackground(db *sqlx.DB, personID uuid.UUID, bg *models.Background) error {
	queryInsert := `
		INSERT INTO backgrounds (
			sc_id, qualification, working_experience,
			directorships, family_relationship, conflict_of_interest, interest_in_securities) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(sc_id) DO UPDATE SET
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
