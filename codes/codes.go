package codes

func AreCodesValid(countryCode, stateCode string) bool {

	if _, countryCodeOK := Countries[countryCode]; countryCodeOK {
		if _, stateCodeOK := States[countryCode][stateCode]; stateCodeOK {
			return true
		}
	}

	return false
}
