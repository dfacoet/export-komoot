package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pieterclaerhout/export-komoot/komoot"
	"github.com/pieterclaerhout/go-log"
	"github.com/pieterclaerhout/go-waitgroup"
)

func main() {

	log.PrintTimestamp = true
	log.PrintColors = true

	emailPtr := flag.String("email", "", "Your Komoot email address")
	passwordPtr := flag.String("password", "", "Your Komoot password")
	filterPtr := flag.String("filter", "", "Filter on the given name")
	formatPtr := flag.String("format", "gpx", "The format to export as: gpx or fit")
	toPtr := flag.String("to", "", "The path to export to")
	noIncrementalPtr := flag.Bool("no-incremental", false, "If specified, all data is redownloaded")
	concurrencyPtr := flag.Int("concurrency", 16, "The number of simultaneous downloads")
	flag.Parse()

	start := time.Now()
	defer func() { log.Info("Elapsed:", time.Since(start)) }()

	format := "gpx"
	if *formatPtr == "fit" {
		format = "fit"
	}

	client := komoot.NewClient(*emailPtr, *passwordPtr)

	userID, err := client.Login()
	log.CheckError(err)

	fullDstPath, _ := filepath.Abs(*toPtr)
	log.Info("Exporting:", *emailPtr)
	log.Info("       to:", fullDstPath)
	log.Info("   format:", format)

	log.Info("Komoot User ID:", userID)

	tours, resp, err := client.Tours(userID, *filterPtr)
	if len(tours) == 0 {
		log.Info("No tours need to be downloaded")
		return
	}

	log.Info("Found", len(tours), "planned tours")

	allTours := []komoot.Tour{}

	if *noIncrementalPtr == false {

		log.Info("Incremental download, checking what has changed")

		changedTours := []komoot.Tour{}

		for _, tour := range tours {

			allTours = append(allTours, tour)

			dstPath := filepath.Join(*toPtr, tour.Filename(format))
			if !fileExists(dstPath) {
				changedTours = append(changedTours, tour)
			}

		}

		tours = changedTours

		if len(tours) == 0 {
			log.Info("No tours need to be downloaded")
			return
		}

		log.Info("Found", len(tours), "which need to be downloaded")

	} else {
		allTours = tours
	}

	log.Info("Downloading with a concurrency of", *concurrencyPtr)
	wg := waitgroup.NewWaitGroup(*concurrencyPtr)

	var downloadCount int

	for _, tour := range tours {

		tourToDownload := tour
		label := fmt.Sprintf("%10d | %-7s | %-15s | %s", tour.ID, tour.Status, tour.FormattedSport(), tour.Name)

		wg.Add(func() {

			if err := func() error {

				r, err := client.Coordinates(tourToDownload)
				if err != nil {
					return err
				}

				var out []byte
				if format == "fit" {
					out, err = r.Fit()
					if err != nil {
						return err
					}
				} else {
					out = r.GPX()
				}

				deleteWithPattern(*toPtr, fmt.Sprintf("%d_*.*", tourToDownload.ID))

				dstPath := filepath.Join(*toPtr, tourToDownload.Filename(format))
				if err = saveTourFile(out, dstPath, tourToDownload); err != nil {
					return err
				}

				log.Info("Downloaded:", label)

				return nil

			}(); err != nil {
				log.Error("Downloaded:", label, "|", err)
			}
			downloadCount++

		})

	}

	wg.Wait()

	log.Info("Downloaded", downloadCount, "tours")

	log.Info("Saving tour list")
	dstPath := filepath.Join(*toPtr, "tours.json")
	err = saveFormattedJSON(resp, dstPath)
	log.CheckError(err)

	var out bytes.Buffer
	err = json.NewEncoder(&out).Encode(allTours)
	log.CheckError(err)

	log.Info("Saving parsed tour list")
	dstPath = filepath.Join(*toPtr, "tours_parsed.json")
	err = saveFormattedJSON(out.Bytes(), dstPath)
	log.CheckError(err)

}
