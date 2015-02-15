package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

type Article struct {
	Section           string    `json:"section"`
	Subsection        string    `json:"subsection"`
	Title             string    `json:"title"`
	Abstract          string    `json:"abstract"`
	URL               string    `json:"url"`
	Byline            string    `json:"byline"`
	ThumbnailStandard string    `json:"thumbnail_standard"`
	ItemType          string    `json:"item_type"`
	Source            string    `json:"source"`
	UpdatedDate       time.Time `json:"updated_date"`
	CreatedDate       time.Time `json:"created_date"`
	PublishedDate     time.Time `json:"published_date"`
	MaterialTypeFacet string    `json:"material_type_facet"`
	Kicker            string    `json:"kicker"`
	Subheadline       string    `json:"subheadline"`
	DesFacet          []string  `json:"des_facet"`
	OrgFacet          []string  `json:"org_facet"`
	PerFacet          []string  `json:"per_facet"`
	GeoFacet          string    `json:"geo_facet"`
	RelatedUrls       string    `json:"related_urls"`
	Multimedia        string    `json:"multimedia"`
}

type Articles []Article

func (a Articles) Len() int           { return len(a) }
func (a Articles) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Articles) Less(i, j int) bool { return a[i].UpdatedDate.After(a[j].UpdatedDate) }

func fetchContent() (Articles, error) {
	url := fmt.Sprintf("http://api.nytimes.com/svc/news/v3/content/all/all.json?api-key=%s", os.Getenv("NYTIME_API_KEY"))

	res, err := http.Get(url)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	r := &struct {
		Articles Articles `json:"results"`
	}{}

	json.NewDecoder(res.Body).Decode(r)
	return r.Articles, err
}

func StreamTheNYT(w rest.ResponseWriter, r *rest.Request) {
	last := Article{}

	for {
		articles, err := fetchContent()
		if err != nil {
			log.Fatal(err)
		}

		results := Articles{}

		for _, e := range articles {
			if e.UpdatedDate.After(last.UpdatedDate) {
				results = append(results, e)
			}
		}
		sort.Sort(results)

		if len(results) > 0 {
			last = results[0]
			sort.Sort(sort.Reverse(results))
			for _, e := range results {
				w.WriteJson(e)
			}
			w.(http.ResponseWriter).Write([]byte("\n"))
			w.(http.Flusher).Flush()
		}

		time.Sleep(time.Duration(30) * time.Second)
	}
}

func main() {
	api := rest.NewApi()
	api.Use(&rest.AccessLogApacheMiddleware{})
	api.Use(rest.DefaultCommonStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/stream", StreamTheNYT},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
