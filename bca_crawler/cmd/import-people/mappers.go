package main

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"bca_crawler/internal/utils"
)

// Ingest handles the sequential loading of all CSV files in the provided directory
func (s *DataStore) Ingest(dir string) error {
	log := utils.Logger
	load := func(filename string, mapper func([]string) error) error {
		path := filepath.Join(dir, filename)
		return s.readCSV(path, mapper)
	}

	loaders := []struct {
		file string
		fn   func([]string) error
	}{
		{"ext_0_ipo_details.csv", s.mapIPODetail},
		{"ext_1_people_profiles.csv", s.mapPeopleProfile},
		{"ext_1a_corporate_directory.csv", s.mapCorporateDirectory},
		{"ext_2_corporate_info.csv", s.mapCorporateInfo},
		{"ext_2a_cosec.csv", s.mapCompanySecretary},
		{"ext_2b_adviser.csv", s.mapAdviser},
		{"ext_3_subsidiaries_associates.csv", s.mapSubsidiary},
		{"ext_4_subsidiary_shareholders.csv", s.mapSubsidiaryShareholder},
		{"ext_5_properties_owned.csv", s.mapPropertyOwned},
		{"ext_6_properties_rented.csv", s.mapPropertyRented},
		{"ext_7_relationships.csv", s.mapRelationship},
		{"ext_8_major_supplier_customer.csv", s.mapMajorPartner},
	}

	for _, loader := range loaders {
		if err := load(loader.file, loader.fn); err != nil {
			log.Warnf("⚠️ Could not load %s: %v", loader.file, err)
		} else {
			log.Infof("✅ Ingested %s", loader.file)
		}
	}

	return nil
}

