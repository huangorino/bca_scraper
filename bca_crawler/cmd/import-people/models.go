package main

// IPODetail represents data from ext_0_ipo_details.csv
type IPODetail struct {
	CompanyName string `db:"company_name"`
	FieldValue  string `db:"field_value"`
	FieldName   string `db:"field_name"`
}

// PeopleProfile represents data from ext_1_people_profiles.csv
type PeopleProfile struct {
	Salutation                      string `db:"salutation"`
	CompanyName                     string `db:"company_name"`
	YearOfQualification             string `db:"year_of_qualification"`
	PhotoPassportSize               string `db:"photo_passport_size"`
	YearOfBirth                     string `db:"year_of_birth"`
	FullLegalName                   string `db:"full_legal_name"`
	ProfessionalQualification       string `db:"professional_qualification"`
	RemarksNotes                    string `db:"remarks_notes"`
	BiographyExperience             string `db:"biography_experience"`
	AppointmentDateToBoard          string `db:"appointment_date_to_board"`
	SourcePageReference             string `db:"source_page_reference"`
	Nationality                     string `db:"nationality"`
	Gender                          string `db:"gender"`
	DisplayName                     string `db:"display_name"`
	OtherCompanyInterest            string `db:"other_company_interest"`
	DateOfSource                    string `db:"date_of_source"`
	Category                        string `db:"category"`
	OtherDirectorshipPublicListedCo string `db:"other_directorship_public_listed_co"`
	SourceOfData                    string `db:"source_of_data"`
	AcademicQualification           string `db:"academic_qualification"`
	AppointmentDateToPosition       string `db:"appointment_date_to_position"`
	University                      string `db:"university"`
	PositionRole                    string `db:"position_role"`
}

// CorporateDirectory represents data from ext_1a_corporate_directory.csv
type CorporateDirectory struct {
	SourceOfData        string `db:"source_of_data"`
	Designation         string `db:"designation"`
	CompanyName         string `db:"company_name"`
	Address             string `db:"address"`
	DateOfSource        string `db:"date_of_source"`
	Category            string `db:"category"`
	Name                string `db:"name"`
	SourcePageReference string `db:"source_page_reference"`
	Nationality         string `db:"nationality"`
	Directorship        string `db:"directorship"`
}

// CorporateInfo represents data from ext_2_corporate_info.csv
type CorporateInfo struct {
	FieldValue  string `db:"field_value"`
	Nan4        string `db:"nan_4"`
	Nan1        string `db:"nan_1"`
	CompanyName string `db:"company_name"`
	Nan         string `db:"nan"`
	Nan3        string `db:"nan_3"`
	FieldName   string `db:"field_name"`
	Nan2        string `db:"nan_2"`
	WinstarVal  string `db:"winstar_capital_berhad"`
}

// CompanySecretary represents data from ext_2a_cosec.csv
type CompanySecretary struct {
	SourceOfData            string `db:"source_of_data"`
	CompanyName             string `db:"company_name"`
	CompanySecretaryRegNo   string `db:"company_secretary_reg_no"`
	DateOfSource            string `db:"date_of_source"`
	CompanySecretaryAddress string `db:"company_secretary_address"`
	CompanySecretaryName    string `db:"company_secretary_name"`
	CompanySecretaryContact string `db:"company_secretary_contact"`
	SourcePageReference     string `db:"source_page_reference"`
}

// Adviser represents data from ext_2b_adviser.csv
type Adviser struct {
	SourceOfData            string `db:"source_of_data"`
	CompanyName             string `db:"company_name"`
	AdviserAddress          string `db:"adviser_address"`
	AdviserName             string `db:"adviser_name"`
	DateOfSource            string `db:"date_of_source"`
	SourcePageReference     string `db:"source_page_reference"`
	AdviserRegistrationNo   string `db:"adviser_registration_number_or_business_registration_number"`
	AdviserContact          string `db:"adviser_contact"`
	TypeOfAdviser           string `db:"type_of_adviser"`
}

// Subsidiary represents data from ext_3_subsidiaries_associates.csv
type Subsidiary struct {
	StatusOfOperation    string `db:"status_of_operation"`
	PlaceOfIncorporation string `db:"place_of_incorporation"`
	DateOfIncorporation  string `db:"date_of_incorporation"`
	SourceOfData         string `db:"source_of_data"`
	CompanyName          string `db:"company_name"`
	RegistrationNumber   string `db:"registration_number"`
	PrincipalActivity    string `db:"principal_activity"`
	DateOfCommencement   string `db:"date_of_commencement"`
	DateOfSource         string `db:"date_of_source"`
	SourcePageReference  string `db:"source_page_reference"`
	ParentCompany        string `db:"parent_company"`
	IssuedShareCapital   string `db:"issued_share_capital"`
}

// SubsidiaryShareholder represents data from ext_4_subsidiary_shareholders.csv
type SubsidiaryShareholder struct {
	SourceOfData           string `db:"source_of_data"`
	SubsidiaryName         string `db:"subsidiary_name"`
	CompanyName            string `db:"company_name"`
	Ownership              string `db:"ownership"`
	DateOfSource           string `db:"date_of_source"`
	NumberOfSharesOwned    string `db:"number_of_shares_owned"`
	SourcePageReference    string `db:"source_page_reference"`
	ShareholderName        string `db:"shareholder_name"`
	EffectiveEquityOwnership string `db:"effective_equity_ownership"`
}

