package config

type Config struct {
	Port       string `envconfig:"PORT"`
	MetricPort string `envconfig:"METRICS_ADDR"`
}
