package models

type Background struct {
	ID                   int    `json:"id,omitempty"`
	PermID               int    `json:"perm_id,omitempty"`
	Qualification        string `json:"qualification,omitempty"`
	WorkingExperience    string `json:"working_experience,omitempty"`
	Directorships        string `json:"directorships,omitempty"`
	FamilyRelationship   string `json:"family_relationship,omitempty"`
	ConflictOfInterest   string `json:"conflict_of_interest,omitempty"`
	InterestInSecurities string `json:"interest_in_securities,omitempty"`
}
