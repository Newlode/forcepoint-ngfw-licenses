package pox

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Newlode/forcepoint-ngfw-licenses/config"
	"github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses/common"
	"github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses/statutes"
	"github.com/mbndr/logo"
	"github.com/snwfdhmp/errlog"
)

var (
	cfg    = &config.Cfg
	Logger *logo.Logger
)

//=================================================================
// PoX List

type PoXList []*PoX

// RefreshStatus
func (poxList PoXList) RefreshStatus() {
	start := time.Now()
	wgWorkers := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *PoX)
	done := make(chan *PoX)

	res := make(PoXList, 0)
	go poxList.waitWorkDone(&wgWaiter, "Scanning", res, done)

	nbWorkers := common.Min(cfg.ConcurrentWorkers, len(poxList))
	wgWorkers.Add(nbWorkers)
	for i := 1; i <= nbWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorkers.Done()
			for pox := range toDo {
				Logger.Debugf("Worker-%d start %s validation", id, pox.pox)
				pox.RefreshStatus(true)
				Logger.Debugf("Worker-%d finished %s, final status is %s", id, pox.pox, pox.Status)
				done <- pox
				count++
			}
			Logger.Debugf("Worker-%d done, %d PoS/PoL processed", id, count)
		}(i)
	}

	for _, pox := range poxList {
		toDo <- pox
	}
	close(toDo)

	wgWorkers.Wait()
	close(done)
	wgWaiter.Wait()

	Logger.Infof("%d PoS/PoL processed in %v\n", len(poxList), time.Since(start).Truncate(time.Millisecond))
}

// Register
func (poxList PoXList) Register() {
	if cfg.ContactInfo == nil {
		Logger.Fatalf("Registrering PoS/PoL require contact informations from config file")
	}
	start := time.Now()
	wgWorkers := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *PoX)
	done := make(chan *PoX)
	counter := int64(0)

	res := make(PoXList, 0)
	go poxList.waitWorkDone(&wgWaiter, "Registrering", res, done)

	nbWorkers := common.Min(cfg.ConcurrentWorkers, len(poxList))
	wgWorkers.Add(nbWorkers)
	for i := 1; i <= nbWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorkers.Done()
			for pox := range toDo {
				currentStatus := pox.Status
				pox.Register()
				pox.WaitForLicenseFileGeneration()
				done <- pox

				if currentStatus != pox.Status {
					Logger.Debugf("%s status changed from %v to %v", pox.pox, currentStatus, pox.Status)
					count++
					atomic.AddInt64(&counter, 1)
				}
			}
			Logger.Debugf("Worker-%d done, %d PoS/PoL processed", id, count)
		}(i)
	}

	for _, pox := range poxList {
		// We want to register only Purchased PoS/PoL
		Logger.Debugf("%+v", pox)
		if pox.Status == statutes.Purchased {
			Logger.Debugf("%s state is 'Purchased', trying to register", pox.pox)
			toDo <- pox
		}
	}
	close(toDo)

	wgWorkers.Wait()
	close(done)
	wgWaiter.Wait()

	if !cfg.Silent {
		fmt.Printf("%d new PoS have been registred\n\n", counter)
	}
	Logger.Infof("%d PoS/PoL processed in %v\n", len(poxList), time.Since(start).Truncate(time.Millisecond))
}

// ChangeBinding
func (poxList PoXList) ChangeBinding() {
	if cfg.ContactInfo == nil {
		Logger.Fatalf("Change binding require contact informations from config file")
	}
	start := time.Now()
	wgWorkers := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *PoX)
	done := make(chan *PoX)
	counter := int64(0)

	res := make(PoXList, 0)
	go poxList.waitWorkDone(&wgWaiter, "Change-binding", res, done)

	nbWorkers := common.Min(cfg.ConcurrentWorkers, len(poxList))
	wgWorkers.Add(nbWorkers)
	for i := 1; i <= nbWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorkers.Done()
			for pox := range toDo {
				initialBinding := pox.Binding
				pox.ChangeBinding()
				pox.WaitForBindingChange()
				done <- pox

				if initialBinding != pox.Binding {
					Logger.Debugf("%s binding changed from %v to %v", pox.pox, initialBinding, pox.Binding)
					count++
					atomic.AddInt64(&counter, 1)
				}
			}
			Logger.Debugf("Worker-%d done, %d PoS/PoL processed", id, count)
		}(i)
	}

	for _, pox := range poxList {
		// We want to register only Purchased PoS/PoL
		Logger.Debugf("%+v", pox)
		if pox.poxType == PoL && pox.Status == statutes.Registered && pox.Binding != cfg.Binding {
			Logger.Debugf("%s state is 'Registered', and Binding is different (%s -> %s), trying to register", pox.pox, pox.Binding, cfg.Binding)
			toDo <- pox
		}
	}
	close(toDo)

	wgWorkers.Wait()
	close(done)
	wgWaiter.Wait()

	if !cfg.Silent {
		fmt.Printf("%d binding have been changed\n\n", counter)
	}
	Logger.Infof("%d PoS/PoL processed in %v\n", len(poxList), time.Since(start).Truncate(time.Millisecond))
}

