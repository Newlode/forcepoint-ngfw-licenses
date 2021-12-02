package pox

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses/common"
	"github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses/statutes"
	"github.com/go-resty/resty/v2"
	"github.com/logrusorgru/aurora"
	"github.com/snwfdhmp/errlog"
)

//=================================================================
// PoX

type PoXType string

const (
	PoL PoXType = "PoL"
	PoS PoXType = "PoS"
)

type PoX struct {
	httpClient *resty.Client

	poxType            PoXType
	pox                string
	PoL                string                 `json:"pol,omitempty"`
	PoS                string                 `json:"pos,omitempty"`
	Status             statutes.LicenseStatus `json:"licence_status" pagser:"div[id=MSC_Content] h2+h3+table tbody tr td->toUpper(0)"`
	LicenseID          string                 `json:"license_id" pagser:"div[id=MSC_Content] h2+h3->extractLicenseID()"`
	ProductName        string                 `json:"product_name" pagser:"div[id=MSC_Content] h2"`
	Binding            string                 `json:"binding" pagser:"div[id=MSC_Content] h2+h3+table tbody tr td->eq(1)"`
	Platform           string                 `json:"platform" pagser:"div[id=MSC_Content] h2+h3+table tbody tr td->eq(3)"`
	LicensePeriod      string                 `json:"license_period,omitempty" pagser:"div[id=MSC_Content] h2+h3+table tbody tr td->eq(4)"`
	LicenseFile        string                 `json:"license_file" pagser:"div[id=MSC_Content] caption:contains('License File')+thead+tbody tr td->eq(0)"`
	MaintenanceStatus  statutes.SupportStatus `json:"support_status" pagser:"div[id=MSC_Content] caption:contains('Support & Maintenance')+thead+tbody tr td->eq(0)"`
	MaintenanceEndDate string                 `json:"support_end_date" pagser:"div[id=MSC_Content] caption:contains('Support & Maintenance')+thead+tbody tr td->eq(1)"`
	SerialNumber       string                 `json:"serial_number,omitempty" pagser:"div[id=MSC_Content] caption:contains('Appliance Hardware')+thead+tbody tr td->eq(0)"`
	Company            string                 `json:"company" pagser:"div[id=MSC_Content] caption:contains('License Company')+thead+tbody"`

	IsSpare bool `json:"is_spare" pagser:"div[id=MSC_Content] caption:contains('Support & Maintenance')+thead+tbody tr th->contains('No Support & Maintenance')"`

	Error string `json:"error,omitempty"`
}

func NewPoL(pol string) *PoX {
	if !reNGFWPoL.MatchString(pol) {
		Logger.Fatalf("%v is not a valid PoL", pol)
	}
	return &PoX{
		httpClient: resty.New(),
		poxType:    PoL,
		pox:        pol,
		PoL:        pol,
		Status:     statutes.Unknown,
	}
}

func NewPoS(pos string) *PoX {
	if !reNGFWPoS.MatchString(pos) {
		Logger.Fatalf("%v is not a valid PoS", pos)
	}
	return &PoX{
		httpClient: resty.New(),
		poxType:    PoS,
		pox:        pos,
		PoS:        pos,
		Status:     statutes.Unknown,
	}
}

func (pox PoX) String() string {
	return fmt.Sprintf("%s", aurora.Green(pox.pox))
}

func (pox PoX) DetailedString() string {
	// maintenanceStatus := aurora.Gray(12, pos.MaintenanceStatus)
	maintenanceEndDate := aurora.Gray(12, pox.MaintenanceEndDate)

	switch pox.MaintenanceStatus {
	case statutes.Expired:
		maintenanceEndDate = aurora.Yellow(pox.MaintenanceEndDate)
	case statutes.Spare:
		maintenanceEndDate = aurora.Blue(pox.MaintenanceEndDate)
	}

	res := ""
	switch pox.poxType {
	case PoL:
		res = fmt.Sprintf(`%s {LicenseStatus:"%s", Binding:"%s", ProductName:"%s", MaintenanceStatus:"%s", MaintenanceEndDate:"%s", Company:"%s"}`,
			pox,
			pox.Status,
			aurora.Gray(12, pox.Binding),
			aurora.Gray(12, pox.ProductName),
			pox.MaintenanceStatus,
			maintenanceEndDate,
			aurora.Gray(12, pox.Company),
		)
	case PoS:
		res = fmt.Sprintf(`%s {LicenseStatus:"%s", SN:"%s", ProductName:"%s", MaintenanceStatus:"%s", MaintenanceEndDate:"%s", Company:"%s"}`,
			pox,
			pox.Status,
			aurora.Gray(12, pox.SerialNumber),
			aurora.Gray(12, pox.ProductName),
			pox.MaintenanceStatus,
			maintenanceEndDate,
			aurora.Gray(12, pox.Company),
		)
	}

	return res
}

func (pox PoX) DetailedError() string {
	return fmt.Sprintf(`%s {LicenseStatus:"%s", Error:"%s"}`,
		pox,
		pox.Status,
		aurora.Yellow(pox.Error),
	)
}

