package services

import (
	"fmt"

	"github.com/KEDigitalMY/kedai_models/db"
	"github.com/KEDigitalMY/kedai_models/models"
	"github.com/KEDigitalMY/kedai_models/utils"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func GetOrCreateEntity(log *logrus.Logger, database *sqlx.DB, entity *models.Entity, background *models.Background) (*int, error) {
	// Step 1: Check if db contains records with the name/display_name
	entities, err := db.FindEntitiesByNameOrDisplay(database, *entity.Name, *entity.OriName)
	if err != nil {
		return nil, fmt.Errorf("FindEntitiesByNameOrDisplay failed: %w", err)
	}

	var permID int
	if len(entities) > 0 {
		log.Infof("Found entities: %s", utils.StringValue(entities[0].Name))

		for _, perm := range entities {
			permID = entities[0].SecondaryPermID

			err = db.UpdatePrimaryPermID(database, perm.ID, permID)
			if err != nil {
				return nil, fmt.Errorf("UpdatePrimaryPermID failed: %w", err)
			}
		}
	} else {
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

func ProcessSingleRoleChange(input models.RoleChangeInput, tracker map[string]*models.EntityRole) []*models.EntityRole {
	var rolesCreated []*models.EntityRole

	// Helper to create and track a new role record
	createRole := func(asAppointed bool, roleName string) {
		role := &models.EntityRole{
			PermID:      input.PermID,
			CompanyName: input.CompanyName,
			StockName:   input.StockCode,
			RoleName:    roleName,
			Category:    input.Category,
		}
		if asAppointed {
			role.DateAppointed = input.DateOfChange
			tracker[input.StockCode] = role
		} else {
			role.DateResigned = input.DateOfChange
		}
		rolesCreated = append(rolesCreated, role)
	}

	switch input.TypeOfChange {
	case "APPOINTMENT", "REDESIGNATION":
		if active, ok := tracker[input.StockCode]; ok && input.TypeOfChange == "REDESIGNATION" {
			active.DateResigned = input.DateOfChange
		}
		createRole(true, input.Designation)

	case "RESIGNATION", "RETIREMENT", "REMOVED", "DEMISED", "OTHERS", "CESSATION", "CESSATION OF OFFICE", "VACATION OF OFFICE":
		if active, ok := tracker[input.StockCode]; ok {
			active.DateResigned = input.DateOfChange
			delete(tracker, input.StockCode)
		} else {
			// Standalone cessation record: use PreviousPosition for role name if available
			roleName := utils.FirstNonEmpty(input.PreviousPosition, input.Designation)
			createRole(false, roleName)
		}
	}

	return rolesCreated
}
