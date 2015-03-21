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

func extractLink( g *graph.Graph, n1, n2 graph.Node ) {
	// have to implement a dfs or bfs or some sort of s here :)
	return
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

	// clear by slicing, doesn't clear the slice cap, slightly more efficient...
	newLinkBreadth = make([]string, 0)

	// go to each link and get their respective links
	for _, v1 := range currentLinkBreadth {
		v1Links := getWikiLinks(config.url + srcWiki)

		// add these to the graph and link them
		for _, v2 := range v1Links {
			// if the value isn't already in there
			if _, ok := links[v2]; !ok {
				links[v2] = linkGraph.MakeNode()
			}

			linkGraph.MakeEdge(links[v1], links[v2])
		}

		// this will be for the next iteration of this loop, concat the slice
		newLinkBreadth = append(newLinkBreadth, v1Links...)
	}


	for _, curWiki := range newLinkBreadth {
		if curWiki == dstWiki {
			fmt.Println("Found it!")
			// not efficient or whatever, but gets the job done.
			//extractLink(linkGraph, links[srcWiki], links[dstWiki])
			//for _, v := range paths {
			//	fmt.Printf("%+v\n", v)
			//}
			//fmt.Printf("%+v\n", paths[1])
		}
	}

	currentLinkBreadth = newLinkBreadth



	// beyond this point is testing material

	//fmt.Println(srcWiki)
	//fmt.Println(dstWiki)


	//linkGraph.MakeEdge(links[srcWiki], links[dstWiki])
	//neighbors := linkGraph.Neighbors(links[srcWiki])
	//for _, node := range neighbors {
	//	for key, linkedNode := range links {
	//		if linkedNode == node {
	//			fmt.Printf("Connected to %s\n", key)
	//		}
	//	}
	//}

}
