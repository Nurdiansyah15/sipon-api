package santriUsecase

import (
	santrirepo "sipon-api/internal/domain/santri/repository"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/app/port"
)

type Dependencies struct {
	SantriRepo         santrirepo.SantriRepository
	SantriDokumenRepo  santrirepo.SantriDokumenRepository
	SantriRequestRepo  santrirepo.SantriRequestRepository
	UserRepo           userrepo.UserRepository
	FileUploader       port.FileUploader
	Hasher             port.PasswordHasher
	Transactor         port.Transactor
}

type UseCases struct {
	GetSantri           *GetSantriUseCase
	UpdateSantri        *UpdateSantriUseCase
	CreateSantri        *CreateSantriUseCase
	ListSantri          *ListSantriUseCase
	RequestSantri       *RequestSantriUseCase
	ListSantriRequests  *ListSantriRequestsUseCase
	ApproveSantriRequest *ApproveSantriRequestUseCase
	RejectSantriRequest  *RejectSantriRequestUseCase
	DokumenPresign      *DokumenPresignUseCase
	DokumenConfirm      *DokumenConfirmUseCase
	DokumenList         *DokumenListUseCase
	DokumenAccess       *DokumenAccessUseCase
	DokumenDelete       *DokumenDeleteUseCase
	DokumenVerify       *DokumenVerifyUseCase
	DokumenReject       *DokumenRejectUseCase
}

func NewUseCases(deps Dependencies) *UseCases {
	return &UseCases{
		GetSantri:            NewGetSantriUseCase(deps.SantriRepo, deps.UserRepo, deps.FileUploader),
		UpdateSantri:         NewUpdateSantriUseCase(deps.SantriRepo, deps.UserRepo),
		CreateSantri:         NewCreateSantriUseCase(deps.UserRepo, deps.SantriRepo, deps.Hasher),
		ListSantri:           NewListSantriUseCase(deps.SantriRepo, deps.UserRepo),
		RequestSantri:        NewRequestSantriUseCase(deps.SantriRepo, deps.SantriRequestRepo),
		ListSantriRequests:   NewListSantriRequestsUseCase(deps.SantriRequestRepo, deps.UserRepo),
		ApproveSantriRequest: NewApproveSantriRequestUseCase(deps.SantriRepo, deps.SantriRequestRepo, deps.UserRepo),
		RejectSantriRequest:  NewRejectSantriRequestUseCase(deps.SantriRequestRepo),
		DokumenPresign:       NewDokumenPresignUseCase(deps.FileUploader),
		DokumenConfirm:       NewDokumenConfirmUseCase(deps.SantriRepo, deps.SantriDokumenRepo, deps.FileUploader),
		DokumenList:          NewDokumenListUseCase(deps.SantriRepo, deps.SantriDokumenRepo),
		DokumenAccess:        NewDokumenAccessUseCase(deps.SantriRepo, deps.SantriDokumenRepo, deps.FileUploader),
		DokumenDelete:        NewDokumenDeleteUseCase(deps.SantriRepo, deps.SantriDokumenRepo, deps.FileUploader),
		DokumenVerify:        NewDokumenVerifyUseCase(deps.SantriDokumenRepo),
		DokumenReject:        NewDokumenRejectUseCase(deps.SantriDokumenRepo),
	}
}
