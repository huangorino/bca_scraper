package main

import (
	"fmt"

	"bca_crawler/internal/utils"

	"github.com/jmoiron/sqlx"
)

// Persist saves all data in the DataStore to the database
func (s *DataStore) Persist(db *sqlx.DB) error {
	log := utils.Logger

	// Define tasks for batch insertion
	tasks := []struct {
		name  string
		table string
		data  interface{}
	}{
		{"CoSec", "ext_cosec", s.CompanySecretaries},
		{"Advisers", "ext_adviser", s.Advisers},
		{"Subsidiaries", "ext_subsidiaries_associates", s.Subsidiaries},
		{"Sub Shareholders", "ext_subsidiary_shareholders", s.SubShareholders},
		{"Props Owned", "ext_properties_owned", s.PropertiesOwned},
		{"Props Rented", "ext_properties_rented", s.PropertiesRented},
		{"Relationships", "ext_relationships", s.Relationships},
		{"Major Partners", "ext_major_supplier_customer", s.MajorPartners},
	}

	for _, t := range tasks {
		if err := s.batchInsert(db, t.table, t.data); err != nil {
			log.Errorf("❌ Failed to insert into %s: %v", t.table, err)
			return err
		}
		log.Infof("💾 Successfully persisted %s to %s", t.name, t.table)
	}

	return nil
}

func (s *DataStore) batchInsert(db *sqlx.DB, table string, data interface{}) error {
	// Check if data is empty
	switch v := data.(type) {
	case []IPODetail:
		if len(v) == 0 {
			return nil
		}
	case []PeopleProfile:
		if len(v) == 0 {
			return nil
		}
	case []CorporateDirectory:
		if len(v) == 0 {
			return nil
		}
	case []CorporateInfo:
		if len(v) == 0 {
			return nil
		}
	case []CompanySecretary:
		if len(v) == 0 {
			return nil
		}
	case []Adviser:
		if len(v) == 0 {
			return nil
		}
	case []Subsidiary:
		if len(v) == 0 {
			return nil
		}
	case []SubsidiaryShareholder:
		if len(v) == 0 {
			return nil
		}
	case []PropertyOwned:
		if len(v) == 0 {
			return nil
		}
	case []PropertyRented:
		if len(v) == 0 {
			return nil
		}
	case []Relationship:
		if len(v) == 0 {
			return nil
		}
	case []MajorPartner:
		if len(v) == 0 {
			return nil
		}
	}

	// Construct dynamic query using NamedExec
	// sqlx will automatically map struct tags to column names
	cols := s.getColumnsForTable(table)
	if cols == "" {
		return fmt.Errorf("unknown table columns for %s", table)
	}

	query := fmt.Sprintf("INSERT INTO %s %s", table, cols)
	_, err := db.NamedExec(query, data)
	return err
}

