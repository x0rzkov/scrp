package main

import (
	"time"

	"github.com/gocolly/colly"
)

//ScrapeDetail :  Actually put the data into Cassandra, will only need to change this file in the future
func ScrapeDetail(c *colly.Collector) {

	//TODO: Process content
	type item struct {
		StoryURL  string
		Source    string
		comments  string
		CrawledAt time.Time
		Comments  string
		Title     string
	}
	// On every a element which has .top-matter attribute call callback
	// This class is unique to the div that holds all information about a story
	stories := []item{}
	c.OnHTML(".selectedidimtryingtofind", func(e *colly.HTMLElement) {
		//import "net"
		//import "net/url"
		// fmt.Println(u.Host)
		// host, port, _ := net.SplitHostPort(u.Host)
		// fmt.Println(host)
		// fmt.Println(port)
		temp := item{}
		temp.StoryURL = e.ChildAttr("a[data-event-action=title]", "href")
		temp.Source = "https://www.reddit.com/r/programming/"
		temp.Title = e.ChildText("a[data-event-action=title]")
		temp.Comments = e.ChildAttr("a[data-event-action=comments]", "href")
		temp.CrawledAt = time.Now()
		stories = append(stories, temp)
	})

}
