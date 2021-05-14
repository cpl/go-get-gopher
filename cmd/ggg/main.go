package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v35/github"
)

func main() {
	if err := run(); err != nil {
		log.Printf("runtime error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	))

	client := github.NewClient(oauthClient)

	fromPage, _ := strconv.Atoi(os.Getenv("FROM_PAGE"))
	fromIndex, _ := strconv.Atoi(os.Getenv("FROM_INDEX"))

	for page := fromPage; page < 1000; page++ {
		repos, err := getGoReposURLs(client, page)
		if err != nil {
			return fmt.Errorf("failed get repos, %w", err)
		}

		for idx, repo := range repos {
			if idx < fromIndex {
				continue
			}

			sources, err := scrapeRepoReadmeImageSources(*repo.HTMLURL)
			if err != nil {
				return fmt.Errorf("failed scrape images, %w", err)
			}

			for _, source := range sources {
				fmt.Println(source)
			}

			time.Sleep(time.Second / 4)
		}
	}

	return nil
}

func getGoReposURLs(client *github.Client, page int) ([]*github.Repository, error) {
	ctx := context.Background()
	result, _, err := client.Search.Repositories(ctx, "language:Go", &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, err
	}

	return result.Repositories, nil
}

func scrapeRepoReadmeImageSources(repoURL string) ([]string, error) {
	res, err := http.Get(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed making repo page request, %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	images := make([]string, 0)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed parsing html body, %w", err)
	}

	doc.Find("readme-toc img").Each(func(idx int, s *goquery.Selection) {
		source, _ := s.Attr("src")
		if source[0] == '/' {
			source = "https://github.com" + source
		}
		images = append(images, source)
	})

	return images, nil
}
