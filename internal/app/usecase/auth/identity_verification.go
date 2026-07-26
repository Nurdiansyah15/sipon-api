package authUsecase

import (
	constant "sipon-api/internal/domain/user/constant"
	verificationConstant "sipon-api/internal/domain/verification/constant"
)

func verificationPurposeFromIdentifier(kind constant.LoginIdentifierKind) verificationConstant.CodePurpose {
	switch kind {
	case constant.LoginIdentifierEmail:
		return verificationConstant.PurposeEmailVerification
	case constant.LoginIdentifierPhone:
		return verificationConstant.PurposePhoneVerification
	default:
		return ""
	}
}

func changeIdentityPurposeFromKind(kind constant.LoginIdentifierKind) verificationConstant.CodePurpose {
	switch kind {
	case constant.LoginIdentifierEmail:
		return verificationConstant.PurposeChangeEmail
	case constant.LoginIdentifierPhone:
		return verificationConstant.PurposeChangePhone
	default:
		return ""
	}
}

func identityKindLabel(kind constant.LoginIdentifierKind) string {
	switch kind {
	case constant.LoginIdentifierEmail:
		return "email"
	case constant.LoginIdentifierPhone:
		return "nomor telepon"
	default:
		return "identitas"
	}
}
