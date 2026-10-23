package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh melihat/mengubah data.
// Syaratnya: Dia pemiliknya (ownerID == UserID), ATAU dia punya permission :any.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true // Owner selalu boleh
	}
	return perms.Can(current.Role, anyPermission)
}