func (s *DataStore) readCSV(path string, mapper func([]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	if _, err := reader.Read(); err != nil {
		return err
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if err := mapper(row); err != nil {
			utils.Logger.Debugf("Skipping row in %s: %v", filepath.Base(path), err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Row Mapping Methods
// -----------------------------------------------------------------------------

func (s *DataStore) mapIPODetail(row []string) error {
	s.IPOs = append(s.IPOs, IPODetail{
		CompanyName: rowAt(row, 0),
		FieldValue:  rowAt(row, 1),
		FieldName:   rowAt(row, 2),
	})
	return nil
}

func (s *DataStore) mapPeopleProfile(row []string) error {
	s.People = append(s.People, PeopleProfile{
		Salutation:                      rowAt(row, 0),
		CompanyName:                     rowAt(row, 1),
		YearOfQualification:             rowAt(row, 2),
		PhotoPassportSize:               rowAt(row, 3),
		YearOfBirth:                     rowAt(row, 4),
		FullLegalName:                   rowAt(row, 5),
		ProfessionalQualification:       rowAt(row, 6),
		RemarksNotes:                    rowAt(row, 7),
		BiographyExperience:             rowAt(row, 8),
		AppointmentDateToBoard:          rowAt(row, 9),
		SourcePageReference:             rowAt(row, 10),
		Nationality:                     rowAt(row, 11),
		Gender:                          rowAt(row, 12),
		DisplayName:                     rowAt(row, 13),
		OtherCompanyInterest:            rowAt(row, 14),
		DateOfSource:                    rowAt(row, 15),
		Category:                        rowAt(row, 16),
		OtherDirectorshipPublicListedCo: rowAt(row, 17),
		SourceOfData:                    rowAt(row, 18),
		AcademicQualification:           rowAt(row, 19),
		AppointmentDateToPosition:       rowAt(row, 20),
		University:                      rowAt(row, 21),
		PositionRole:                    rowAt(row, 22),
	})
	return nil
}

func (s *DataStore) mapCorporateDirectory(row []string) error {
	s.CorporateDirectory = append(s.CorporateDirectory, CorporateDirectory{
		SourceOfData:        rowAt(row, 0),
		Designation:         rowAt(row, 1),
		CompanyName:         rowAt(row, 2),
		Address:             rowAt(row, 3),
		DateOfSource:        rowAt(row, 4),
		Category:            rowAt(row, 5),
		Name:                rowAt(row, 6),
		SourcePageReference: rowAt(row, 7),
		Nationality:         rowAt(row, 8),
		Directorship:        rowAt(row, 9),
	})
	return nil
}

func (s *DataStore) mapCorporateInfo(row []string) error {
	s.CorporateInfo = append(s.CorporateInfo, CorporateInfo{
		FieldValue:  rowAt(row, 0),
		Nan4:        rowAt(row, 1),
		Nan1:        rowAt(row, 2),
		CompanyName: rowAt(row, 3),
		Nan:         rowAt(row, 4),
		Nan3:        rowAt(row, 5),
		FieldName:   rowAt(row, 6),
		Nan2:        rowAt(row, 7),
		WinstarVal:  rowAt(row, 8),
	})
	return nil
}

func (s *DataStore) mapCompanySecretary(row []string) error {
	s.CompanySecretaries = append(s.CompanySecretaries, CompanySecretary{
		SourceOfData:            rowAt(row, 0),
		CompanyName:             rowAt(row, 1),
		CompanySecretaryRegNo:   rowAt(row, 2),
		DateOfSource:            rowAt(row, 3),
		CompanySecretaryAddress: rowAt(row, 4),
		CompanySecretaryName:    rowAt(row, 5),
		CompanySecretaryContact: rowAt(row, 6),
		SourcePageReference:     rowAt(row, 7),
	})
	return nil
}

func (s *DataStore) mapAdviser(row []string) error {
	s.Advisers = append(s.Advisers, Adviser{
		SourceOfData:          rowAt(row, 0),
		CompanyName:           rowAt(row, 1),
		AdviserAddress:        rowAt(row, 2),
		AdviserName:           rowAt(row, 3),
		DateOfSource:          rowAt(row, 4),
		SourcePageReference:   rowAt(row, 5),
		AdviserRegistrationNo: rowAt(row, 6),
		AdviserContact:        rowAt(row, 7),
		TypeOfAdviser:         rowAt(row, 8),
	})
	return nil
}

func (s *DataStore) mapSubsidiary(row []string) error {
	s.Subsidiaries = append(s.Subsidiaries, Subsidiary{
		StatusOfOperation:    rowAt(row, 0),
		PlaceOfIncorporation: rowAt(row, 1),
		DateOfIncorporation:  rowAt(row, 2),
		SourceOfData:         rowAt(row, 3),
		CompanyName:          rowAt(row, 4),
		RegistrationNumber:   rowAt(row, 5),
		PrincipalActivity:    rowAt(row, 6),
		DateOfCommencement:   rowAt(row, 7),
		DateOfSource:         rowAt(row, 8),
		SourcePageReference:  rowAt(row, 9),
		ParentCompany:        rowAt(row, 10),
		IssuedShareCapital:   rowAt(row, 11),
	})
	return nil
}

func (s *DataStore) mapSubsidiaryShareholder(row []string) error {
	s.SubShareholders = append(s.SubShareholders, SubsidiaryShareholder{
		SourceOfData:             rowAt(row, 0),
		SubsidiaryName:           rowAt(row, 1),
		CompanyName:              rowAt(row, 2),
		Ownership:                rowAt(row, 3),
		DateOfSource:             rowAt(row, 4),
		NumberOfSharesOwned:      rowAt(row, 5),
		SourcePageReference:      rowAt(row, 6),
		ShareholderName:          rowAt(row, 7),
		EffectiveEquityOwnership: rowAt(row, 8),
	})
	return nil
}

func (s *DataStore) mapPropertyOwned(row []string) error {
	s.PropertiesOwned = append(s.PropertiesOwned, PropertyOwned{
		DescriptionExistingUse:  rowAt(row, 0),
		ValueEstimate:           rowAt(row, 1),
		CompanyName:             rowAt(row, 2),
		LandAreaSqFt:            rowAt(row, 3),
		ValueAudited:            rowAt(row, 4),
		ExpiryOfLease:           rowAt(row, 5),
		DateOfValuation:         rowAt(row, 6),
		TitleDetails:            rowAt(row, 7),
		LandArea:                rowAt(row, 8),
		No:                      rowAt(row, 9),
		ExistingUse:             rowAt(row, 10),
		Tenure:                  rowAt(row, 11),
		BuiltUpAreaSqFt:         rowAt(row, 12),
		DateOfPurchase:          rowAt(row, 13),
		SourcePageReference:     rowAt(row, 14),
		Encumbrances:            rowAt(row, 15),
		DescriptionOfProperty:   rowAt(row, 16),
		FloorArea:               rowAt(row, 17),
		ExpressConditions:       rowAt(row, 18),
		AuditedNbvRm000:         rowAt(row, 19),
		RestrictionInInterest:   rowAt(row, 20),
		DateOfSource:            rowAt(row, 21),
		CompanyOwner:            rowAt(row, 22),
		TitleParcelNo:           rowAt(row, 23),
		RegisteredOwnerFullName: rowAt(row, 24),
		PostalAddress:           rowAt(row, 25),
		SourceOfData:            rowAt(row, 26),
		Encumbrance:             rowAt(row, 27),
		DateOfCcc:               rowAt(row, 28),
		CategoryOfLandUse:       rowAt(row, 29),
		CertificateOfCompletion: rowAt(row, 30),
	})
	return nil
}

func (s *DataStore) mapPropertyRented(row []string) error {
	s.PropertiesRented = append(s.PropertiesRented, PropertyRented{
		DescriptionExistingUse: rowAt(row, 0),
		CompanyName:            rowAt(row, 1),
		LandAreaSqFt:           rowAt(row, 2),
		InitialDateOfTenancy:   rowAt(row, 3),
		Tenant:                 rowAt(row, 4),
		LandArea:               rowAt(row, 5),
		No:                     rowAt(row, 6),
		RentalPerAnnumRm:       rowAt(row, 7),
		Notes:                  rowAt(row, 8),
		BuiltUpAreaSqFt:        rowAt(row, 9),
		RentalPeriod:           rowAt(row, 10),
		SourcePageReference:    rowAt(row, 11),
		TenantFullName:         rowAt(row, 12),
		RentalAmount:           rowAt(row, 13),
		FloorArea:              rowAt(row, 14),
		ExpressConditions:      rowAt(row, 15),
		DateOfSource:           rowAt(row, 16),
		TitleParcelNo:          rowAt(row, 17),
		PostalAddress:          rowAt(row, 18),
		SourceOfData:           rowAt(row, 19),
		Landlord:               rowAt(row, 20),
		DateOfCcc:              rowAt(row, 21),
		PeriodOfTenancy:        rowAt(row, 22),
	})
	return nil
}

func (s *DataStore) mapRelationship(row []string) error {
	s.Relationships = append(s.Relationships, Relationship{
		SourceOfData:        rowAt(row, 0),
		CompanyName:         rowAt(row, 1),
		DateOfSource:        rowAt(row, 2),
		PersonEntityA:       rowAt(row, 3),
		PersonEntityB:       rowAt(row, 4),
		RelationshipType:    rowAt(row, 5),
		SourcePageReference: rowAt(row, 6),
		NatureDetails:       rowAt(row, 7),
	})
	return nil
}

func (s *DataStore) mapMajorPartner(row []string) error {
	s.MajorPartners = append(s.MajorPartners, MajorPartner{
		ContactNumber:          rowAt(row, 0),
		CompanyName:            rowAt(row, 1),
		Address:                rowAt(row, 2),
		BusinessRegistration:   rowAt(row, 3),
		PurchaseValueFYE2022:   rowAt(row, 4),
		CountryOfSupplier:      rowAt(row, 5),
		PurchaseValueFYE2021:   rowAt(row, 6),
		PurchaseValueFPE2024:   rowAt(row, 7),
		ContactPerson:          rowAt(row, 8),
		SourcePageReference:    rowAt(row, 9),
		LengthOfRelationship:   rowAt(row, 10),
		DateOfSource:           rowAt(row, 11),
		Category:               rowAt(row, 12),
		ItemsSupplyDescription: rowAt(row, 13),
		PurchaseValueFYE2023:   rowAt(row, 14),
		EmailContact:           rowAt(row, 15),
		SourceOfData:           rowAt(row, 16),
		ItemsSupply:            rowAt(row, 17),
		Name:                   rowAt(row, 18),
	})
	return nil
}

// rowAt provides bounds-checked access to CSV row columns
func rowAt(row []string, index int) string {
	if index >= 0 && index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}
