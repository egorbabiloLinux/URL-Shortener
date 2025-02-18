package auth

type errorKeyType struct{}
type uidKeyType struct{}
type isAdminKeyType struct{}

var (
	errorKey   = errorKeyType{}
	uidKey     = uidKeyType{}
	isAdminKey = isAdminKeyType{}
)