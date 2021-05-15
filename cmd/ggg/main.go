package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"

	_ "github.com/mattn/go-sqlite3"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-github/v35/github"
)

func main() {
	if err := run(); err != nil {
		log.Printf("runtime error: %s", err.Error())
		os.Exit(1)
	}
}

type imageInfo struct {
	sourceURL string
	canonURL  string
	alt       string
}

func run() error {
	// prelude
	ctx := context.Background()

	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	))

	client := github.NewClient(oauthClient)

	fromPage, _ := strconv.Atoi(os.Getenv("FROM_PAGE"))
	fromIndex, _ := strconv.Atoi(os.Getenv("FROM_INDEX"))

	// database
	//fp, err := os.Create("./data/gophers.sqlite3")
	//if err != nil {
	//	return fmt.Errorf("error creating db, %w", err)
	//}
	//fp.Close()

	db, err := sql.Open("sqlite3", "./data/gophers.sqlite3")
	if err != nil {
		return fmt.Errorf("error opening db, %w", err)
	}
	defer db.Close()

	// logic
	for page := fromPage; page < 1000; page++ {
		fmt.Print("+")

		repos, err := getGoReposURLs(client, page)
		if err != nil {
			return fmt.Errorf("failed get repos, %w", err)
		}

		for idx, repo := range repos {
			fmt.Print("-")

			if idx < fromIndex {
				continue
			}
			if err = dbInsertRepo(db, repo, page, idx); err != nil {
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
					return fmt.Errorf("insert repo error, %w", err)
				}
			}

			sources, err := scrapeRepoReadmeImageSources(*repo.HTMLURL)
			if err != nil {
				return fmt.Errorf("failed scrape images, %w", err)
			}

			for _, source := range sources {
				if err = dbInsertSource(db, repo.GetFullName(), source.sourceURL, source.canonURL, source.alt); err != nil {
					if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
						return fmt.Errorf("insert image error, %w", err)
					}
				}
			}

			time.Sleep(time.Second / 20)
		}

		fmt.Println()
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

func dbInsertRepo(db *sql.DB, repo *github.Repository, page, pageIndex int) error {
	stmt, err := db.Prepare("INSERT INTO repos(repo_name, repo_url, page, page_index) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(repo.GetFullName(), repo.GetHTMLURL(), page, pageIndex); err != nil {
		return err
	}

	return nil
}

func dbInsertSource(db *sql.DB, repoName, sourceURL, canonURL, alt string) error {
	stmt, err := db.Prepare("INSERT INTO sources(repo_name, url, domain, ext, alt, canon_url, canon_domain, canon_ext) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	uS, _ := url.Parse(sourceURL)
	if uS == nil {
		uS = new(url.URL)
	}
	uC, _ := url.Parse(canonURL)
	if uC == nil {
		uC = new(url.URL)
	}

	if _, err = stmt.Exec(repoName, sourceURL, uS.Host, path.Ext(sourceURL), alt, canonURL, uC.Host, path.Ext(canonURL)); err != nil {
		return err
	}

	return nil
}

func scrapeRepoReadmeImageSources(repoURL string) ([]imageInfo, error) {
	res, err := http.Get(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed making repo page request, %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	images := make([]imageInfo, 0, 1)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed parsing html body, %w", err)
	}

	doc.Find("readme-toc img").Each(func(idx int, s *goquery.Selection) {
		source, _ := s.Attr("src")
		if source[0] == '/' {
			source = "https://github.com" + source
		}
		alt, _ := s.Attr("alt")
		canon, _ := s.Attr("data-canonical-src")

		images = append(images, imageInfo{
			sourceURL: source,
			canonURL:  canon,
			alt:       alt,
		})
	})

	return images, nil
}