// Download
func (poxList PoXList) Download() {
	_, err := os.Stat(cfg.LicensesOutputDir)

	if os.IsNotExist(err) {
		err = os.Mkdir(cfg.LicensesOutputDir, os.ModePerm)
		if errlog.Debug(err) {
			Logger.Fatalf("Unable to create directory %s", cfg.LicensesOutputDir)
		}
	}

	start := time.Now()
	wgWorkers := sync.WaitGroup{}
	wgWaiter := sync.WaitGroup{}
	toDo := make(chan *PoX)
	done := make(chan *PoX)
	counter := int64(0)

	res := make(PoXList, 0)
	go poxList.waitWorkDone(&wgWaiter, "Downloading", res, done)

	nbWorkers := common.Min(cfg.ConcurrentWorkers, len(poxList))
	wgWorkers.Add(nbWorkers)
	for i := 1; i <= nbWorkers; i++ {
		go func(id int) {
			count := 0
			Logger.Debugf("Worker-%d started", id)
			defer wgWorkers.Done()
			for pox := range toDo {
				if pox.Download() {
					count++
					atomic.AddInt64(&counter, 1)
				}

				done <- pox
			}
			Logger.Debugf("Worker-%d done, %d PoS/PoL processed", id, count)
		}(i)
	}

	for _, pox := range poxList {
		// We want to dopwnload only Registered PoS/PoL
		if pox.Status != statutes.Registered {
			continue
		}

		toDo <- pox
	}
	close(toDo)

	wgWorkers.Wait()
	close(done)
	wgWaiter.Wait()

	if !cfg.Silent {
		fmt.Printf("%d license files have been downloaded in './%s/' directory\n", counter, cfg.LicensesOutputDir)
	}
	Logger.Infof("%d PoS/PoL processed in %v\n", len(poxList), time.Since(start).Truncate(time.Millisecond))
}

//================================================================
// Helpers

func (poxList PoXList) waitWorkDone(wg *sync.WaitGroup, prefix string, res PoXList, done <-chan *PoX) {
	var c = 0
	defer func() {
		wg.Done()
		Logger.Debugf("Waiter done, %d PoS/PoL processed", c)
	}()
	wg.Add(1)

	//! fmt.Println("")
	Logger.Debugf("Waiter started")
	for pox := range done {
		res = append(res, pox)
		c++
		if !cfg.Silent {
			fmt.Printf("\r%s %d/%d", prefix, len(res), len(poxList))
		}
	}
	if !cfg.Silent {
		fmt.Printf("\r                                  \r")
	}
	bufStdout := bufio.NewWriter(os.Stdout)
	bufStdout.Flush()
}

func (poxList PoXList) CountByStatus(state statutes.LicenseStatus) (res int) {
	for _, pox := range poxList {
		if pox.Status == state {
			res++
		}
	}

	return res
}

func (poxList PoXList) GetByStatus(state statutes.LicenseStatus) (res PoXList) {
	res = make(PoXList, 0)
	for _, pox := range poxList {
		if pox.Status == state {
			res = append(res, pox)
		}
	}

	return res
}

func (poxList PoXList) getByType(_type PoXType) (res PoXList) {
	res = make(PoXList, 0)
	for _, pox := range poxList {
		if pox.poxType == _type {
			res = append(res, pox)
		}
	}

	return res
}

func (poxList PoXList) GetAllPoL() (res PoXList) {
	return poxList.getByType(PoL)
}