func (s *DataStore) getColumnsForTable(table string) string {
	switch table {
	case "ext_ipo_details":
		return "(company_name, field_value, field_name) VALUES (:company_name, :field_value, :field_name)"
	case "ext_people_profiles":
		return `(salutation, company_name, year_of_qualification, photo_passport_size, year_of_birth, full_legal_name, 
		         professional_qualification, remarks_notes, biography_experience, appointment_date_to_board, 
		         source_page_reference, nationality, gender, display_name, other_company_interest, date_of_source, 
		         category, other_directorship_public_listed_co, source_of_data, academic_qualification, 
		         appointment_date_to_position, university, position_role) 
		        VALUES (:salutation, :company_name, :year_of_qualification, :photo_passport_size, :year_of_birth, :full_legal_name, 
		         :professional_qualification, :remarks_notes, :biography_experience, :appointment_date_to_board, 
		         :source_page_reference, :nationality, :gender, :display_name, :other_company_interest, :date_of_source, 
		         :category, :other_directorship_public_listed_co, :source_of_data, :academic_qualification, 
		         :appointment_date_to_position, :university, :position_role)`
	case "ext_corporate_directory":
		return `(source_of_data, designation, company_name, address, date_of_source, category, name, 
		         source_page_reference, nationality, directorship) 
		        VALUES (:source_of_data, :designation, :company_name, :address, :date_of_source, :category, :name, 
		         :source_page_reference, :nationality, :directorship)`
	case "ext_corporate_info":
		return `(field_value, nan_4, nan_1, company_name, nan, nan_3, field_name, nan_2, winstar_capital_berhad) 
		        VALUES (:field_value, :nan_4, :nan_1, :company_name, :nan, :nan_3, :field_name, :nan_2, :winstar_capital_berhad)`
	case "ext_cosec":
		return `(source_of_data, company_name, company_secretary_reg_no, date_of_source, company_secretary_address, 
		         company_secretary_name, company_secretary_contact, source_page_reference) 
		        VALUES (:source_of_data, :company_name, :company_secretary_reg_no, :date_of_source, :company_secretary_address, 
		         :company_secretary_name, :company_secretary_contact, :source_page_reference)`
	case "ext_adviser":
		return `(source_of_data, company_name, adviser_address, adviser_name, date_of_source, source_page_reference, 
		         adviser_registration_number_or_business_registration_number, adviser_contact, type_of_adviser) 
		        VALUES (:source_of_data, :company_name, :adviser_address, :adviser_name, :date_of_source, :source_page_reference, 
		         :adviser_registration_number_or_business_registration_number, :adviser_contact, :type_of_adviser)`
	case "ext_subsidiaries_associates":
		return `(status_of_operation, place_of_incorporation, date_of_incorporation, source_of_data, company_name, 
		         registration_number, principal_activity, date_of_commencement, date_of_source, source_page_reference, 
		         parent_company, issued_share_capital) 
		        VALUES (:status_of_operation, :place_of_incorporation, :date_of_incorporation, :source_of_data, :company_name, 
		         :registration_number, :principal_activity, :date_of_commencement, :date_of_source, :source_page_reference, 
		         :parent_company, :issued_share_capital)`
	case "ext_subsidiary_shareholders":
		return `(source_of_data, subsidiary_name, company_name, ownership, date_of_source, number_of_shares_owned, 
		         source_page_reference, shareholder_name, effective_equity_ownership) 
		        VALUES (:source_of_data, :subsidiary_name, :company_name, :ownership, :date_of_source, :number_of_shares_owned, 
		         :source_page_reference, :shareholder_name, :effective_equity_ownership)`
	case "ext_properties_owned":
		return `(description_existing_use, value_estimate, company_name, land_area_sq_ft, value_audited, expiry_of_lease, 
		         date_of_valuation, title_details, land_area, no, existing_use, tenure, built_up_area_sq_ft, 
		         date_of_purchase, source_page_reference, encumbrances, description_of_property, floor_area, 
		         express_conditions, audited_nbv_rm_000, restriction_in_interest, date_of_source, company_owner, 
		         title_parcel_no, registered_owner_full_name, postal_address, source_of_data, encumbrance, 
		         date_of_ccc, category_of_land_use, certificate_of_completion) 
		        VALUES (:description_existing_use, :value_estimate, :company_name, :land_area_sq_ft, :value_audited, :expiry_of_lease, 
		         :date_of_valuation, :title_details, :land_area, :no, :existing_use, :tenure, :built_up_area_sq_ft, 
		         :date_of_purchase, :source_page_reference, :encumbrances, :description_of_property, :floor_area, 
		         :express_conditions, :audited_nbv_rm_000, :restriction_in_interest, :date_of_source, :company_owner, 
		         :title_parcel_no, :registered_owner_full_name, :postal_address, :source_of_data, :encumbrance, 
		         :date_of_ccc, :category_of_land_use, :certificate_of_completion)`
	case "ext_properties_rented":
		return `(description_existing_use, company_name, land_area_sq_ft, initial_date_of_tenancy, tenant, land_area, 
		         no, rental_per_annum_rm, notes, built_up_area_sq_ft, rental_period, source_page_reference, 
		         tenant_full_name, rental_amount, floor_area, express_conditions, date_of_source, title_parcel_no, 
		         postal_address, source_of_data, landlord, date_of_ccc, period_of_tenancy) 
		        VALUES (:description_existing_use, :company_name, :land_area_sq_ft, :initial_date_of_tenancy, :tenant, :land_area, 
		         :no, :rental_per_annum_rm, :notes, :built_up_area_sq_ft, :rental_period, :source_page_reference, 
		         :tenant_full_name, :rental_amount, :floor_area, :express_conditions, :date_of_source, :title_parcel_no, 
		         :postal_address, :source_of_data, :landlord, :date_of_ccc, :period_of_tenancy)`
	case "ext_relationships":
		return `(source_of_data, company_name, date_of_source, person_entity_a, person_entity_b, relationship_type, 
		         source_page_reference, nature_details) 
		        VALUES (:source_of_data, :company_name, :date_of_source, :person_entity_a, :person_entity_b, :relationship_type, 
		         :source_page_reference, :nature_details)`
	case "ext_major_supplier_customer":
		return `(contact_number, company_name, address, business_registration, purchase_value_revenue_fye2022, 
		         country_of_supplier, purchase_value_revenue_fye2021, purchase_value_revenue_fpe2024, contact_person, 
		         source_page_reference, length_of_relationship_years_as_at_lpd, date_of_source, category, 
		         items_supply_description, purchase_value_revenue_fye2023, email_contact, source_of_data, items_supply, name) 
		        VALUES (:contact_number, :company_name, :address, :business_registration, :purchase_value_revenue_fye2022, 
		         :country_of_supplier, :purchase_value_revenue_fye2021, :purchase_value_revenue_fpe2024, :contact_person, 
		         :source_page_reference, :length_of_relationship_years_as_at_lpd, :date_of_source, :category, 
		         :items_supply_description, :purchase_value_revenue_fye2023, :email_contact, :source_of_data, :items_supply, :name)`
	}
	return ""
}
