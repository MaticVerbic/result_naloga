package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// Fetcher is responsible for spawning the correct amount of workers and then parsing the resulting html.
type Fetcher struct {
	Log     *logrus.Entry
	Urls    []string
	Results []string
}

func NewFetcher(urls []string, log *logrus.Entry) *Fetcher {
	return &Fetcher{
		Log:  log,
		Urls: urls,
	}
}

// Fetch is the main method supporting worker spawning.
func (f *Fetcher) Fetch(ctx context.Context, numWorker int) (int, []string, int, []error) {
	f.Log.WithField("numWorkers", numWorker).Info("initiated Fetch method")
	out := make(chan string, len(f.Urls))
	errCh := make(chan error, len(f.Urls))
	var wg sync.WaitGroup

	for i, url := range f.Urls {
		if i%numWorker == 0 && i != 0 {
			f.Log.WithField("routineCounter", i).Info("waiting")
			wg.Wait()
		}

		f.Log.WithField("routineCounter", i).Info("starting routine")
		wg.Add(1)
		go fetch(fetchInput{url: url, out: out, errCh: errCh, wg: &wg, log: f.Log})
	}

	wg.Wait()
	close(out)
	close(errCh)

	successCount := 0
	successResponse := []string{}
	errorCount := 0
	errorResponse := []error{}

	for item := range out {
		f.Log.Info("appending result")
		f.Results = append(f.Results, item)
		successResponse = append(successResponse, item)
		successCount++
	}

	for err := range errCh {
		f.Log.Errorf("error fetching: %v", err)
		errorResponse = append(errorResponse, err)
		errorCount++
	}

	return successCount, successResponse, errorCount, errorResponse
}

type fetchInput struct {
	url   string
	out   chan string
	errCh chan error
	wg    *sync.WaitGroup
	log   *logrus.Entry
}

func fetch(in fetchInput) {
	defer in.wg.Done()

	response, err := http.Get(in.url)
	if err != nil {
		in.errCh <- err
		return
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		in.errCh <- fmt.Errorf("response returned invalid code: %q, %d", in.url, response.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		in.errCh <- fmt.Errorf("failed to read html: %q", in.url)
	}

	var text []string

	doc.Find(".et_pb_header_content_wrapper").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		if s.Find("p").Text() != "" {
			text = append(text, s.Find("p").Text())
		}
	})

	if len(text) == 0 {
		doc.Find(".et_pb_fullwidth_header_subhead").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the title
			if s.Text() != "" {
				text = append(text, s.Text())
			}

		})
	}

	if len(text) > 0 {
		in.out <- text[0]
		return
	}

	in.errCh <- fmt.Errorf("failed to fetch any results for url: %q", in.url)
}
