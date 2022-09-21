package main

import (
	"flag"
	"github.com/nexstorm/dynastorm/config"
	"github.com/nexstorm/dynastorm/tools"
	"log"
	"time"
)

var (
	cfg = make(map[string]string)
)

func ErrHDL(err error) {
	if err != nil {
		log.Println(err)
	}
}

func main() {
	cfgptr := flag.String("c", "config.yaml", "config file")
	flag.Parse()
	err, cf_email, cf_api, domain, update_interval := config.Parseconfig(*cfgptr)
	ErrHDL(err)
	client := config.NewClient()
	ip := tools.GetIP(client)
	zone_id := tools.GetZoneID(cf_email, cf_api, tools.SplitSR(domain), client)
	record_id := tools.GetDNSRecordID(cf_email, cf_api, domain, zone_id)
	log.Println("\nDetected Zone ID:", zone_id, "\nDetected Record ID:", record_id, "\nCurrent IP:", ip)
	tools.UpdateIP(cf_email, cf_api, domain, zone_id, ip, record_id, client)
	log.Println("Updating IP every", update_interval, "seconds")
	timer := time.Tick(time.Duration(update_interval) * time.Second)
	for _ = range timer {
		log.Println("Connecting to Cloudflare...")
		ip := tools.GetIP(client)
		log.Println("Current IP address:", ip)
		if ip != tools.GetLastIP(domain) {
			log.Println("IP address changed, updating...")
			tools.UpdateIP(cf_email, cf_api, domain, zone_id, ip, record_id, client)
		} else {
			log.Println("IP address not changed, skipping update...")
		}
	}
}
