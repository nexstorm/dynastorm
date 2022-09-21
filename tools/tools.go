package tools

import (
	"encoding/json"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type ZoneList struct {
	Result []struct {
		ID string `json:"id"`
	} `json:"result"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

type RecordList struct {
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
	Result  []struct {
		ID string `json:"id"`
	} `json:"result"`
}

type Record struct {
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
}

func ErrHDL(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func SplitSR(domain string) string {
	rdm, _ := publicsuffix.Domain(domain)
	return rdm
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

func GetDNSRecordID(email, api, domain, zoneID string) string {
	var d RecordList
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
	err = json.Unmarshal(b, &d)
	ErrHDL(err)
	if !d.Success {
		log.Fatal(d.Errors)
	}
	return d.Result[0].ID
}

func GetZoneID(email, api, domain string, client http.Client) string {
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
	if !z.Success {
		log.Fatal(z.Errors)
	}
	return z.Result[0].ID
}

func UpdateIP(cf_email, cf_api, domain, zone_id, ip, record_id string, client http.Client) {
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

func GetLastIP(domain string) string {
	ips, err := net.ResolveIPAddr("ip4", domain)
	ErrHDL(err)
	return ips.String()
}
