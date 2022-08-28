package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var fmtURL = "https://ko.wikipedia.org/wiki/%v%%EC%%9B%%94_%v%%EC%%9D%%BC"

type info struct {
	month int
	day int
	article string
}

func main() {
	pages := getPages()
	writePages(pages)
}

func writePages(pages []info) {
	file, err := os.Create("dates.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string { "Month", "Day", "Information"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, page := range pages {
		pageSlice := []string {
			strconv.Itoa(page.month),
			strconv.Itoa(page.day),
			page.article,
		}
		jwErr := w.Write(pageSlice)
		checkErr(jwErr)
	}
}

func getPages() []info {
	document := []info {}
	c := make(chan info)

	for month := 1; month <= 12; month++ {
		maxDay := 31
		if strings.Contains("2 4 6 9 11", strconv.Itoa(month)) == true {
			maxDay = 30
		}
		for day := 1; day <= maxDay; day++ {
			go getPage(month, day, c)
		}
	}

	for i := 1; i <= 12 * 30; i++ {
		document = append(document, <-c)
	}

	return document
}

func getPage(month int, day int, c chan<- info) int {
	baseURL := getBaseURL(month, day)
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res, month, day)

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	breakable := false
	doc.Find(".mw-parser-output p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		startIdx := 0

		if len(text) == 1 || breakable {
			return;
		}
		if i == 0 {
			startIdx = 1
		}

		c <- info { month, day, text[startIdx : len(text) - 1] }
		breakable = true
	})
	return 0
}

func getBaseURL(month int, day int) string {
	return fmt.Sprintf(fmtURL, month, day)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response, month int, day int) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode, month, day)
	}
}