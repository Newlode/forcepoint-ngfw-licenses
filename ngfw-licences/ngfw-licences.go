package ngfwlicences

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/mbndr/logo"
)

var (
	Logger    *logo.Logger
	reNGFWPoS = regexp.MustCompile(`[a-fA-F0-9]{10}-[a-fA-F0-9]{10}`)
)

//=================================================================
// LicenceStatus

type LicenceStatus string

const (
	Unknown      LicenceStatus = "UNKNOWN"
	Valid        LicenceStatus = "VALID"
	Invalid      LicenceStatus = "INVALID"
	Registered   LicenceStatus = "REGISTERED"
	Unregistered LicenceStatus = "UNREGISTERED"
	Registering  LicenceStatus = "REGISTERING"
)

func (s LicenceStatus) String() string {
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
// POS

type POS struct {
	httpClient      *http.Client
	POS             string
	Status          LicenceStatus
	LicenceID       string
	ProductName     string
	LicenceFileHREF string
	SerialNumber    string
	Company         string
}

func NewPOS(pos string) *POS {
	cookieJar, _ := cookiejar.New(nil)
	return &POS{
		httpClient: &http.Client{Timeout: time.Second * 10, Jar: cookieJar},
		POS:        pos,
		Status:     Unknown,
	}
}

func (pos POS) String() string {
	return fmt.Sprintf("%s", aurora.Green(pos.POS))
}

func (pos POS) DetailedString() string {
	return fmt.Sprintf(`%s {LicenceStatus:"%s", SN:"%s", ProductName:"%s", Company:"%s"}`,
		pos,
		pos.Status,
		aurora.Gray(12, pos.SerialNumber),
		aurora.Gray(12, pos.ProductName),
		aurora.Gray(12, pos.Company),
	)
}

// CheckValidity is in charge of transitionning from state New to [Valid|Invalid]
func (pos *POS) CheckValidity() {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("licenseIdentification", pos.POS)
	w.Close()

	resp, _ := pos.httpClient.Post("https://stonesoftlicenses.forcepoint.com/license/load.do", w.FormDataContentType(), &buf)
	body, _ := ioutil.ReadAll(resp.Body)
	if match, _ := regexp.MatchString("(No license found with the given identifier|Permission denied)", string(body)); match {
		pos.Status = Invalid
		Logger.Errorf("Unable to load PoS %s", pos)
		return
	}
	pos.Status = Valid

	re := regexp.MustCompile(`(?m)<h3>Your Forcepoint License (\d+) .*</h3>(?:.*\s+)*<table class="LicenseData">\s+<caption>(.*)</caption>(?:.*\s+)*<caption>Appliance Hardware</caption>(?:.*\s+)*?<td>\s*(.*)\s*</td>(?:.*\s+)*<caption>License Company</caption>(?:.*\s+)*<td scope="row" class="Scope_row">\s+(.*?)\s+</td>\s+`)

	m := re.FindStringSubmatch(string(body))
	if len(m) != 5 {
		Logger.Errorf("Unable to parse licence data for %s", pos)
	} else {
		pos.LicenceID = m[1]
		pos.ProductName = m[2]
		pos.SerialNumber = m[3]
		pos.Company = m[4]
	}

	if match, _ := regexp.MatchString("Registered", string(body)); match {
		pos.Status = Registered
	}

}

//=================================================================
// POS List

func init() {
}

type POSList []*POS

func CreatePOSFormFiles() POSList {
	tmp := make([]string, 0)
	filenames, _ := filepath.Glob("*.html")
	for _, filename := range filenames {
		data, _ := ioutil.ReadFile(filename)
		tmp = append(tmp, reNGFWPoS.FindAllString(string(data), -1)...)
	}

	res := make(POSList, len(tmp))
	for i, s := range tmp {
		res[i] = NewPOS(s)
	}

	fmt.Printf("%d PoS read from %d files\n", len(res), len(filenames))
	return res
}

func (posList POSList) CheckValidity(concurrentWorkers int) {
	start := time.Now()
	wg := sync.WaitGroup{}
	toDo := make(chan *POS)
	done := make(chan *POS)

	wg.Add(concurrentWorkers)
	for i := 1; i <= concurrentWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wg.Done()
			for pos := range toDo {
				pos.CheckValidity()
				done <- pos
				count++
			}
			Logger.Debugf("Worker-%d done, %d PoS processed", id, count)
		}(i)
	}

	res := make(POSList, 0)
	go func() {
		for pos := range done {
			res = append(res, pos)
		}
	}()

	for _, pos := range posList {
		toDo <- pos
	}

	close(toDo)
	wg.Wait()
	close(done)

	Logger.Infof("%d PoS processed in %v\n", len(posList), time.Since(start).Truncate(time.Millisecond))
}

func (posList POSList) CountByStatus(state LicenceStatus) (res int) {
	for _, pos := range posList {
		if pos.Status == state {
			res++
		}
	}

	return res
}

func (posList POSList) GetByStatus(state LicenceStatus) (res POSList) {
	res = make(POSList, 0)
	for _, pos := range posList {
		if pos.Status == state {
			res = append(res, pos)
		}
	}

	return res
}

func (posList POSList) GetByNotStatus(state LicenceStatus) (res POSList) {
	res = make(POSList, 0)
	for _, pos := range posList {
		if pos.Status != state {
			res = append(res, pos)
		}
	}

	return res
}
