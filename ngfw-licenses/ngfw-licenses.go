package ngfwlicenses

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/logrusorgru/aurora"
	"github.com/mbndr/logo"
	"github.com/snwfdhmp/errlog"
)

var (
	Logger    *logo.Logger
	reNGFWPoS = regexp.MustCompile(`[a-fA-F0-9]{10}-[a-fA-F0-9]{10}`)
)

//=================================================================
// LicenseStatus

type LicenseStatus string

const (
	Unknown           LicenseStatus = "UNKNOWN"
	Valid             LicenseStatus = "VALID"
	Invalid           LicenseStatus = "INVALID"
	Unregistered      LicenseStatus = "UNREGISTERED"
	Registering       LicenseStatus = "REGISTERING"
	Registered        LicenseStatus = "REGISTERED"
	RegistrationError LicenseStatus = "REGISTRATION_ERROR"
)

func (s LicenseStatus) String() string {
	var r aurora.Value
	switch s {
	case Unknown:
		r = aurora.Blue(string(s))
	case Valid:
		r = aurora.Yellow(string(s))
	case Invalid:
		r = aurora.Red(string(s))
	case Registered:
		r = aurora.Green(string(s))
	default:
		r = aurora.Cyan(string(s))
	}

	return r.String()
}

//=================================================================
// ContactInfo

