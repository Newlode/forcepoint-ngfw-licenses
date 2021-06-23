package statutes

import "github.com/logrusorgru/aurora"

//=================================================================
// LicenseStatus

type LicenseStatus string
type SupportStatus string

const (
	Unknown LicenseStatus = "UNKNOWN"
	// Valid             LicenseStatus = "VALID"
	Invalid           LicenseStatus = "INVALID"
	Unregistered      LicenseStatus = "UNREGISTERED"
	Registering       LicenseStatus = "REGISTERING"
	Registered        LicenseStatus = "REGISTERED"
	Purchased         LicenseStatus = "PURCHASED"
	RegistrationError LicenseStatus = "REGISTRATION_ERROR"
	Activated         SupportStatus = "Activated"
	Expired           SupportStatus = "Expired"
	Spare             SupportStatus = "Spare"
)

var (
	// LicenseStatuses = []LicenseStatus{Registered, Unregistered, Registering, Unknown, Purchased, Valid, Invalid, RegistrationError}
	LicenseStatuses = []LicenseStatus{Registered, Unregistered, Registering, Unknown, Purchased, Invalid, RegistrationError}
)

func (s LicenseStatus) String() string {
	var r aurora.Value
	switch s {
	case Unknown:
		r = aurora.Blue(string(s))
	// case Valid:
	// 	r = aurora.Yellow(string(s))
	case Invalid:
		r = aurora.Red(string(s))
	case Registered:
		r = aurora.Green(string(s))
	default:
		r = aurora.Cyan(string(s))
	}

	return r.String()
}

func (s SupportStatus) String() string {
	var r aurora.Value
	switch s {
	case Activated:
		r = aurora.Green(string(s))
	case Expired:
		r = aurora.Yellow(string(s))
	case Spare:
		r = aurora.Blue(string(s))
	default:
		r = aurora.Gray(12, string(s))
	}

	return r.String()
}
