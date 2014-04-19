package main

import (
	"encoding/json"
	"fmt"
	// "io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Seen struct {
	cache map[string]bool
	m     sync.Mutex
}

func (s *Seen) Visit(x string) (found bool) {
	s.m.Lock()
	defer s.m.Unlock()
	ok := s.cache[x]
	if ok {
		return true
	}
	s.cache[x] = true
	return false
}

var (
	twitterHandle = "tiborvass"
	mazeURL       = "http://apibunny.com/mazes/v7yJKuRMd4DTXJO9"
	rgx           = regexp.MustCompile(`\{([^\}])*\}`)
	wg            sync.WaitGroup
	seen          = Seen{cache: map[string]bool{}}
	exitLink      string
)

type Document struct {
	Id         string
	Links      map[string]string
	Attributes map[string]interface{}
}

type Link struct {
	Href string
	Type string
}

type Documents map[string][]Document

type Links map[string]Link

type Node struct {
	Links      Links
	Documents  Documents
	Attributes map[string]interface{}
}

func main() {
	fmt.Println("digraph apibunny {")
	defer fmt.Println("}")
	wg.Add(1)
	VisitAll(mazeURL)
	wg.Wait()

	fmt.Fprintf(os.Stderr, "exit_link: %s\n", exitLink)
	/* Trying to make the exitLink work but couldn't.
	resp, err := http.Post(exitLink, "application/json", strings.NewReader(fmt.Sprintf(`{"submit": [{"twitter_handle": "%s"}]}`, twitterHandle)))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stderr, resp.Body)
	*/
}

func VisitAll(url string) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	raw := map[string]json.RawMessage{}
	err = d.Decode(&raw)
	if err != nil {
		log.Fatal(err)
	}
	node := Node{
		Links:      Links{},
		Documents:  Documents{},
		Attributes: map[string]interface{}{},
	}
	for k, v := range raw {
		if k == "links" {
			if err := json.Unmarshal(v, &node.Links); err != nil {
				log.Fatal(err)
			}
			continue
		}
		arr := []map[string]json.RawMessage{}
		if err := json.Unmarshal(v, &arr); err == nil {
			node.Documents[k] = make([]Document, len(arr))
			for i, m := range arr {
				doc := Document{
					Links:      map[string]string{},
					Attributes: map[string]interface{}{},
				}
				for kk, vv := range m {
					switch kk {
					case "id":
						if err := json.Unmarshal(vv, &doc.Id); err != nil {
							log.Fatal(err)
						}
					case "links":
						if err := json.Unmarshal(vv, &doc.Links); err != nil {
							log.Fatal(err)
						}
					default:
						var x interface{}
						if err := json.Unmarshal(vv, &x); err != nil {
							log.Fatal(err)
						}
						doc.Attributes[kk] = x
					}
				}
				node.Documents[k][i] = doc
			}
		} else {
			var x interface{}
			json.Unmarshal(v, &x)
			node.Attributes[k] = x
		}
	}

	for name, docs := range node.Documents {
		for _, doc := range docs {
			fmt.Printf("	\"%s\" [label=\"Id: http://apibunny.com/%s/%s\n%s\"];\n", doc.Id, name, doc.Id, concat(doc.Attributes))
			seen.Visit(doc.Id)
			if x, ok := doc.Attributes["exit_link"]; ok {
				exitLink = x.(string)
				exitLink = strings.Fields(exitLink)[0]
			}
			for linkName, link := range doc.Links {
				if seen.Visit(link) {
					continue
				}
				x := fmt.Sprintf("%s.%s", name, linkName)
				linkRef := node.Links[x]
				fmt.Printf("	\"%s\" -> \"%s\" [label=%s];\n", doc.Id, link, linkName)
				linkUrl := rgx.ReplaceAllStringFunc(linkRef.Href, func(s string) string {
					if s != fmt.Sprintf("{%s}", x) {
						panic(fmt.Errorf("'%s' != '%s'", s, x))
					}
					return link
				})
				wg.Add(1)
				go VisitAll(linkUrl)
			}
		}
	}
}

func concat(m map[string]interface{}) string {
	var s string
	for k, v := range m {
		s = fmt.Sprintf("%s%s: %v\n", s, k, v)
	}
	return s
}
