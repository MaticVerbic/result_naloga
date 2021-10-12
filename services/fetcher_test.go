package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

var successResponses []string = []string{
	"Vrhunska ekipa za informacijsko optimizacijo poslovanja. Prinaša popoln nadzor nad celotnim poslovanjem, manjše stroške in premoč nad konkurenti.",
	"Naše rešitve so doma na šestih celinah in v več kot 100 državah sveta, že 30 let. Od globalnih multinacionalk, kot so SIXT, BOSCH in SCANIA, do super nišnih domačih igralcev.",
	// nekaj presledkov je nbsp
	"Vsebine, ki vam pomagajo do boljšega poslovanja.\u00a0Redno spremljanje\u00a0povzroča odlično počutje in\u00a0boljše poslovne rezultate.",
	"Če želiš reševati nerešljivo, sem ter tja narediti kak čudež, ustvariti vsaj za odtenek boljši svet in pomagati drugim vizionarjem in inovatorjem, potem se nam pridruži.",
}

func TestFetch(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	log := logrus.NewEntry(logger)

	var wg sync.WaitGroup
	in := &fetchInput{
		log: log,
		wg:  &wg,
	}

	tests := []struct {
		in          *fetchInput
		url         string
		output      string
		errExpected bool
	}{
		{
			in:     in,
			url:    "https://www.result.si/projekti",
			output: "Naše rešitve so doma na šestih celinah in v več kot 100 državah sveta, že 30 let. Od globalnih multinacionalk, kot so SIXT, BOSCH in SCANIA, do super nišnih domačih igralcev.",
		},
		{
			in:     in,
			url:    "https://www.result.si/o-nas/",
			output: "Vrhunska ekipa za informacijsko optimizacijo poslovanja. Prinaša popoln nadzor nad celotnim poslovanjem, manjše stroške in premoč nad konkurenti.",
		},
		{
			in:          in,
			url:         "test fail",
			errExpected: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := make(chan string, 1)
			errCh := make(chan error, 1)

			test.in.url = test.url
			test.in.out = out
			test.in.errCh = errCh

			test.in.wg.Add(1)
			fetch(test.in)
			test.in.wg.Wait()

			close(out)
			close(errCh)

			if test.errExpected {
				if len(errCh) == 0 {
					t.Log("should've thrown an error but doesn't")
					t.Fail()
				}

				if len(out) == 1 {
					t.Log("produces output but shouldn't")
					t.Fail()
				}

				return
			}

			if len(errCh) == 1 {
				t.Logf("returns error but shouldn't: %v", <-errCh)
				t.Fail()
			}

			if len(out) == 0 {
				t.Log("should have produced output, but didn't")
			}

			got := <-out

			if got != test.output {
				t.Logf("invalid output: got: %q expected: %q", got, test.output)
				t.Fail()
			}

		})
	}
}

func TestFetcherFetch(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	log := logrus.NewEntry(logger)
	f := NewFetcher(
		[]string{"https://www.result.si/projekti", "https://www.result.si/o-nas/", "https://www.result.si/kariera/", "https://www.result.si/blog/"},
		log,
	)

	response := struct {
		SuccessCount    int
		ErrorCount      int
		SuccessResponse []string
		ErrorResponse   []error
	}{
		SuccessCount:    4,
		ErrorCount:      0,
		SuccessResponse: successResponses,
		ErrorResponse:   []error{},
	}

	for i := 1; i <= 4; i++ {
		t.Run(fmt.Sprintf("number of workers: %d", i), func(t *testing.T) {
			ctx := context.Background()
			successCount, successResponse, errorCount, errorResponse := f.Fetch(ctx, i)

			if successCount != response.SuccessCount {
				t.Logf("invalid successCount: got: %d expected: %d", successCount, response.SuccessCount)
				t.Fail()
			}

			if errorCount != response.ErrorCount {
				t.Logf("invalid errorCount: got: %d expected: %d", errorCount, response.ErrorCount)
				t.Fail()
			}

			if len(errorResponse) > 0 {
				t.Logf("returned error responses, should return 0")
			}

			for _, expected := range response.SuccessResponse {
				found := false
				for _, got := range successResponse {
					if expected == got {
						found = true
					}
				}

				if !found {
					t.Logf("failed to find expected response: %q", expected)
					t.Fail()
				}
			}
		})
	}
}
