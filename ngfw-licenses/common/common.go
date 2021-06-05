package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/foolin/pagser"
)

func Dump(filename string, content []byte) {
	os.MkdirAll(filepath.Dir(filename), os.ModeDir|0755)
	ioutil.WriteFile(filename, content, 0500)
}

// ==== Pagser ====

func NewPagser() *pagser.Pagser {
	p := pagser.New()
	p.RegisterFunc("extractLicenseID", ExtractLicenseID)
	p.RegisterFunc("toUpper", ToUpper)
	p.RegisterFunc("contains", Contains)

	return p
}

func ExtractLicenseID(node *goquery.Selection, args ...string) (out interface{}, err error) {
	if node.Text() == "" {
		return "", nil
	}
	return strings.Split(node.Text(), " ")[3], nil
}

func ToUpper(node *goquery.Selection, args ...string) (out interface{}, err error) {
	i, _ := strconv.Atoi(args[0])
	return strings.ToUpper(strings.TrimSpace(node.Eq(i).Text())), nil
}

func Contains(node *goquery.Selection, args ...string) (out interface{}, err error) {
	return strings.Contains(node.Text(), args[0]), nil
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
