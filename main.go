package main

import (
	"fmt"
	"github.com/twmb/algoimpl/go/graph"
	"github.com/jackdanger/collectlinks"
	"net/http"
	"log"
	"strings"
	"sort"
	"time"
)

var config struct {
	url string
	maxConcurrentRequests int
	GETTries int
}

func removeDuplicates( links []string ) []string {
	sort.Strings(links)
	var ret []string
	lastVal := ""
	for _, v := range links {
		if lastVal != v {
			lastVal = v
			ret = append(ret, v)
		}
	}

	return ret
}

func stringifyWikiChain( links map[string]graph.Node, chain []graph.Node ) string {
	ret := ""

	// assign all the string values to the nodes
	for k, v := range links {
		*v.Value = k
	}

	for i := range chain {
		if str, ok := (*chain[i].Value).(string); ok {
			// add arrows
			if len(ret) > 0 {
				ret += " -> "
			}
			ret += str
		}
	}
	return ret

}

// This is a dfs
func extractLink( g *graph.Graph, n1, n2 graph.Node ) []graph.Node {
	var ret []graph.Node

	// same nodes
	if n1 == n2 {
		ret = append(ret, n1)
		return ret
	}

	// differing nodes
	neighbors := g.Neighbors(n1)
	for i := range neighbors {
		cn := neighbors[i]
		if cn == n2 {
			ret = append(ret, n1)
			ret = append(ret, cn)
			return ret
		} else {
			// recurse
			gns := extractLink( g, cn, n2 )
			if len(gns) > 0 {
				ret = append(ret, n1)
				ret = append(ret, gns...)
				return ret
			}
		}
	}
	return ret
}

func getWikiLinks(page string) []string{
	var err error
	var res *http.Response
	for i := 0; i < config.GETTries; i++ {
		res, err = http.Get(page)
		if err == nil {
			break
		}
		// close and try again
		res.Body.Close()
		time.Sleep(time.Second)
	}

	if err != nil {
		log.Fatalf("%s (failed after %d retries)\n", err, config.GETTries)
	}
	defer res.Body.Close()

	unfilteredLinks := collectlinks.All(res.Body)
	var filteredLinks []string
	for _, ul := range unfilteredLinks {
		// We only need wiki links
		if strings.HasPrefix(ul, "/wiki/") {
			// Remove anchors
			ul = strings.Split(ul, "#")[0]
			if !strings.Contains(ul, "Special:") &&
				 !strings.Contains(ul, "Help:") &&
				 !strings.Contains(ul, "Talk:") &&
				 !strings.Contains(ul, "Template:") &&
				 !strings.Contains(ul, "Template_talk:") &&
				 !strings.Contains(ul, "File:") &&
				 !strings.Contains(ul, "Wikipedia:") &&
				 !strings.Contains(ul, "Category:") &&
				 !strings.Contains(ul, "Portal:") {
				filteredLinks = append(filteredLinks, ul)
			}
		}
	}

	filteredLinks = removeDuplicates(filteredLinks)

	return filteredLinks
}

func main() {
	// wikipedia, no https
	config.url = "http://en.wikipedia.com"
	// at most 50 concurrent GET requests
	config.maxConcurrentRequests = 50
	// two retries
	config.GETTries = 3

	// grab the input vars from the user
	var srcWiki, dstWiki string
	var err error

	fmt.Printf("Enter the source wikipedia link: ")
	_, err = fmt.Scanf("%s", &srcWiki)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Enter the destination wikipedia link: ")
	_, err = fmt.Scanf("%s", &dstWiki)
	if err != nil {
		fmt.Println(err)
	}

	// Make sure we're on a real wiki page, not a base
	if !strings.Contains(srcWiki, "/wiki/") {
		log.Fatalf( "\"%s\" is not a valid wiki page\n", srcWiki )
	}
	if !strings.Contains(dstWiki, "/wiki/") {
		log.Fatalf( "\"%s\" is not a valid wiki page\n", dstWiki )
	}

	// filter the strings for the tail
	srcWiki = "/wiki/" + strings.Split(srcWiki, "/wiki/")[1]
	dstWiki = "/wiki/" + strings.Split(dstWiki, "/wiki/")[1]

	linkGraph := graph.New(graph.Directed)
	links := make(map[string]graph.Node)

	// add source and dest to the graph
	links[srcWiki] = linkGraph.MakeNode()
	links[dstWiki] = linkGraph.MakeNode()

	var newLinkBreadth []string
	currentLinkBreadth := []string{srcWiki}
	reached := false

	// {source, links}
	type linkType []string
	linkChan := make(chan struct {string; linkType})


	for !reached {
		newLinkBreadth = make([]string, 0)

		go func() {
			sem := make(chan bool, config.maxConcurrentRequests)
			// go to each link and get their respective links
			for _, v1 := range currentLinkBreadth {
			  sem <- true
				go func( v1Value string) {
					defer func() { <-sem }()
					fmt.Printf( "[ GET ] %s\n", v1Value )
					v1Links := getWikiLinks(config.url + v1Value)
					linkChan <- struct {string; linkType}{v1Value, v1Links}
				}(v1)
			}

			// fill up the semaphore for our remaining requests so they can finish
			for i := 0; i < cap(sem); i++ {
			    sem <- true
			}
		}()

		// listen for the requests to come back and process them until we've
		// exhausted the current breadth
		var v1LinksFiltered []string
		for i := 0; i<len(currentLinkBreadth); i++ {
			ans := <- linkChan
			v1 := ans.string
			v1Links := ans.linkType
			fmt.Printf("[ Processing ] %s\n", v1)
			// add these to the graph and link them
			for _, v2 := range v1Links {
				// if the value isn't already in there
				if _, ok := links[v2]; !ok {
					fmt.Printf("[ Adding %s ==> %s] \n", v1, v2)
					links[v2] = linkGraph.MakeNode()
					linkGraph.MakeEdge(links[v1], links[v2])
					v1LinksFiltered = append(v1LinksFiltered, v2)
				}

				// if it's the destination add it anyway, we done!
				if links[v2] == links[dstWiki] {
					linkGraph.MakeEdge(links[v1], links[v2])
					fmt.Println("The page has been reached!, determining links....")
					wikiChain := extractLink( linkGraph, links[srcWiki], links[dstWiki] )
					fmt.Printf("Wiki links chain: %s\n", stringifyWikiChain( links, wikiChain ))
					return
				}

			}
			// this will be for the next iteration of this loop, concat the slice
			newLinkBreadth = append(newLinkBreadth, v1LinksFiltered...)
		}
		currentLinkBreadth = newLinkBreadth
	}
}
