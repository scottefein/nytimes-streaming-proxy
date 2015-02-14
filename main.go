package main

import (
    "fmt"
    "os"
    "github.com/ant0ine/go-json-rest/rest"
    "log"
    "net/http"
    "time"
    "io/ioutil"
    "encoding/json"
    "sort"
)

type JsonBodyNYT struct {
  Status string `json:"status"`
  Copyright string `json:"copyright"`
  NumResults int `json:"num_results"`
  Results []Article `json:"results"`
}

type Article struct {
    Section string `json:"section"`
    Subsection string `json:"subsection"`
    Title string `json:"title"`
    Abstract string `json:"abstract"`
    URL string `json:"url"`
    Byline string `json:"byline"`
    ThumbnailStandard string `json:"thumbnail_standard"`
    ItemType string `json:"item_type"`
    Source string `json:"source"`
    UpdatedDate time.Time `json:"updated_date"`
    CreatedDate time.Time `json:"created_date"`
    PublishedDate time.Time `json:"published_date"`
    MaterialTypeFacet string `json:"material_type_facet"`
    Kicker string `json:"kicker"`
    Subheadline string `json:"subheadline"`
    DesFacet []string `json:"des_facet"`
    OrgFacet []string `json:"org_facet"`
    PerFacet []string `json:"per_facet"`
    GeoFacet string `json:"geo_facet"`
    RelatedUrls string `json:"related_urls"`
    Multimedia string `json:"multimedia"`
}

type ByUpdatedDate []Article

func (a ByUpdatedDate) Len() int           { return len(a) }
func (a ByUpdatedDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByUpdatedDate) Less(i, j int) bool { return a[i].UpdatedDate.After(a[j].UpdatedDate) }

func get_nytimes_content()([]byte, error){
  url := fmt.Sprintf("http://api.nytimes.com/svc/news/v3/content/all/all.json?api-key=%s", os.Getenv("NYTIME_API_KEY"))
  resp, err := http.Get(url)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  return body, nil
}

func process_nytimes_content()([]Article){
  body, err := get_nytimes_content()
  if err != nil {
    //handle this
  }
  var response_body JsonBodyNYT
  var all_articles []Article
  json_err := json.Unmarshal(body, &response_body)
  if json_err != nil {
    //fmt.Println("error:", json_err)
  }
  all_articles = response_body.Results
  sort.Sort(ByUpdatedDate(all_articles))
  return all_articles
}

func StreamTheNYT(w rest.ResponseWriter, r *rest.Request) {
    cpt := 0
    var last_post Article
    for {
        cpt++
        var all_articles []Article
        all_articles = process_nytimes_content()
        var new_results []Article
        for _, e := range all_articles {
          if e.UpdatedDate.After(last_post.UpdatedDate) {
            new_results = append(new_results, e)
          }
        }
        sort.Sort(ByUpdatedDate(new_results))
        if len(new_results) > 0 {
          last_post = new_results[0]
          sort.Sort(sort.Reverse(ByUpdatedDate(new_results)))
          for _, e := range new_results{
            w.WriteJson(e)
          }
          w.(http.ResponseWriter).Write([]byte("\n"))
          // Flush the buffer to client
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