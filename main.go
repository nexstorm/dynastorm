package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RecordList struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
		Data       struct {
		} `json:"data"`
		Meta struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
	} `json:"result"`
}

type Record struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
		Data       struct {
		} `json:"data"`
		Meta struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
	} `json:"result"`
}

type ZoneList struct {
	Result []struct {
		ID                  string      `json:"id"`
		Name                string      `json:"name"`
		Status              string      `json:"status"`
		Paused              bool        `json:"paused"`
		Type                string      `json:"type"`
		DevelopmentMode     int         `json:"development_mode"`
		NameServers         []string    `json:"name_servers"`
		OriginalNameServers []string    `json:"original_name_servers"`
		OriginalRegistrar   string      `json:"original_registrar"`
		OriginalDnshost     interface{} `json:"original_dnshost"`
		ModifiedOn          time.Time   `json:"modified_on"`
		CreatedOn           time.Time   `json:"created_on"`
		ActivatedOn         time.Time   `json:"activated_on"`
		Meta                struct {
			Step                    int  `json:"step"`
			CustomCertificateQuota  int  `json:"custom_certificate_quota"`
			PageRuleQuota           int  `json:"page_rule_quota"`
			PhishingDetected        bool `json:"phishing_detected"`
			MultipleRailgunsAllowed bool `json:"multiple_railguns_allowed"`
		} `json:"meta"`
		Owner struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Email string `json:"email"`
		} `json:"owner"`
		Account struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
		Tenant struct {
			ID   interface{} `json:"id"`
			Name interface{} `json:"name"`
		} `json:"tenant"`
		TenantUnit struct {
			ID interface{} `json:"id"`
		} `json:"tenant_unit"`
		Permissions []string `json:"permissions"`
		Plan        struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			Price             int    `json:"price"`
			Currency          string `json:"currency"`
			Frequency         string `json:"frequency"`
			IsSubscribed      bool   `json:"is_subscribed"`
			CanSubscribe      bool   `json:"can_subscribe"`
			LegacyID          string `json:"legacy_id"`
			LegacyDiscount    bool   `json:"legacy_discount"`
			ExternallyManaged bool   `json:"externally_managed"`
		} `json:"plan"`
	} `json:"result"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

var (
	cfg = make(map[string]string)
	zD  net.Dialer
)

func ErrHDL(err error) {
	if err != nil {
		log.Println(err)
	}
}

func GetIP(client http.Client) string {
	url := "https://www.cloudflare.com/cdn-cgi/trace"
	resp, err := client.Get(url)
	ErrHDL(err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	ErrHDL(err)
	var ip string
	sps := string(b)
	splitted := strings.Fields(sps)
	if len(splitted) >= 2 && strings.Contains(splitted[2], "ip=") {
		ip = strings.Split(splitted[2], "ip=")[1]
	}
	return ip
}

func GetLastIP(domain string) string {
	ips, err := net.ResolveIPAddr("ip4", domain)
	ErrHDL(err)
	return ips.String()
}

func GetZoneID(email, api, domain string) string {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	url := "https://api.cloudflare.com/client/v4/zones?name=" + domain
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", api)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	ErrHDL(err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	ErrHDL(err)
	var z ZoneList
	err = json.Unmarshal(b, &z)
	ErrHDL(err)
	return z.Result[0].ID
}

func GetDNSRecordID(email, api, domain, zoneID string) string {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	url := "https://api.cloudflare.com/client/v4/zones/" + zoneID + "/dns_records?name=" + domain
	req, err := http.NewRequest("GET", url, nil)
	ErrHDL(err)
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", api)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	ErrHDL(err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	ErrHDL(err)
	var d RecordList
	err = json.Unmarshal(b, &d)
	ErrHDL(err)
	return d.Result[0].ID
}

func SplitSR(domain string) string {
	rdm, _ := publicsuffix.Domain(domain)
	return rdm
}

func UpdateIP(cf_email, cf_api, domain, zone_id, ip, record_id string) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	url := "https://api.cloudflare.com/client/v4/zones/" + zone_id + "/dns_records/" + record_id
	req, err := http.NewRequest("PUT", url, nil)
	ErrHDL(err)
	req.Header.Set("X-Auth-Email", cf_email)
	req.Header.Set("X-Auth-Key", cf_api)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(`{"type":"A","name":"` + domain + `","content":"` + ip + `","ttl":60,"proxied":false}`))
	resp, err := client.Do(req)
	ErrHDL(err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	ErrHDL(err)
	var u Record
	err = json.Unmarshal(b, &u)
	ErrHDL(err)
	if u.Success {
		log.Println("IP updated successfully.")
	} else {
		log.Println("IP update failed.")
	}
}

func main() {

	cf_email, cf_api, domain := cfg["Email"], cfg["API-key"], cfg["Domain"]
	if cf_email == "" || cf_api == "" || domain == "" {
		log.Fatalln("Email, API-key and Domain are required")
	}
	update_interval, err := strconv.Atoi(cfg["Interval"])
	if err != nil {
		log.Println("Interval not set, using default value 60s")
		update_interval = 60
	}
	fmt.Println("Account:", cf_email, "\nDomain:", domain, "\nUpdate interval:", update_interval)
	ErrHDL(err)
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return zD.DialContext(ctx, "tcp4", addr)
	}
	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
	rdm := SplitSR(domain)
	print(rdm)
	ip := GetIP(client)
	zone_id := GetZoneID(cf_email, cf_api, rdm)
	record_id := GetDNSRecordID(cf_email, cf_api, domain, zone_id)
	log.Println("\nDetected Root Domain:", rdm, "\nDetected Zone ID:", zone_id, "\nDetected Record ID:", record_id, "\nCurrent IP:", ip)
	UpdateIP(cf_email, cf_api, domain, zone_id, ip, record_id)
	log.Println("Updating IP every", update_interval, "seconds")
	timer := time.Tick(time.Duration(update_interval) * time.Second)
	for _ = range timer {
		log.Println("Connecting to Cloudflare...")
		ip := GetIP(client)
		log.Println("Current IP address:", ip)
		if ip != GetLastIP(domain) {
			log.Println("IP address changed, updating...")
			UpdateIP(cf_email, cf_api, domain, zone_id, ip, record_id)
		} else {
			log.Println("IP address not changed, skipping update...")
		}
	}
}