type ContactInfo struct {
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

//=================================================================
// POS

type POS struct {
	httpClient   *resty.Client
	POS          string
	Status       LicenseStatus
	LicenseID    string
	ProductName  string
	LicenseFile  string
	SerialNumber string
	Company      string
}

func NewPOS(pos string) *POS {
	return &POS{
		httpClient: resty.New(),
		POS:        pos,
		Status:     Unknown,
	}
}

func (pos POS) String() string {
	return fmt.Sprintf("%s", aurora.Green(pos.POS))
}

func (pos POS) DetailedString() string {
	return fmt.Sprintf(`%s {LicenseStatus:"%s", SN:"%s", ProductName:"%s", Company:"%s"}`,
		pos,
		pos.Status,
		aurora.Gray(12, pos.SerialNumber),
		aurora.Gray(12, pos.ProductName),
		aurora.Gray(12, pos.Company),
	)
}

// RefreshStatus is in charge of transitionning from state New to [Valid|Invalid]
func (pos *POS) RefreshStatus() {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.Close()

	resp, _ := pos.httpClient.R().
		SetFormData(map[string]string{"licenseIdentification": pos.POS}).
		Post("https://stonesoftlicenses.forcepoint.com/license/load.do")

	body := resp.Body()
	if match, _ := regexp.MatchString("(No license found with the given identifier|Permission denied)", string(body)); match {
		pos.Status = Invalid
		Logger.Infof("Unable to load PoS %s", pos)
		return
	}
	pos.Status = Valid

	re := regexp.MustCompile(`(?m)<h3>Your Forcepoint License (\d+) .*</h3>(?:.*\s+)*<table class="LicenseData">\s+<caption>(.*)</caption>(?:.*\s+)*<caption>Appliance Hardware</caption>(?:.*\s+)*?<td>\s*(.*)\s*</td>(?:.*\s+)*<caption>License Company</caption>(?:.*\s+)*<td scope="row" class="Scope_row">\s+(.*?)\s+</td>\s+`)
	m := re.FindStringSubmatch(string(body))
	if len(m) != 5 {
		Logger.Errorf("Unable to parse license data for %s", pos)
	} else {
		pos.LicenseID = m[1]
		pos.ProductName = m[2]
		pos.SerialNumber = m[3]
		pos.Company = m[4]
	}

	if match, _ := regexp.MatchString("Registered", string(body)); match {
		pos.Status = Registered

		// If jar is available, we store it as LicenseFile
		reSglic := regexp.MustCompile(`sglic-\d+-\d+\.jar`)
		pos.LicenseFile = string(reSglic.Find(body))
	}
}

// Register is in charge to register the POS using contactInfo and resseller
func (pos *POS) Register(contactInfo *ContactInfo, resseller string) {
	if pos.Status != Valid {
		Logger.Debugf("POS Status has to been 'Valid', current state is %v", pos.Status)
		return
	}

	// Send POST request
	pos.httpClient.R().
		SetFormData(map[string]string{
			"bindingtype[2]": "product.bindtype.pos",
			"binding[2]":     "",
			"platform[2]":    "Appliance",
			"terms":          "true",
			"firstname":      contactInfo.Firstname,
			"lastname":       contactInfo.Lastname,
			"email":          contactInfo.Email,
			"phone":          contactInfo.Phone,
			"company":        contactInfo.Company,
			"address":        contactInfo.Address,
			"zip":            contactInfo.Zip,
			"city":           contactInfo.City,
			"country":        contactInfo.Country,
			"state":          contactInfo.State,
			"reseller":       resseller,
		}).
		Post("https://stonesoftlicenses.forcepoint.com/license/registerstonegate/save.do")

	/*
		if errlog.Debug(err) {
			Logger.Errorf("%s: %v", pos.POS, err)
		}
	*/
}

func (pos *POS) WaitForLicenseFileGeneration() {
	maxEnd := time.Now().Add(time.Minute)
	for {
		pos.RefreshStatus()

		if pos.Status == Registered {
			break
		}

		if time.Now().After(maxEnd) {
			pos.Status = RegistrationError
			Logger.Errorf("%s: there was a problem when registering this POS", pos.POS)
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func (pos *POS) Download(outputDirectory string) bool {
	// Get the data
	_, err := pos.httpClient.R().
		SetOutput(outputDirectory + "/" + pos.LicenseFile).
		Get("https://stonesoftlicenses.forcepoint.com/license/licensefile.do?file=" + pos.LicenseFile)
	if errlog.Debug(err) {
		Logger.Errorf("%s: %v", pos.POS, err)
	}

	return true
}

//=================================================================
// POS List

type POSList []*POS

// CreatePOSFormFiles
func CreatePOSFormFiles() POSList {
	tmp := make([]string, 0)
	filenames, _ := filepath.Glob("*.html")
	for _, filename := range filenames {
		data, _ := ioutil.ReadFile(filename)
		tmp = append(tmp, reNGFWPoS.FindAllString(string(data), -1)...)
	}

	res := make(POSList, 0)
	posList := make([]string, 0)
	for _, pos := range tmp {
		// trying to get only unique PoS
		present := false
		for _, p := range posList {
			if pos == p {
				present = true
				break
			}
		}
		// allready here, skipping this one
		if present {
			continue
		}

		posList = append(posList, pos)
		res = append(res, NewPOS(pos))
	}

	fmt.Printf("%d PoS read from %d files\n", len(res), len(filenames))
	return res
}

// RefreshStatus
func (posList POSList) RefreshStatus(concurrentWorkers int) {
	start := time.Now()
	wgWorker := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *POS)
	done := make(chan *POS)

	res := make(POSList, 0)
	go posList.waitWorkDone(&wgWaiter, "Scanning", res, done)

	wgWorker.Add(concurrentWorkers)
	for i := 1; i <= concurrentWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorker.Done()
			for pos := range toDo {
				pos.RefreshStatus()
				done <- pos
				count++
			}
			Logger.Debugf("Worker-%d done, %d PoS processed", id, count)
		}(i)
	}

	for _, pos := range posList {
		toDo <- pos
	}
	close(toDo)

	wgWorker.Wait()
	close(done)
	wgWaiter.Wait()

	Logger.Infof("%d PoS processed in %v\n", len(posList), time.Since(start).Truncate(time.Millisecond))
}

// Register
func (posList POSList) Register(concurrentWorkers int, contactInfo *ContactInfo, resseller string) {
	if contactInfo == nil {
		Logger.Fatalf("Registrering PoS require contact informations from config file")
	}
	start := time.Now()
	wgWorkers := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *POS)
	done := make(chan *POS)
	counter := int64(0)

	res := make(POSList, 0)
	go posList.waitWorkDone(&wgWaiter, "Registrering", res, done)

	wgWorkers.Add(concurrentWorkers)
	for i := 1; i <= concurrentWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorkers.Done()
			for pos := range toDo {
				currentStatus := pos.Status
				pos.Register(contactInfo, resseller)
				pos.WaitForLicenseFileGeneration()
				done <- pos

				if currentStatus != pos.Status {
					count++
					atomic.AddInt64(&counter, 1)
				}
			}
			Logger.Debugf("Worker-%d done, %d PoS processed", id, count)
		}(i)
	}

	for _, pos := range posList {
		toDo <- pos
	}
	close(toDo)

	wgWorkers.Wait()
	close(done)
	wgWaiter.Wait()

	fmt.Printf("%d new POS were registred\n\n", counter)
	Logger.Infof("%d PoS processed in %v\n", len(posList), time.Since(start).Truncate(time.Millisecond))
}

