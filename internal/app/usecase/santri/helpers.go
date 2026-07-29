package santriUsecase

import (
	"context"
	"fmt"
	"time"

	"sipon-api/internal/app/port"
	santriconstant "sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/entity"
)

func confirmDokumenMediaKey(ctx context.Context, fileUploader port.FileUploader, key string) {
	if err := fileUploader.ConfirmUpload(ctx, key); err != nil {
		_ = fmt.Errorf("confirm upload failed for key=%s: %w", key, err)
	}
}

func mapSantriToResponse(s *entity.Santri, fullname, email *string, phone, avatarURL *string) *GetSantriResponse {
	r := &GetSantriResponse{
		ID:       s.ID,
		UserID:   s.UserID,
		Fullname: fullname,
		Email:    strPtr(email),
		Phone:    phone,
		AvatarURL: avatarURL,

		Nickname:        s.Nickname,
		Program:         s.Program,
		Option:          s.Option,
		Hobby:           s.Hobby,
		Purpose:         s.Purpose,
		MotivationEntry: s.MotivationEntry,
		POB:             s.POB,
		Blood:           s.Blood,

		Address:     s.Address,
		SubDistrict: s.SubDistrict,
		District:    s.District,
		Province:    s.Province,
		PostalCode:  s.PostalCode,

		PreviousPondokName:    s.PreviousPondokName,
		PreviousPondokAddress: s.PreviousPondokAddress,
		PreviousPondokDiv:     s.PreviousPondokDiv,
		PreviousPondokTime:    s.PreviousPondokTime,

		NIK:   s.NIK,
		NoKK:  s.NoKK,
		NISN:  s.NISN,
		NoKIP: s.NoKIP,
		NoKKS: s.NoKKS,
		NoPKH: s.NoPKH,

		Workplace:  s.Workplace,
		Department: s.Department,

		HomeStatus:           s.HomeStatus,
		Father:               s.Father,
		FatherPN:             s.FatherPN,
		FatherNIK:            s.FatherNIK,
		FatherJob:            s.FatherJob,
		FatherGraduate:       s.FatherGraduate,
		FatherIncome:         s.FatherIncome,
		Mother:               s.Mother,
		MotherPN:             s.MotherPN,
		MotherNIK:            s.MotherNIK,
		MotherJob:            s.MotherJob,
		MotherGraduate:       s.MotherGraduate,
		MotherIncome:         s.MotherIncome,
		GuardianRelationship: s.GuardianRelationship,
		Guardian:             s.Guardian,
		GuardianPN:           s.GuardianPN,
		GuardianNIK:          s.GuardianNIK,
		GuardianJob:          s.GuardianJob,
		GuardianGraduate:     s.GuardianGraduate,
		GuardianIncome:       s.GuardianIncome,
	}

	if s.NIS != nil {
		v := s.NIS.Value()
		r.NIS = &v
	}

	if s.DOB != nil {
		dobStr := s.DOB.Format("2006-01-02")
		r.DOB = &dobStr
	}

	return r
}

func applySantriUpdate(s *entity.Santri, req UpdateSantriRequest) {
	if req.Nickname != nil            { s.Nickname = req.Nickname }
	if req.Hobby != nil               { s.Hobby = req.Hobby }
	if req.Purpose != nil             { s.Purpose = req.Purpose }
	if req.MotivationEntry != nil     { s.MotivationEntry = req.MotivationEntry }
	if req.POB != nil                 { s.POB = req.POB }
	if req.DOB != nil {
		t, err := time.Parse("2006-01-02", *req.DOB)
		if err == nil { s.DOB = &t }
	}
	if req.Blood != nil      { s.Blood = req.Blood }
	if req.Address != nil    { s.Address = req.Address }
	if req.SubDistrict != nil { s.SubDistrict = req.SubDistrict }
	if req.District != nil   { s.District = req.District }
	if req.Province != nil   { s.Province = req.Province }
	if req.PostalCode != nil { s.PostalCode = req.PostalCode }
	if req.PreviousPondokName != nil    { s.PreviousPondokName = req.PreviousPondokName }
	if req.PreviousPondokAddress != nil { s.PreviousPondokAddress = req.PreviousPondokAddress }
	if req.PreviousPondokDiv != nil     { s.PreviousPondokDiv = req.PreviousPondokDiv }
	if req.PreviousPondokTime != nil    { s.PreviousPondokTime = req.PreviousPondokTime }
	if req.NIK != nil   { s.NIK = req.NIK }
	if req.NoKK != nil  { s.NoKK = req.NoKK }
	if req.NISN != nil  { s.NISN = req.NISN }
	if req.NoKIP != nil { s.NoKIP = req.NoKIP }
	if req.NoKKS != nil { s.NoKKS = req.NoKKS }
	if req.NoPKH != nil { s.NoPKH = req.NoPKH }
	if req.Workplace != nil  { s.Workplace = req.Workplace }
	if req.Department != nil { s.Department = req.Department }
	if req.HomeStatus != nil          { s.HomeStatus = req.HomeStatus }
	if req.Father != nil              { s.Father = req.Father }
	if req.FatherPN != nil            { s.FatherPN = req.FatherPN }
	if req.FatherNIK != nil           { s.FatherNIK = req.FatherNIK }
	if req.FatherJob != nil           { s.FatherJob = req.FatherJob }
	if req.FatherGraduate != nil      { s.FatherGraduate = req.FatherGraduate }
	if req.FatherIncome != nil        { s.FatherIncome = req.FatherIncome }
	if req.Mother != nil              { s.Mother = req.Mother }
	if req.MotherPN != nil            { s.MotherPN = req.MotherPN }
	if req.MotherNIK != nil           { s.MotherNIK = req.MotherNIK }
	if req.MotherJob != nil           { s.MotherJob = req.MotherJob }
	if req.MotherGraduate != nil      { s.MotherGraduate = req.MotherGraduate }
	if req.MotherIncome != nil        { s.MotherIncome = req.MotherIncome }
	if req.GuardianRelationship != nil { s.GuardianRelationship = req.GuardianRelationship }
	if req.Guardian != nil             { s.Guardian = req.Guardian }
	if req.GuardianPN != nil           { s.GuardianPN = req.GuardianPN }
	if req.GuardianNIK != nil          { s.GuardianNIK = req.GuardianNIK }
	if req.GuardianJob != nil          { s.GuardianJob = req.GuardianJob }
	if req.GuardianGraduate != nil     { s.GuardianGraduate = req.GuardianGraduate }
	if req.GuardianIncome != nil       { s.GuardianIncome = req.GuardianIncome }
}

func mapDokumenToItem(d *entity.SantriDokumen) DokumenItem {
	return DokumenItem{
		ID:               d.ID,
		Kind:             string(d.Kind),
		Key:              d.Key,
		Status:           string(d.Status),
		OriginalFilename: d.OriginalFilename,
		MimeType:         d.MimeType,
		Size:             d.Size,
		Notes:            d.Notes,
		VerifiedBy:       d.VerifiedBy,
		VerifiedAt:       d.VerifiedAt,
		CreatedAt:        d.CreatedAt,
	}
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseDokumenKind(kind string) (santriconstant.DokumenKind, error) {
	k := santriconstant.DokumenKind(kind)
	if !santriconstant.ValidDokumenKinds[k] {
		return "", fmt.Errorf("invalid dokumen kind: %s", kind)
	}
	return k, nil
}
