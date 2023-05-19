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
	PageToMatch    string `kong:"help='Page to match',alias='page',short='p'"`
}

type feature struct {
	Title     string
	Body      string
	Category  string
	PageTitle string
}

func main() {

	// Parse command line arguments
	kong.Parse(&cli)

	// Use the command line arguments
	eclipseVersion := cli.EclipseVersion
	titleToMatch := cli.TitleToMatch
	pageToMatch := cli.PageToMatch

	// URLs of the subPages
	subPageURLs := []string{
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/platform.php",
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/jdt.php",
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/platform_isv.php",
		"https://www.eclipse.org/eclipse/news/" + eclipseVersion + "/pde.php",
	}

	matchedFeatures := make([]feature, 0)

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
		pageTitle := ""
		currentCategory := ""

		// Find the features
		doc.Find("h2,td.title,td.content").Each(func(i int, s *goquery.Selection) {
			if s.Is("h2") {
				if pageTitle == "" {
					pageTitle = s.Text()
					fmt.Printf("Processing page '%s'\n", pageTitle)
				} else {
					currentCategory = s.Text()
					fmt.Printf("Found category '%s'\n", currentCategory)
				}
			} else if s.Is("td.title") {
				featureTitle := s.Text()
				fmt.Printf("Found feature title '%s'\n", featureTitle)
				if strings.Contains(strings.ToLower(featureTitle), strings.ToLower(titleToMatch)) {
					if strings.Contains(strings.ToLower(pageTitle), strings.ToLower(pageToMatch)) {
						nextSibling := s.Next()
						if nextSibling.Is("td.content") {
							featureBody := nextSibling.Text()
							converter := md.NewConverter("", true, nil)
							markdown, _ := converter.ConvertString(featureBody)
							matchedFeatures = append(matchedFeatures, feature{
								Title:     featureTitle,
								Body:      markdown,
								Category:  currentCategory,
								PageTitle: pageTitle,
							})
						}
					}
				}
			}
		})
	}

	if len(matchedFeatures) > 1 {
		// If more than one title matched, just print the titles
		for _, title := range matchedFeatures {
			fmt.Printf("'%s - %s from '%s'\n", title.Title, title.Category, title.PageTitle)
		}
	} else if len(matchedFeatures) == 1 {
		// If exactly one title matched, print the title and body
		markdownResults := make([]string, 0)
		feature := matchedFeatures[0]
		body := feature.Body

		// Convert to markdown
		converter := md.NewConverter("", true, nil)
		markdown, _ := converter.ConvertString(body)

		// Print the title and body in markdown format
		markdownResult := "> ## " + feature.Title + "\n"
		for _, line := range strings.Split(markdown, "\n") {
			markdownResult += "> " + line + "\n"
		}
		markdownResults = append(markdownResults, markdownResult)

		// Copy the results to the clipboard
		clipboard.WriteAll(strings.Join(markdownResults, "\n"))
		fmt.Println("Results copied to clipboard.")
	} else {
		fmt.Println("No results found.")
	}
}
