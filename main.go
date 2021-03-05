package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pieterclaerhout/export-komoot/komoot"
	"github.com/pieterclaerhout/go-log"
	"github.com/pieterclaerhout/go-waitgroup"
)

func main() {

	log.PrintTimestamp = true
	log.PrintColors = true

	emailPtr := flag.String("email", "", "Your Komoot email address")
	passwordPtr := flag.String("password", "", "Your Komoot password")
	toPtr := flag.String("to", "", "The path to export to")
	concurrencyPtr := flag.Int("concurrency", 16, "The number of simultaneous downloads")
	flag.Parse()

	log.Info("Exporting:", *emailPtr, "to:", *toPtr)

	client := komoot.NewClient(*emailPtr, *passwordPtr)

	userID, err := client.Login()
	log.CheckError(err)

	log.Info("> Komoot User ID:", userID)

	tours, resp, err := client.Tours(userID)
	log.Info("> Found", len(tours), "planned tours")

	log.Info("> Downloading with a concurrency of", *concurrencyPtr)
	wg := waitgroup.NewWaitGroup(*concurrencyPtr)

	for _, tour := range tours {

		tourToDownload := tour

		wg.Add(func() {

			log.Info("> Downloading:", tourToDownload.ID, "|", tourToDownload.Name)

			gpx, err := client.Download(int(tourToDownload.ID))
			if err != nil {
				log.Error(err)
				return
			}

			dstPath := filepath.Join(*toPtr, fmt.Sprintf("%d.gpx", tourToDownload.ID))

			err = ioutil.WriteFile(dstPath, []byte(gpx), 0755)
			if err != nil {
				log.Error(err)
				return
			}

		})

	}

	wg.Wait()

	log.Info("> Saving tour list")

	var out bytes.Buffer
	err = json.Indent(&out, resp, "", "  ")
	log.CheckError(err)

	dstPath := filepath.Join(*toPtr, "tours.json")
	err = ioutil.WriteFile(dstPath, out.Bytes(), 0755)
	log.CheckError(err)

}