func (poxList PoXList) GetAllPoS() (res PoXList) {
	return poxList.getByType(PoS)
}

func (poxList PoXList) GetByNotStatus(state statutes.LicenseStatus) (res PoXList) {
	res = make(PoXList, 0)
	for _, pox := range poxList {
		if pox.Status != state {
			res = append(res, pox)
		}
	}

	return res
}

func (poxList PoXList) Display() {
	for _, status := range statutes.LicenseStatuses {
		poxList := poxList.GetByStatus(status)
		if len(poxList) > 0 {
			for _, poxType := range []PoXType{PoL, PoS} {
				if len(poxList.getByType(poxType)) == 0 {
					continue
				}
				fmt.Printf("\nFound %d %v %s:\n",
					len(poxList.getByType(poxType)),
					strings.ToLower(string(status)),
					poxType,
				)
				for _, pox := range poxList.getByType(poxType) {
					_ = pox
					if pox.Error != "" {
						fmt.Printf("- %v\n", pox.DetailedError())
					} else {
						fmt.Printf("- %v\n", pox.DetailedString())
					}
				}
			}
		}
	}
}

//=================================================================
// PoL/PoS List

var (
	reNGFWPoL = regexp.MustCompile(`[a-fA-F0-9]{5}-[a-fA-F0-9]{5}-[a-fA-F0-9]{5}-[a-fA-F0-9]{5}`)
	reNGFWPoS = regexp.MustCompile(`[a-fA-F0-9]{10}-[a-fA-F0-9]{10}`)
)

// ReadPoXFormArgs
func ReadPoXFormArgs(args []string, posOnly, polOnly bool) PoXList {
	polList, posList := make([]string, 0), make([]string, 0)

	countPoLFromArgs, countPoSFromArgs := 0, 0
	countPoLFromFiles, countPoSFromFiles, countFiles := 0, 0, 0

	for _, arg := range args {
		// if arg is a PoL
		if reNGFWPoL.MatchString(arg) {
			// and if we want to read PoL
			if !posOnly {
				polList = append(polList, arg)
				countPoLFromArgs++
			}
		} else if reNGFWPoS.MatchString(arg) {
			// and if we want to read PoS
			if !polOnly {
				posList = append(posList, arg)
				countPoSFromArgs++
			}
		} else {
			// else, it should be a filename
			data, err := ioutil.ReadFile(arg)
			if err != nil {
				Logger.Fatalf("%v", err)
			}
			var r []string
			if !polOnly {
				r = dedup(reNGFWPoL.FindAllString(string(data), -1))
				polList = append(polList, r...)
				countPoLFromFiles += len(r)
			}
			if !posOnly {
				r = dedup(reNGFWPoS.FindAllString(string(data), -1))
				posList = append(posList, r...)
				countPoSFromFiles += len(r)
			}
			countFiles++
		}
	}

	Logger.Infof("%d PoL and %d PoS read, %d PoL and %d PoS from command-line, and %d PoL and %d PoS from %d files",
		countPoLFromArgs+countPoLFromFiles, countPoSFromArgs+countPoSFromFiles,
		countPoLFromArgs, countPoSFromArgs, countPoLFromFiles, countPoSFromFiles, countFiles)
	if !cfg.Silent {
		fmt.Printf("%d PoL and %d PoS read, %d PoL and %d PoS from command-line, and %d PoL and %d PoS from %d files\n",
			countPoLFromArgs+countPoLFromFiles, countPoSFromArgs+countPoSFromFiles,
			countPoLFromArgs, countPoSFromArgs, countPoLFromFiles, countPoSFromFiles, countFiles)
	}

	return append(createPoX(polList, NewPoL), createPoX(posList, NewPoS)...)
}

func dedup(list []string) (res []string) {
	r := make(map[string]bool)
	for _, s := range list {
		r[s] = true
	}
	for k := range r {
		res = append(res, k)
	}
	return res
}

func createPoX(list []string, fct func(string) *PoX) PoXList {
	res := make(PoXList, 0)
	dedupList := make([]string, 0)

	for _, pox := range list {
		// trying to get only unique PoS
		present := false
		for _, p := range dedupList {
			if pox == p {
				present = true
				break
			}
		}

		// allready here, skipping this one
		if present {
			continue
		}

		dedupList = append(dedupList, pox)

		res = append(res, fct(pox))
	}

	return res
}
