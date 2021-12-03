package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/ingridhq/comptest/_example/binary/config"
)

func main() {
	log.Println("START")
	var cfg config.Config
	envconfig.MustProcess("", &cfg)

	mainMux := http.NewServeMux()
	metricsMux := http.NewServeMux()
	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Potato endpoint")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Potato"))
	})

	metricsMux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Readiness endpoint")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready!!!"))
	})

	go func() {
		// Sleep to simulate system is not ready to use
		time.Sleep(time.Second * 1)
		log.Println("Start metrics server")
		if err := http.ListenAndServe(cfg.MetricPort, metricsMux); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Start main server")
	if err := http.ListenAndServe(cfg.Port, mainMux); err != nil {
		log.Fatal(err)
	}
}
