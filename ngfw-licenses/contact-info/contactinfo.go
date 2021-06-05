package contact_info

import "fmt"

//=================================================================
// ContactInfo

type ContactInfo struct {
	Binding   string `mapstructure:"binding"`
	Platform  string `mapstructure:"platform"`
	Firstname string `mapstructure:"firstname"`
	Lastname  string `mapstructure:"lastname"`
	Email     string `mapstructure:"email"`
	Phone     string `mapstructure:"phone"`
	Company   string `mapstructure:"company"`
	Address   string `mapstructure:"address"`
	Zip       string `mapstructure:"zip"`
	City      string `mapstructure:"city"`
	Country   string `mapstructure:"country"`
	State     string `mapstructure:"state"`
}

func validateContactInfoField(fieldName, fieldValue string, minLength, maxLength int) error {
	if len(fieldValue) < minLength {
		return fmt.Errorf("%s field cannot be less than %d", fieldName, minLength)
	} else if len(fieldValue) > maxLength {
		return fmt.Errorf("%s field cannot exceed %d characters (%d)", fieldName, maxLength, len(fieldValue))
	}
	return nil
}

func (contactInfo ContactInfo) Validate() error {
	if err := validateContactInfoField("contactinfo.firstname", contactInfo.Firstname, 2, 40); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.lastname", contactInfo.Lastname, 2, 40); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.email", contactInfo.Email, 2, 241); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.phone", contactInfo.Phone, 2, 25); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.company", contactInfo.Company, 2, 40); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.address", contactInfo.Address, 2, 60); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.zip", contactInfo.Zip, 2, 10); err != nil {
		return err
	} else if err := validateContactInfoField("contactinfo.city", contactInfo.City, 2, 40); err != nil {
		return err
	}

	return nil
}

func (contactInfo ContactInfo) GetFormData() map[string]string {
	return map[string]string{
		"terms":     "true",
		"firstname": contactInfo.Firstname,
		"lastname":  contactInfo.Lastname,
		"email":     contactInfo.Email,
		"phone":     contactInfo.Phone,
		"company":   contactInfo.Company,
		"address":   contactInfo.Address,
		"zip":       contactInfo.Zip,
		"city":      contactInfo.City,
		"country":   contactInfo.Country,
		"state":     contactInfo.State,
	}
}
