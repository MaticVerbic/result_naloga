package server

import (
	"context"
	"encoding/json"
	"fmt"
	"naloga/services"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type key int

const (
	LogKey key = iota
)

type Server struct {
	FetcherSvc *services.Fetcher
	Log        *logrus.Entry
}

func New(fetcherSvc *services.Fetcher, log *logrus.Entry) *Server {
	return &Server{
		FetcherSvc: fetcherSvc,
		Log:        log,
	}
}

type Response struct {
	SuccessCount    int      `json:"successCount"`
	ErrorCount      int      `json:"errorCount"`
	SuccessResponse []string `json:"successResponse"`
	ErrorResponse   []error  `json:"errorResponse"`
}

func (s *Server) Result(w http.ResponseWriter, r *http.Request) {
	workers, ok := r.URL.Query()["workers"]
	if !ok || len(workers[0]) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid input parameters, workers should be between 1 and 4"))
		return
	}

	nw := workers[0]

	numWorkers, err := strconv.Atoi(nw)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - failed to convert input to integer"))
	}

	ctx := r.Context()
	data, ok := ctx.Value(LogKey).(logrus.Fields)
	if !ok {
		s.Log.Error("failed to fetch logger from context")
	}

	s.Log.Data = data

	successCount, successResponse, errorCount, errorResponse := s.FetcherSvc.Fetch(ctx, numWorkers)

	resp := Response{
		SuccessCount:    successCount,
		ErrorCount:      errorCount,
		SuccessResponse: successResponse,
		ErrorResponse:   errorResponse,
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - failed to convert response to json"))
	}

	s.Log.Info("success")
	w.Write(b)
}

// Serve creates HTTP server.
// It also catches different SIG__ signals.
func (s *Server) Serve() {
	router := chi.NewRouter()

	// Middlewares
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(HTTPLog)
	router.Use(NewCORS())

	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	router.HandleFunc("/result", s.Result)

	errChan := make(chan error, 1)
	defer close(errChan)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		port := viper.GetString("PORT")
		if port == "" {
			port = "80"
		}
		s.Log.WithFields(logrus.Fields{"transport": "http", "state": "listening"}).Info("http init")
		errChan <- http.ListenAndServe(":"+port, router)
	}()

	s.Log.WithFields(logrus.Fields{"transport": "http", "state": "terminated"}).Error(<-errChan)
	os.Exit(1)
}

// NewCORS returns Cors struct
func NewCORS() func(next http.Handler) http.Handler {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Core-Api-Key", "X-Core-Accept-Currency", "X-Core-Accept-Language", "X-Accept-Locale", "X-Active-Role"},
		AllowCredentials: true,
	})
	return cors.Handler
}

// HTTPLog adds execution timing and request logging to http.Handler.
var HTTPLog = func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := time.Now()

		requestID := middleware.GetReqID(r.Context())
		if requestID == "" {
			return
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		log := logrus.WithFields(logrus.Fields{
			"request_id": requestID,
		})

		fields := logrus.Fields{
			"http_scheme": scheme,
			"http_proto":  r.Proto,
			"http_method": r.Method,
			"remote_addr": r.RemoteAddr,
			"request_id":  requestID,
			"user_agent":  r.UserAgent(),
		}

		log.WithFields(fields).Info("http request")

		defer func(s time.Time, log *logrus.Entry) {
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"elapsed":    time.Since(s),
			}).Info("http request processed")
		}(s, log)

		r = r.WithContext(context.WithValue(r.Context(), LogKey, log.Data))

		h.ServeHTTP(w, r)
	})
}
