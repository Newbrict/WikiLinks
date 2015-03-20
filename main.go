package main

import (
	"fmt"
	"github.com/twmb/algoimpl/go/graph"
	"github.com/jackdanger/collectlinks"
	"net/http"
	"log"
	"strings"
)

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
	return filteredLinks
}

func main() {
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

	fmt.Println(srcWiki)
	fmt.Println(dstWiki)

	linkGraph := graph.New(graph.Directed)

	links := make(map[string]graph.Node)

	links[srcWiki] = linkGraph.MakeNode()
	links[dstWiki] = linkGraph.MakeNode()
	linkGraph.MakeEdge(links[srcWiki], links[dstWiki])

	neighbors := linkGraph.Neighbors(links[srcWiki])
	for _, node := range neighbors {
		for key, linkedNode := range links {
			if linkedNode == node {
				fmt.Printf("Connected to %s\n", key)
			}
		}
	}

	for _, y := range getWikiLinks(srcWiki) {
		fmt.Println(y)
	}
}
