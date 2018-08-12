package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed/atom"
	"github.com/xyproto/unzip"
)

const mmlURL = "https://tiedostopalvelu.maanmittauslaitos.fi/tp/feed/mtp/kiinteistorekisterikartta/karttalehdittain?api_key=1s8so0nmpmnuspiud66uemonkb&format=application/x-shapefile"

const zippedFilePath = "zippedFiles/"

func main() {
	fmt.Println("Operation started", time.Now())
	fmt.Println("First fetch")
	nextURL, err := readFeedItemsFromURL(mmlURL)

	if err != nil {
		fmt.Printf("There was an error: %v", err)
	}

	hasNextURL := true
	for hasNextURL {
		fmt.Println("Next fetch")
		feedURL, err := readFeedItemsFromURL(nextURL)

		if feedURL == "" {
			hasNextURL = false
		} else {
			nextURL = feedURL
		}

		if err != nil {
			fmt.Printf("There was error: %v", err)
		}
	}

	fmt.Println("Operation ended", time.Now())
}

func getNextURL(atomFeed *atom.Feed) string {
	if len(atomFeed.Links) == 2 {
		return atomFeed.Links[1].Href
	}

	return ""
}

func readFeedItemsFromURL(feedURL string) (string, error) {
	xmlStr, err := getKiinteistorajatAtomXMLFromURL(feedURL)
	if err != nil {
		fmt.Printf("Failed to get XML: %v", err)
	}

	atomFeedParser := atom.Parser{}
	atomFeed, err := atomFeedParser.Parse(strings.NewReader(xmlStr))

	if err != nil {
		fmt.Printf("Error reading the feed: %v", err)
	}

	for _, element := range atomFeed.Entries {
		fileURL := element.Links[0].Href
		filePath := zippedFilePath + element.Title
		err := downloadFile(filePath, fileURL)
		if err != nil {
			panic(err)
		}
		extractZipFile(filePath)
	}

	return getNextURL(atomFeed), nil
}

func getKiinteistorajatAtomXMLFromURL(feedURL string) (string, error) {
	resp, err := http.Get(feedURL)

	if err != nil {
		return "", fmt.Errorf("GET error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("Read body: %v", err)
	}

	return string(data), nil
}

func extractZipFile(zipFileName string) {
	destination := "unzippedFiles"
	err := unzip.Extract(zipFileName, destination)
	if err != nil {
		panic(err)
	}
}

func downloadFile(filepath string, url string) error {
	if _, err := os.Stat(filepath); err == nil {
		return nil
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
