package main

import (
	"errors"
	"expvar"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var reqs = expvar.NewInt("Requests")
var errs = expvar.NewInt("Errors")

func work(log logrus.FieldLogger) error { // pretend work
	defer func(t time.Time) {
		log.WithField("duration", time.Since(t).Seconds()).Info("[work] complete")
	}(time.Now())

	s := rand.Intn(99) + 1 // 1..100
	time.Sleep(time.Duration(s) * time.Millisecond)

	var err error
	if s <= 25 { // ~25% of the time the work errors
		err = errors.New("OMG Error!")
		log.WithFields(logrus.Fields{
			"s": s,
		}).Error("[work] incomplete")
	}
	return err
}

func httpLogginghandler(log logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK // net/http returns 200 by default
		log = log.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.String(),
		})
		reqs.Add(1)
		defer func(t time.Time) {
			log.WithFields(logrus.Fields{
				"status":   status,
				"duration": time.Since(t).Seconds(),
			}).Info()
		}(time.Now())

		if err := work(log); err != nil {
			errs.Add(1)
			status = http.StatusBadRequest
			http.Error(w, "Nope", status)
			log.WithFields(logrus.Fields{
				"status": status,
			}).Error(err.Error())
			return
		}

		w.Write([]byte(`:-)`))
	}
}

func main() {

	// curried log
	log := logrus.WithField("app", "logs_lab_2")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", httpLogginghandler(log))

	log.WithField("port", port).Info("Listening at: http://localhost")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Errored with: " + err.Error())
	}
}
