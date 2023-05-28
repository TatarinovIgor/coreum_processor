package user

import "coreum_processor/modules/storage"

func IsBlocked(access storage.UserAccess) bool {
	return ((access) & storage.UserRegistered) == storage.UserBlocked
}
func IsRegistered(access storage.UserAccess) bool {
	return access == storage.UserRegistered
}
func IsOnboarding(access storage.UserAccess) bool {
	return ((access & storage.UserRegistered) != storage.UserBlocked) &&
		((access & storage.UserOnboarded) != storage.UserOnboarded)
}
func IsOnboarded(access storage.UserAccess) bool {
	return ((access & storage.UserRegistered) != storage.UserBlocked) &&
		((access & storage.UserOnboarded) == storage.UserOnboarded)
}
func IsSysAdmin(access storage.UserAccess) bool {
	return ((access & storage.UserRegistered) != storage.UserBlocked) &&
		((access & storage.UserSysAdmin) == storage.UserSysAdmin)
}

func SetBlocked(access storage.UserAccess) storage.UserAccess {
	return access & ^storage.UserRegistered
}
func SetRegistered(access storage.UserAccess) storage.UserAccess {
	return access | storage.UserRegistered
}

func SetOnboarded(access storage.UserAccess) storage.UserAccess {
	return access | storage.UserOnboarded
}
func RemoveOnboarded(access storage.UserAccess) storage.UserAccess {
	return access & ^storage.UserOnboarded
}

func SetSysAdmin(access storage.UserAccess) storage.UserAccess {
	return access | storage.UserSysAdmin
}
func RemoveSysAdmin(access storage.UserAccess) storage.UserAccess {
	return access & ^storage.UserSysAdmin
}
