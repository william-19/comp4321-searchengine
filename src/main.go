package main

import (
	"fmt"
	"time"

	"./crawler"
	"./stopstem"
  "./database"
)

// Crawl crawl a given url as its base url,
// returning a mapping of url --> page struct
func Crawl(baseURL string) map[string]*crawler.Page {
	pagesMap := make(map[string]*crawler.Page)
	basePage := crawler.Page{
		URL:          baseURL,
		Title:        "",
		LastModified: "",
		PageSize:     "",
		Keywords:     make([]string, 0),
		ParentURL:    nil,
		ChildrenURL:  make([]string, 0)}

	basePage.ExtractTitle()
	basePage.ExtractLastModified()
	basePage.ExtractWords()
	basePage.ExtractSize()
	basePage.ExtractLinks()
	pagesMap[baseURL] = &basePage
	basePage.MakeChildren(&pagesMap)
	basePage.WriteIndexed(&pagesMap)

	return pagesMap
}

func main() {
	const baseURL = "https://www.cse.ust.hk/"
	fmt.Println(time.Now()) // buat ngecek dia brp lama runnya

	pagesMap := Crawl(baseURL) // get the mapping of url --> page struct
	fmt.Println("Len of map %v", len(pagesMap))


	// contoh cara ngambil page dari map
	// for _, page := range pagesMap {
	// 	fmt.Println(page.GetTitle())
	// 	fmt.Println(page.GetKeywords())
	// }

	// mapAwal := pagesMap["https://www.cse.ust.hk/admin/people/staff/"]
	// fmt.Println(mapAwal.GetTitle())
	// fmt.Println(mapAwal.GetKeywords())
	/*
		contoh cara lain buat ngambil page dari map
		another := pagesMap["http://epublish.ust.hk/cgi-bin/eng/story.php?id=96&catid=97&keycode=88b7aae0ae45ddb0e6e000ee2682721a&token=17b43a00aeb0f8f8f08df16ae664909f"]
		fmt.Println(another.GetTitle())
	*/
	fmt.Println("-------------------------------------------------------")
	stopstem.InputStopWords()
	newMap := stopstem.StemThemAll(&pagesMap)
	fmt.Println(len(newMap))
  database.OpenPageDb()

	for _, page := range newMap {
    _ = database.GetPageId(page.GetURL())
	}

  for _, page := range newMap {
    var test int64 = 0
    test = database.GetPageId(page.GetURL())
    fmt.Println(test)
  }

	fmt.Println(time.Now())
	// mapAkhir := newMap["https://www.cse.ust.hk/admin/people/staff/"]
	// fmt.Println(mapAkhir.GetTitle())
	// fmt.Println(mapAkhir.GetKeywords())
}