// RefreshStatus is in charge of transitionning from state New to [Valid|Invalid]
func (pox *PoX) RefreshStatus(showErrors bool) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.Close()

	var body []byte
	for {
		resp, _ := pox.httpClient.R().
			SetFormData(map[string]string{"licenseIdentification": pox.pox}).
			Post("https://stonesoftlicenses.forcepoint.com/license/load.do")

		body = resp.Body()
		if strings.Contains(string(body), "No license found with the given identifier") {
			pox.Status = statutes.Invalid
			pox.Error = "No license found with the given identifier"
			Logger.Infof("'No license found with the given identifier' for %s", pox)
			return
		}
		if strings.Contains(string(body), "Permission denied") {
			pox.Status = statutes.Invalid
			pox.Error = "Permission denied"
			Logger.Infof("'Permission denied' for %s", pox)
			time.Sleep(10 * time.Second)
			continue
		}

		// pox.Status = statutes.Valid
		pox.Error = ""
		break
	}

	pox.refreshStatus(showErrors, body)
}

func (pox *PoX) refreshStatus(showErrors bool, body []byte) {
	Logger.Infof("refreshStatus call for %s %s", pox.poxType, pox.pox)

	common.Dump("dumps/"+pox.pox+"/"+time.Now().Format("20060102-150405")+"-refresh.html", body)
	err := common.NewPagser().Parse(pox, string(body))
	//check error
	if err != nil {
		log.Fatal(err)
	}

	if pox.IsSpare {
		pox.MaintenanceStatus = statutes.Spare
		pox.MaintenanceEndDate = "Spare"
	}
}

func (pox *PoX) getFormData() map[string]string {
	res := cfg.ContactInfo.GetFormData()

	res["resseller"] = cfg.Reseller

	switch pox.poxType {
	case PoL:
		res["bindingtype[1]"] = "product.bindtype.pol"
		res["binding[1]"] = cfg.Binding
		res["platform[1]"] = "Linux"
	case PoS:
		res["bindingtype[2]"] = "product.bindtype.pos"
		res["binding[2]"] = ""
		res["platform[2]"] = "Appliance"
	}

	return res
}

// Register is in charge to register the PoS using contactInfo and resseller
func (pox *PoX) Register() {
	if pox.poxType == PoL && pox.Status != statutes.Purchased {
		Logger.Debugf("PoL Status has to been 'Purchased', current state is %v", pox.Status)
		return
	}

	if pox.poxType == PoS && pox.Status != statutes.Purchased {
		Logger.Debugf("PoS Status has to been 'Purchased', current state is %v", pox.Status)
		return
	}

	// Send POST request
	resp, _ := pox.httpClient.R().
		SetFormData(pox.getFormData()).
		Post("https://stonesoftlicenses.forcepoint.com/license/registerstonegate/save.do")

	common.Dump("dumps/"+pox.pox+"/"+time.Now().Format("20060102-150405")+"-register.html", resp.Body())
}

func (pox *PoX) ChangeBinding() {
	if pox.poxType == PoS {
		Logger.Debugf("Binding change on PoS is not supported. please open an issue if you need it.", pox.Status)
		return
	}

	resp, _ := pox.httpClient.R().
		SetFormData(pox.getFormData()).
		Post("https://stonesoftlicenses.forcepoint.com/license/changeaddress/save.do")

	common.Dump("dumps/"+pox.pox+"/"+time.Now().Format("20060102-150405")+"-changebinding.html", resp.Body())
}

func (pox *PoX) WaitForLicenseFileGeneration() {
	maxEnd := time.Now().Add(time.Minute * 2)
	for {
		pox.RefreshStatus(true)

		if pox.Status == statutes.Registered {
			break
		}

		if time.Now().After(maxEnd) {
			pox.Status = statutes.RegistrationError
			Logger.Errorf("%s: there was a problem when registering this %s", pox.pox, pox.poxType)
			break
		}
		time.Sleep(15 * time.Second)
	}
}

func (pox *PoX) WaitForBindingChange() {
	maxEnd := time.Now().Add(time.Minute * 2)
	for {
		pox.RefreshStatus(true)

		if pox.Binding == cfg.Binding {
			break
		}

		if time.Now().After(maxEnd) {
			pox.Status = statutes.RegistrationError
			Logger.Errorf("%s: there was a problem when change-binding this %s", pox.pox, pox.poxType)
			break
		}
		time.Sleep(15 * time.Second)
	}
}

func (pox *PoX) Download() bool {
	// Get the data
	resp, err := pox.httpClient.R().
		SetOutput(cfg.LicensesOutputDir + "/" + pox.LicenseFile).
		Get("https://stonesoftlicenses.forcepoint.com/license/licensefile.do?file=" + pox.LicenseFile)
	if errlog.Debug(err) {
		Logger.Errorf("%s: %v", pox.pox, err)
	}

	common.Dump("dumps/"+pox.pox+"/"+time.Now().Format("20060102-150405")+"-download.html", resp.Body())

	return true
}
