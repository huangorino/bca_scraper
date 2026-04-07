package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
		ann_id, title, company_name, stock_name, date_posted, category, ref_number, attachments, content
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT(ann_id)
	DO UPDATE SET
		title = EXCLUDED.title,
		company_name = EXCLUDED.company_name,
		stock_name = EXCLUDED.stock_name,
		date_posted = EXCLUDED.date_posted,
		category = EXCLUDED.category,
		ref_number = EXCLUDED.ref_number,
		attachments = EXCLUDED.attachments,
		content = EXCLUDED.content;`,
		a.AnnID, a.Title, a.CompanyName, a.StockName, a.DatePosted, a.Category, a.RefNumber, attachmentsJSON, a.Content)
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
			directorate, type_of_change, related_perm
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
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
			type_of_change = excluded.type_of_change,
			related_perm = excluded.related_perm
	`
	_, err := db.Exec(db.Rebind(query),
		change.AnnID, change.CompanyName, change.StockCode,
		change.PersonName, change.PersonTitle, change.PersonBirthYear,
		change.PersonGender, change.PersonNationality,
		change.DateAnnounced, change.DateOfChange,
		change.Designation, change.PreviousPosition, change.Remarks,
		change.Directorate, change.TypeOfChange, change.RelatedPerm)
	if err != nil {
		return fmt.Errorf("failed to insert or update boardroom change: %w", err)
	}
	return nil
}

func UpdateShareholdingChange(db *sqlx.DB, changes []*models.ShareholdingChange) error {
	if len(changes) == 0 {
		return nil
	}

	annID := changes[0].AnnID

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing rows for this announcement
	_, err = tx.Exec("DELETE FROM shareholding_change WHERE ann_id = $1", annID)
	if err != nil {
		return fmt.Errorf("delete existing records: %w", err)
	}

	query := `
	INSERT INTO shareholding_change (
		ann_id,
		stock_code,
		company_name,
		change_type,
		person_name,
		person_address,
		person_nationality,
		company_no,
		security_description,
		registered_holder,
		registered_holder_address,
		transaction_type,
		transaction_desc,
		currency,
		date_of_change,
		date_interest_acquired,
		date_of_cessation,
		securities_changed,
		price_transacted,
		nature_of_interest,
		circumstances,
		consideration,
		direct_units,
		direct_percent,
		indirect_units,
		indirect_percent,
		total_securities,
		date_of_notice,
		date_notice_received,
		remarks
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
		$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
		$21,$22,$23,$24,$25,$26,$27,$28,$29,$30
	)
	`

	stmt, err := tx.Preparex(tx.Rebind(query))
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, change := range changes {

		_, err := stmt.Exec(
			change.AnnID,
			change.StockCode,
			change.CompanyName,
			change.ChangeType,
			change.PersonName,
			change.PersonAddress,
			change.PersonNationality,
			change.CompanyNo,
			change.SecurityDescription,
			change.RegisteredHolder,
			change.RegisteredHolderAddress,
			change.TransactionType,
			change.TransactionDesc,
			change.Currency,
			change.DateOfChange,
			change.DateInterestAcquired,
			change.DateOfCessation,
			change.SecuritiesChanged,
			change.PriceTransacted,
			change.NatureOfInterest,
			change.Circumstances,
			change.Consideration,
			change.DirectUnits,
			change.DirectPercent,
			change.IndirectUnits,
			change.IndirectPercent,
			change.TotalSecurities,
			change.DateOfNotice,
			change.DateNoticeReceived,
			change.Remarks,
		)

		if err != nil {
			return fmt.Errorf("insert record: %w", err)
		}
	}

	return tx.Commit()
}

func UpdateShareholdingChangePerm(db *sqlx.DB, id int, permID int) error {
	_, err := db.Exec("UPDATE shareholding_change SET related_perm = $1 WHERE id = $2", permID, id)
	return err
}

func InsertEntity(db *sqlx.DB, e *models.Entity) (int, error) {
	query := `
		INSERT INTO entities (
			primary_perm_id,
			display_name, ori_name, name, salutation, stock_code,
			birth_year, gender, nationality, new_ic,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING secondary_perm_id
	`

	var scID int
	err := db.QueryRowx(
		db.Rebind(query),
		e.PrimaryPermID,
		e.DisplayName,
		e.OriName,
		e.Name,
		e.Salutation,
		e.StockCode,
		e.BirthYear,
		e.Gender,
		e.Nationality,
		e.NewIC,
		e.CreatedAt,
	).Scan(&scID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert entity: %w", err)
	}

	// Update primary_perm_id to be the same as secondary_perm_id
	// updateQuery := `UPDATE entities SET primary_perm_id = ?, updated_at = CURRENT_TIMESTAMP WHERE secondary_perm_id = ?`
	// _, err = db.Exec(db.Rebind(updateQuery), scID, scID)
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to update primary_perm_id: %w", err)
	// }

	return scID, nil
}

func UpdateBackground(db *sqlx.DB, personID int, bg *models.Background) error {
	queryInsert := `
		INSERT INTO backgrounds (
			perm_id, qualification, working_experience,
			directorships, family_relationship, conflict_of_interest, interest_in_securities) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(perm_id) DO UPDATE SET
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

func InsertEntityRoles(db *sqlx.DB, roles []models.EntityRole) error {
	query := `
		INSERT INTO entities_role (
			perm_id, company_name, stock_name,
			date_appointed, date_resigned,
			category, role_name,
			promoted, alternative, independentcy, executive, chairmanship
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(perm_id, stock_name, role_name, date_appointed) DO UPDATE SET
			date_resigned = EXCLUDED.date_resigned,
			date_updated = CURRENT_TIMESTAMP
	`
	for i := range roles {
		r := &roles[i]
		r.RoleName = strings.ToUpper(strings.TrimSpace(r.RoleName))
		_, err := db.Exec(db.Rebind(query),
			r.PermID, r.CompanyName, r.StockName,
			r.DateAppointed, r.DateResigned,
			r.Category, r.RoleName,
			r.Promoted, r.Alternative, r.Independentcy, r.Executive, r.Chairmanship,
		)
		if err != nil {
			return fmt.Errorf("insert entity role perm_id=%d stock=%s: %w", r.PermID, r.StockName, err)
		}
	}
	return nil
}
