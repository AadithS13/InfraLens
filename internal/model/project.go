package model

import (
	"encoding/json"
	"time"
)

// API response wrappers

type APIResponse[T any] struct {
	Message        string `json:"message"`
	Status         string `json:"status"`
	ResponseObject T      `json:"responseObject"`
}

// Auth

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expires_in"`
}

// Project general details — maps to getProjectGeneralDetailsByProjectId

type ProjectGeneral struct {
	ProjectID                         int     `json:"projectId"`
	ProjectName                       string  `json:"projectName"`
	ProjectRegistrationNo             string  `json:"projectRegistartionNo"`
	AcknowledgementNumber             string  `json:"acknowledgementNumber"`
	ProjectTypeName                   string  `json:"projectTypeName"`
	ProjectStatusName                 string  `json:"projectStatusName"`
	ProjectCurrentStatus              string  `json:"projectCurrentStatus"`
	ReraRegistrationDate              string  `json:"reraRegistrationDate"`
	ProjectProposeComplitionDate      string  `json:"projectProposeComplitionDate"`
	OriginalProjectProposeCompletionDate string `json:"originalProjectProposeCompletionDate"`
	TotalNumberOfUnits                int     `json:"totalNumberOfUnits"`
	TotalNumberOfSoldUnits            int     `json:"totalNumberOfSoldUnits"`
	UserProfileID                     int     `json:"userProfileId"`
	IsMigrated                        int     `json:"isMigrated"`
	IsProjectLapsed                   int     `json:"isProjectLapsed"`
	ProjectFeesPayableAmount          float64 `json:"projectFeesPayableAmount"`
}

// getProjectAndAssociatedPromoterDetails returns a nested object

type PromoterAssociationResponse struct {
	PromoterDetails PromoterAssociation `json:"promoterDetails"`
}

type PromoterAssociation struct {
	PromoterName      string `json:"promoterName"`
	ContactPersonName string `json:"contactPersonName"`
	BuildingName      string `json:"buildingName"`
	Pincode           int    `json:"pincode"`
	StateName         string `json:"stateName"`
	DistrictName      string `json:"districtName"`
	CityName          string `json:"cityName"`
	MobileNo          string `json:"mobileNo"`
	EmailID           string `json:"emailId"`
}

// Promoter general details — maps to fetchPromoterGeneralDetails

type PromoterGeneral struct {
	UserProfileID   int    `json:"userProfileId"`
	PromoterName    string `json:"promoterName"`
	Pan             string `json:"pan"`
	Gstin           string `json:"gstin"`
	PromoterType    string `json:"promoterType"`
	MobileNumber    string `json:"mobileNumber"`
	EmailID         string `json:"emailId"`
}

// ProjectAddress — maps to getProjectLandAddressDetails

type ProjectAddress struct {
	AddressLine  string `json:"addressLine"`
	Street       string `json:"street"`
	Locality     string `json:"locality"`
	DistrictName string `json:"districtName"`
	StateName    string `json:"stateName"`
	PinCode      string `json:"pinCode"`
	TalukaName   string `json:"talukaName"`
	VillageName  string `json:"villageName"`
}

// PromoterAddress — maps to getPromoterAddressDetails

type PromoterAddress struct {
	UserProfileID int    `json:"userProfileId"`
	BuildingName  string `json:"buildingName"`
	UnitNumber    string `json:"unitNumber"`
	StreetName    string `json:"streetName"`
	Locality      string `json:"locality"`
	Landmark      string `json:"landmark"`
	DistrictName  string `json:"districtName"`
	StateName     string `json:"stateName"`
	PinCode       string `json:"pinCode"`
	TalukaName    string `json:"talukaName"`
	VillageName   string `json:"villageName"`
}

// Professional — maps to getProjectProfessionalByType

type Professional struct {
	ProfessionalName string `json:"professionalName"`
	ProfessionalType string `json:"professionalTypeName"`
	MobileNumber     string `json:"mobileNumber"`
	EmailID          string `json:"emailId"`
}

// Agent — maps to getAgentByProjectId

type Agent struct {
	AgentName    string `json:"agentName"`
	ReraRegNo    string `json:"reraRegNo"`
	MobileNumber string `json:"mobileNumber"`
	EmailID      string `json:"emailId"`
}

// Contact personnel — maps to fetchPromoterPersonnelContactAddressDetails

type PersonnelContact struct {
	Name         string `json:"personName"`
	Designation  string `json:"designation"`
	MobileNumber string `json:"mobileNumber"`
	EmailID      string `json:"emailId"`
}

// DB models

type Project struct {
	ID                    int
	MahaID                int
	ReraRegistrationNo    string
	AcknowledgementNumber string
	ProjectName           string
	ProjectType           string
	ProjectStatus         string
	ProjectCurrentStatus  string
	ReraRegistrationDate  *time.Time
	ProposedCompletionDate *time.Time
	OriginalCompletionDate *time.Time
	TotalUnits            int
	TotalSoldUnits        int
	UserProfileID         int
	PromoterID            *int
	IsMigrated            bool
	RawJSON               json.RawMessage
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type Promoter struct {
	ID            int
	UserProfileID int
	Name          string
	Pan           string
	Gstin         string
	PromoterType  string
	RawJSON       json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Address struct {
	ID         int
	EntityType string
	EntityID   int
	Line1      string
	Line2      string
	City       string
	District   string
	State      string
	Pincode    string
	RawJSON    json.RawMessage
	CreatedAt  time.Time
}

type Contact struct {
	ID        int
	ProjectID int
	Role      string
	Name      string
	Phone     string
	Email     string
	RawJSON   json.RawMessage
	CreatedAt time.Time
}

type ProjectSnapshot struct {
	ID        int
	ProjectID int
	FetchedAt time.Time
	Checksum  string
	RawJSON   json.RawMessage
}

type ProjectChange struct {
	ProjectID  int
	FieldName  string
	OldValue   string
	NewValue   string
	DetectedAt time.Time
}
