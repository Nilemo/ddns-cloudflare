package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-co-op/gocron/v2"
)

func main() {
	zone := cloudflare.ZoneIdentifier(os.Getenv("ZONE_ID"))
	println(os.Getenv("API_TOKEN"))
	println(os.Getenv("CLOUDFLARE_EMAIL"))
	recordId1 := os.Getenv("DNS_RECORD_ID_1")
	recordId2 := os.Getenv("DNS_RECORD_ID_2")

	client, err := cloudflare.New(os.Getenv("CLOUDFLARE_APIKEY"), os.Getenv("CLOUDFLARE_EMAIL"))
	// client, err := cloudflare.NewWithAPIToken(os.Getenv("API_TOKEN"))
	if err != nil {
		println(err)
	}

	record1, err := getDNSRecord(client, zone, recordId1)
	if err != nil {
		println("error en getDNS 1")
		println(err)
	}

	record2, err := getDNSRecord(client, zone, recordId2)
	if err != nil {
		println("error en getDNS 2")
		println(err)
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		fmt.Println("Pucha, algo fallo")
		println(err)
	}

	j, err := s.NewJob(
		gocron.DurationJob(
			5*time.Minute,
		),
		gocron.NewTask(

			func(client *cloudflare.API, record1 cloudflare.DNSRecord, record2 cloudflare.DNSRecord, zone *cloudflare.ResourceContainer) {

				err1 := UpdateDNS(client, record1, zone)
				if err1 != nil {
					println("error en UpdateDNS")
					println(err1)
				}

				err2 := UpdateDNS(client, record2, zone)
				if err2 != nil {
					println("error en UpdateDNS")
					println(err2)
				}
			}, client, record1, record2, zone,
		),
	)
	if err != nil {
		fmt.Println("Pucha, algo fallo")
		println(err)
	}

	fmt.Println(j.ID())

	s.Start()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown

	if err := s.Shutdown(); err != nil {
		fmt.Printf("Error shutting down scheduler: %v\n", err)
	}
}

func getDNSRecord(api *cloudflare.API, zone *cloudflare.ResourceContainer, recordId string) (cloudflare.DNSRecord, error) {
	record, err := api.GetDNSRecord(context.Background(), zone, recordId)
	if err != nil {
		return record, err
	}

	return record, nil
}

func UpdateDNS(client *cloudflare.API, record cloudflare.DNSRecord, zone *cloudflare.ResourceContainer) error {
	ip := getip2()
	println(ip)
	recordResponse, err := client.UpdateDNSRecord(
		context.Background(),
		zone,
		cloudflare.UpdateDNSRecordParams{
			Content: ip,
			ID:      record.ID,
		},
	)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", recordResponse)
	return nil
}

func getip2() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return err.Error()
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	var ip IP
	json.Unmarshal(body, &ip)

	return ip.Query
}

type IP struct {
	Query string
}
