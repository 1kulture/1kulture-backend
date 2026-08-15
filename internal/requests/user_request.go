package requests

type UpdateProfileRequest struct {
	FirstName   string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName    string `json:"last_name" validate:"omitempty,min=2,max=100"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=10,max=20"`
	AvatarURL   string `json:"avatar_url" validate:"omitempty,url,max=500"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=vendor event_manager"`
}

type KYCRequest struct {
	RoleID            string `json:"role_id" validate:"required,uuid"`
	DocumentType      string `json:"document_type" validate:"required,oneof=national_id passport drivers_license business_registration"`
	DocumentNumber    string `json:"document_number" validate:"required,min=5,max=100"`
	DocumentURL       string `json:"document_url" validate:"required,url,max=500"`
	SelfieURL         string `json:"selfie_url" validate:"required,url,max=500"`
	AddressProofURL   string `json:"address_proof_url" validate:"required,url,max=500"`
	BusinessName      string `json:"business_name" validate:"omitempty,min=2,max=255"`
	BusinessRegNumber string `json:"business_reg_number" validate:"omitempty,min=5,max=100"`
}
