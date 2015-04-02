package main

import (
	"fmt"
	"github.com/twmb/algoimpl/go/graph"
	"github.com/jackdanger/collectlinks"
	"net/http"
	"log"
	"strings"
	"sort"
)

var config struct {
	url string
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
				for gi := range gns {
					ret = append(ret, gns[gi])
				}
				return ret
			}
		}
	}
	return ret
}

func getWikiLinks(page string) []string{
	res, err := http.Get(page)
	if err != nil {
		log.Fatal(err)
	}
	unfilteredLinks := collectlinks.All(res.Body)
	var filteredLinks []string
	for _, ul := range unfilteredLinks {
		// We only need wiki links
		if strings.HasPrefix(ul, "/wiki/") {
			// Remove anchors
			ul = strings.Split(ul, "#")[0]
			filteredLinks = append(filteredLinks, ul)
		}
	}

	filteredLinks = removeDuplicates(filteredLinks)

	return filteredLinks
}

func main() {
	config.url = "http://en.wikipedia.com"

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

	for !reached {
		// clear by slicing, doesn't clear the slice cap, slightly more efficient...
		newLinkBreadth = make([]string, 0)

		// go to each link and get their respective links
		for _, v1 := range currentLinkBreadth {
			fmt.Println( v1 )
			v1Links := getWikiLinks(config.url + v1)

			// add these to the graph and link them
			for _, v2 := range v1Links {
				// if the value isn't already in there
				if _, ok := links[v2]; !ok {
					links[v2] = linkGraph.MakeNode()
					linkGraph.MakeEdge(links[v1], links[v2])
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
			newLinkBreadth = append(newLinkBreadth, v1Links...)
		}
		currentLinkBreadth = newLinkBreadth
	}
}