// PropertyOwned represents data from ext_5_properties_owned.csv
type PropertyOwned struct {
	DescriptionExistingUse  string `db:"description_existing_use"`
	ValueEstimate           string `db:"value_estimate"`
	CompanyName             string `db:"company_name"`
	LandAreaSqFt            string `db:"land_area_sq_ft"`
	ValueAudited            string `db:"value_audited"`
	ExpiryOfLease           string `db:"expiry_of_lease"`
	DateOfValuation         string `db:"date_of_valuation"`
	TitleDetails            string `db:"title_details"`
	LandArea                string `db:"land_area"`
	No                      string `db:"no"`
	ExistingUse             string `db:"existing_use"`
	Tenure                  string `db:"tenure"`
	BuiltUpAreaSqFt         string `db:"built_up_area_sq_ft"`
	DateOfPurchase          string `db:"date_of_purchase"`
	SourcePageReference     string `db:"source_page_reference"`
	Encumbrances            string `db:"encumbrances"`
	DescriptionOfProperty   string `db:"description_of_property"`
	FloorArea               string `db:"floor_area"`
	ExpressConditions       string `db:"express_conditions"`
	AuditedNbvRm000         string `db:"audited_nbv_rm_000"`
	RestrictionInInterest   string `db:"restriction_in_interest"`
	DateOfSource            string `db:"date_of_source"`
	CompanyOwner            string `db:"company_owner"`
	TitleParcelNo           string `db:"title_parcel_no"`
	RegisteredOwnerFullName string `db:"registered_owner_full_name"`
	PostalAddress          string `db:"postal_address"`
	SourceOfData            string `db:"source_of_data"`
	Encumbrance             string `db:"encumbrance"`
	DateOfCcc               string `db:"date_of_ccc"`
	CategoryOfLandUse       string `db:"category_of_land_use"`
	CertificateOfCompletion string `db:"certificate_of_completion"`
}

// PropertyRented represents data from ext_6_properties_rented.csv
type PropertyRented struct {
	DescriptionExistingUse string `db:"description_existing_use"`
	CompanyName            string `db:"company_name"`
	LandAreaSqFt           string `db:"land_area_sq_ft"`
	InitialDateOfTenancy   string `db:"initial_date_of_tenancy"`
	Tenant                 string `db:"tenant"`
	LandArea               string `db:"land_area"`
	No                     string `db:"no"`
	RentalPerAnnumRm       string `db:"rental_per_annum_rm"`
	Notes                  string `db:"notes"`
	BuiltUpAreaSqFt        string `db:"built_up_area_sq_ft"`
	RentalPeriod           string `db:"rental_period"`
	SourcePageReference    string `db:"source_page_reference"`
	TenantFullName         string `db:"tenant_full_name"`
	RentalAmount           string `db:"rental_amount"`
	FloorArea              string `db:"floor_area"`
	ExpressConditions      string `db:"express_conditions"`
	DateOfSource           string `db:"date_of_source"`
	TitleParcelNo          string `db:"title_parcel_no"`
	PostalAddress          string `db:"postal_address"`
	SourceOfData           string `db:"source_of_data"`
	Landlord               string `db:"landlord"`
	DateOfCcc              string `db:"date_of_ccc"`
	PeriodOfTenancy        string `db:"period_of_tenancy"`
}

// Relationship represents data from ext_7_relationships.csv
type Relationship struct {
	SourceOfData        string `db:"source_of_data"`
	CompanyName         string `db:"company_name"`
	DateOfSource        string `db:"date_of_source"`
	PersonEntityA       string `db:"person_entity_a"`
	PersonEntityB       string `db:"person_entity_b"`
	RelationshipType    string `db:"relationship_type"`
	SourcePageReference string `db:"source_page_reference"`
	NatureDetails       string `db:"nature_details"`
}

// MajorPartner represents data from ext_8_major_supplier_customer.csv
type MajorPartner struct {
	ContactNumber          string `db:"contact_number"`
	CompanyName            string `db:"company_name"`
	Address                string `db:"address"`
	BusinessRegistration   string `db:"business_registration"`
	PurchaseValueFYE2022   string `db:"purchase_value_revenue_fye2022"`
	CountryOfSupplier      string `db:"country_of_supplier"`
	PurchaseValueFYE2021   string `db:"purchase_value_revenue_fye2021"`
	PurchaseValueFPE2024   string `db:"purchase_value_revenue_fpe2024"`
	ContactPerson          string `db:"contact_person"`
	SourcePageReference    string `db:"source_page_reference"`
	LengthOfRelationship   string `db:"length_of_relationship_years_as_at_lpd"`
	DateOfSource           string `db:"date_of_source"`
	Category               string `db:"category"`
	ItemsSupplyDescription string `db:"items_supply_description"`
	PurchaseValueFYE2023   string `db:"purchase_value_revenue_fye2023"`
	EmailContact           string `db:"email_contact"`
	SourceOfData           string `db:"source_of_data"`
	ItemsSupply            string `db:"items_supply"`
	Name                   string `db:"name"`
}

// DataStore is the central container for all ingested data
type DataStore struct {
	IPOs                []IPODetail
	People              []PeopleProfile
	CorporateDirectory  []CorporateDirectory
	CorporateInfo       []CorporateInfo
	CompanySecretaries  []CompanySecretary
	Advisers            []Adviser
	Subsidiaries        []Subsidiary
	SubShareholders     []SubsidiaryShareholder
	PropertiesOwned     []PropertyOwned
	PropertiesRented    []PropertyRented
	Relationships       []Relationship
	MajorPartners       []MajorPartner
}
