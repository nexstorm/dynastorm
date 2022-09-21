package config

import (
	"context"
	"errors"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	cfg = make(map[string]string)
	zD  net.Dialer
)

func Parseconfig(cfgpath string) (error, string, string, string, int) {
	configf, err := os.ReadFile(cfgpath)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(configf, &cfg)
	cf_email, cf_api, domain := cfg["Email"], cfg["API-key"], cfg["Domain"]
	if cf_email == "" || cf_api == "" || domain == "" {
		err = errors.New("config file is not valid")
	}
	update_interval, err := strconv.Atoi(cfg["Interval"])
	if err != nil {
		log.Println("Interval not set, using default value 60s")
		update_interval = 60
	}
	log.Println("Account:", cf_email, "\nDomain:", domain, "\nUpdate interval:", update_interval)
	return err, cf_email, cf_api, domain, update_interval
}

func NewClient() http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return zD.DialContext(ctx, "tcp4", addr)
	}
	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
	return client
}