// Download
func (posList POSList) Download(concurrentWorkers int, outputDirectory string) {
	_, err := os.Stat(outputDirectory)

	if os.IsNotExist(err) {
		err = os.Mkdir(outputDirectory, os.ModePerm)
		if errlog.Debug(err) {
			Logger.Fatalf("Unable to create directory %s", outputDirectory)
		}
	}

	start := time.Now()
	wgWorkers := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *POS)
	done := make(chan *POS)
	counter := int64(0)

	res := make(POSList, 0)
	go posList.waitWorkDone(&wgWaiter, "Downloading", res, done)

	wgWorkers.Add(concurrentWorkers)
	for i := 1; i <= concurrentWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorkers.Done()
			for pos := range toDo {
				pos.RefreshStatus()
				if pos.Download(outputDirectory) {
					count++
					atomic.AddInt64(&counter, 1)
				}

				done <- pos
			}
			Logger.Debugf("Worker-%d done, %d PoS processed", id, count)
		}(i)
	}

	for _, pos := range posList {
		toDo <- pos
	}
	close(toDo)

	wgWorkers.Wait()
	close(done)
	wgWaiter.Wait()

	fmt.Printf("%d license files were downloaded in './%s/' directory\n", counter, outputDirectory)
	Logger.Infof("%d PoS processed in %v\n", len(posList), time.Since(start).Truncate(time.Millisecond))
}

//================================================================
// Helpers

func (posList POSList) waitWorkDone(wg *sync.WaitGroup, prefix string, res POSList, done <-chan *POS) {
	var c = 0
	defer func() {
		wg.Done()
		Logger.Debugf("Waiter done, %d PoS processed", c)
	}()
	wg.Add(1)

	fmt.Println("")
	Logger.Debugf("Waiter started")
	for pos := range done {
		res = append(res, pos)
		c++
		fmt.Printf("\r%s %d/%d", prefix, len(res), len(posList))
	}
	fmt.Printf("\r                                  \r")
	bufStdout := bufio.NewWriter(os.Stdout)
	bufStdout.Flush()

}

func (posList POSList) CountByStatus(state LicenseStatus) (res int) {
	for _, pos := range posList {
		if pos.Status == state {
			res++
		}
	}

	return res
}

func (posList POSList) GetByStatus(state LicenseStatus) (res POSList) {
	res = make(POSList, 0)
	for _, pos := range posList {
		if pos.Status == state {
			res = append(res, pos)
		}
	}

	return res
}

func (posList POSList) GetByNotStatus(state LicenseStatus) (res POSList) {
	res = make(POSList, 0)
	for _, pos := range posList {
		if pos.Status != state {
			res = append(res, pos)
		}
	}

	return res
}

func (posList POSList) Display() {
	posOk := posList.GetByNotStatus(Invalid)
	fmt.Printf("Found %d valid PoS:\n", len(posOk))
	for _, pos := range posOk {
		fmt.Printf("- %v\n", pos.DetailedString())
	}

	posKo := posList.GetByStatus(Invalid)
	if len(posKo) > 0 {
		fmt.Printf("Found %d invalid PoS:\n", len(posKo))
		for _, pos := range posKo {
			fmt.Printf("- %v\n", pos.DetailedString())
		}
	}
}
