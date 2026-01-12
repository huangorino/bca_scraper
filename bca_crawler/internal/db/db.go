package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"bca_crawler/internal/models"
)

func FetchMissingAnnID(db *sqlx.DB) ([]int64, error) {
	rows, err := db.Query(`
		SELECT t1.ann_id + 1 AS missing_ann_id
		FROM announcements t1
		LEFT JOIN announcements t2
			ON t2.ann_id = t1.ann_id + 1
		WHERE t2.ann_id IS NULL
		ORDER BY missing_ann_id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query missing ann_id gaps: %w", err)
	}
	defer rows.Close()

	var missing []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan missing ann_id: %w", err)
		}
		missing = append(missing, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return missing, nil
}

func FetchUnparsedAnnouncements(db *sqlx.DB) ([]*models.Announcement, error) {
	rows, err := db.Query(`
	SELECT id, ann_id, content 
	FROM announcements 
	WHERE ref_number = '' ORDER BY ann_id ASC`)

	if err != nil {
		return nil, fmt.Errorf("query announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*models.Announcement
	for rows.Next() {
		var ann models.Announcement
		if err := rows.Scan(&ann.ID, &ann.AnnID, &ann.Content); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		announcements = append(announcements, &ann)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return announcements, nil
}

func GetMaxAnnID(db *sqlx.DB) (int, error) {
	var maxID int
	err := db.QueryRow(`SELECT MAX(ann_id) FROM announcements`).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("query max ann_id: %w", err)
	}
	return maxID, nil
}

func FetchAnnouncementsByCategory(db *sqlx.DB, category string) ([]*models.Announcement, error) {
	sqlQuery := `SELECT id, ann_id,
      link, company_name, stock_name,
      date_posted, category, ref_number,
      attachments, content 
      FROM announcements`

	var args []interface{}
	if category != "" {
		if category == "attachments" {
			sqlQuery += " WHERE attachments != 'null' AND date_posted >= CURRENT_DATE - INTERVAL '7 days'"
		} else {
			sqlQuery += " WHERE category = $1 AND date_posted >= CURRENT_DATE - INTERVAL '7 days'"
			args = append(args, category)
		}
	}

	sqlQuery += " ORDER BY ann_id ASC"

	var announcements []models.AnnouncementDB
	err := db.Select(&announcements, sqlQuery, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx select: %w", err)
	}

	var result []*models.Announcement
	for i := range announcements {
		ann, err := ConvertAnnouncementDBToAnnouncement(announcements[i])
		if err != nil {
			return nil, fmt.Errorf("convert announcement db to announcement: %w", err)
		}

		result = append(result, ann)
	}

	return result, nil
}

func ConvertAnnouncementDBToAnnouncement(a models.AnnouncementDB) (*models.Announcement, error) {
	attachments := []string{}
	if err := json.Unmarshal([]byte(a.Attachments.String), &attachments); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &models.Announcement{
		ID:          a.ID,
		AnnID:       a.AnnID,
		Title:       a.Title.String,
		Link:        a.Link.String,
		CompanyName: a.CompanyName.String,
		StockName:   a.StockName.String,
		DatePosted:  a.DatePosted.Time,
		Category:    a.Category.String,
		RefNumber:   a.RefNumber.String,
		Content:     a.Content.String,
		Attachments: attachments,
	}, nil
}

// FindEntitiesByNameOrDisplay finds all entities matching the given name or display_name
func FindEntitiesByNameOrDisplay(db *sqlx.DB, name string, displayName string) ([]models.Entity, error) {
	var entities []models.Entity
	err := db.Select(&entities, `
		SELECT id, primary_perm_id, secondary_perm_id, display_name, name, salutation, 
		       stock_code, birth_year, gender, nationality, created_at, updated_at
		FROM entities
		WHERE name = $1 OR display_name = $2
		ORDER BY secondary_perm_id ASC`, name, displayName)
	if err != nil {
		return nil, fmt.Errorf("query entities: %w", err)
	}
	return entities, nil
}

// UpdatePrimaryPermID updates all entities matching name or display_name to set their primary_perm_id
func UpdatePrimaryPermID(db *sqlx.DB, name string, displayName string, primaryPermID int) error {
	_, err := db.Exec(`
		UPDATE entities
		SET primary_perm_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE name = $2 OR display_name = $3`, primaryPermID, name, displayName)
	if err != nil {
		return fmt.Errorf("update primary_perm_id: %w", err)
	}
	return nil
}
