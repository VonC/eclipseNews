package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/kong"
	"github.com/atotto/clipboard"
)

var cli struct {
	EclipseVersion string `kong:"required,help='Eclipse version',alias='eversion',short='e'"`
	TitleToMatch   string `kong:"required,help='Title or part of a title to match',alias='title',short='t'"`
}

func main() {

	// Parse command line arguments
	kong.Parse(&cli)

	// Use the command line arguments
	eclipseVersion := cli.EclipseVersion
	titleToMatch := cli.TitleToMatch

	// URLs of the subPages
	subPageURLs := []string{
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/platform.php",
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/java.php",
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/pde.php",
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/jdt.php",
	}

	matchedTitles := make([]string, 0)

	for _, url := range subPageURLs {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching URL:", err)
			continue
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
		if err != nil {
			fmt.Println("Error loading HTML into goquery:", err)
			continue
		}

		// Find the features
		doc.Find("td.title").Each(func(i int, s *goquery.Selection) {
			featureTitle := s.Text()
			if strings.Contains(strings.ToLower(featureTitle), strings.ToLower(titleToMatch)) {
				// Found a match, add it to the list
				matchedTitles = append(matchedTitles, featureTitle)
			}
		})
	}

	if len(matchedTitles) > 1 {
		// If more than one title matched, just print the titles
		for _, title := range matchedTitles {
			fmt.Println(title)
		}
	} else if len(matchedTitles) == 1 {
		// If exactly one title matched, print the title and body
		markdownResults := make([]string, 0)
		for _, url := range subPageURLs {
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error fetching URL:", err)
				continue
			}
			defer resp.Body.Close()

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
			if err != nil {
				fmt.Println("Error loading HTML into goquery:", err)
				continue
			}

			// Find the features again
			doc.Find("td.title").Each(func(i int, s *goquery.Selection) {
				featureTitle := s.Text()
				if featureTitle == matchedTitles[0] {
					// Found the match again, now get the body
					body := s.Parent().Next().Find("td.content").Text()

					// Convert to markdown
					converter := md.NewConverter("", true, nil)
					markdown, _ := converter.ConvertString(body)

					// Print the title and body in markdown format
					markdownResult := "> ## " + featureTitle + "\n"
					for _, line := range strings.Split(markdown, "\n") {
						markdownResult += "> " + line + "\n"
					}
					markdownResults = append(markdownResults, markdownResult)
				}
			})
		}

		// Copy the results to the clipboard
		clipboard.WriteAll(strings.Join(markdownResults, "\n"))
		fmt.Println("Results copied to clipboard.")
	}
}
