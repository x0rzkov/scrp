package main

import (
	"fmt"
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
		temp := item{}
		temp.StoryURL = e.ChildAttr("a[data-event-action=title]", "href")
		temp.Source = "https://www.reddit.com/r/programming/"
		temp.Title = e.ChildText("a[data-event-action=title]")
		temp.Comments = e.ChildAttr("a[data-event-action=comments]", "href")
		temp.CrawledAt = time.Now()
		stories = append(stories, temp)
	})

}

// InsertContent will choose a random server in the cluster to write to until a successful write
// occurs, logging each unsuccessful. If all servers fail, return error.
func (i *Cassandra) InsertContent(records map[string]string) error {
	//TODO: performance test against batching
	//fmt.Fprintf(os.Stderr, "Input packet", metrics)
	// This will get set to nil if a successful write occurs
	err := fmt.Errorf("Could not write to any cassandra server in cluster")
	//counters := make(map[string]int)
	// regexCount, _ := regexp.Compile(`\.count\.(.*)`)
	// regexUpdate, _ := regexp.Compile(`\.update\.(.*)`)
	//insertBatch := i.session.NewBatch(gocql.UnloggedBatch)
	// for k, v := range records {
	// 	//fmt.Println("%s", tags) //Debugging only
	// 	if regexCount.MatchString(records["name"]) {
	// 		counter := regexCount.FindStringSubmatch(records["name"])[1]
	// 		counters[counter] = counters[counter] + 1
	// 	} else if regexUpdate.MatchString(tags["name"]) && tags["msg"] != "" {
	// 		timestamp := time.Now().UTC()
	// 		if tags["updated"] != "" {
	// 			millis, err := strconv.ParseInt(tags["updated"], 10, 64)
	// 			if err == nil {
	// 				timestamp = time.Unix(0, millis*int64(time.Millisecond))
	// 			}
	// 		}
	// 		if rowError := i.session.Query(`INSERT INTO updates (id, updated, msg) values (?,?,?)`,
	// 			regexUpdate.FindStringSubmatch(tags["name"])[1],
	// 			timestamp,
	// 			tags["msg"]).Exec(); rowError != nil {
	// 			err = rowError //And let it continue
	// 		} else {
	// 			err = nil
	// 		}
	// 	} else {
	// 		if tags["id"] == "" {
	// 			tags["id"] = gocql.TimeUUID().String()
	// 		}
	// 		serialized, _ := json.Marshal(tags)
	// 		//insertBatch.Query(`INSERT INTO logs JSON ?`, string(serialized))
	// 		if rowError := i.session.Query(`INSERT INTO logs JSON ?`, string(serialized)).Exec(); rowError != nil {
	// 			err = rowError //And let it continue
	// 		} else {
	// 			err = nil
	// 		}
	// 	}
	// }

	// for key, value := range counters {
	// 	if rowError := i.session.Query(`UPDATE counters set total=total+? where id=?;`, value, key).Exec(); rowError != nil {
	// 		err = rowError //And let it continue
	// 	} else {
	// 		err = nil
	// 	}
	// }

	// //err = i.session.ExecuteBatch(insertBatch)
	// if !i.Retry && err != nil {
	// 	fmt.Fprintf(os.Stderr, "!E CASSANDRA OUTPUT PLUGIN - NOT RETRYING %s", err.Error())
	// 	err = nil //Do not retry
	// }
	return err
}
