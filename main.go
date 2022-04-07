package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/NYTimes/gziphandler"
	"github.com/oschwald/maxminddb-golang"
	"github.com/realclientip/realclientip-go"
)

//go:embed GeoLite2-City.mmdb.tar.gz
var mmdbTarBytes []byte

func main() {
	mmdbBytes, err := mmdbFromTarBall(mmdbTarBytes)
	if err != nil {
		log.Fatalf("Failed finding mmdb %v", err)
	}

	db, err := maxminddb.FromBytes(mmdbBytes)
	if err != nil {
		log.Fatalf("Failed opening maxminddb %v", err)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}
	http.Handle("/v1.json", gziphandler.GzipHandler(handleV1JSON(db)))
	log.Printf("Listening on %q", addr)
	err = http.ListenAndServe(addr, nil)
	log.Fatalf("listening on %q %v", addr, err)
}

func handleV1JSON(db *maxminddb.Reader) http.HandlerFunc {
	strategy, err := realclientip.NewRightmostTrustedCountStrategy("X-Forwarded-For", 2)
	if err != nil {
		panic(fmt.Errorf("failed creating realclientip strat %v", err))
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		clientIP := strategy.ClientIP(r.Header, r.RemoteAddr)
		var location Location
		if err = db.Lookup(net.ParseIP(clientIP), &location); err != nil {
			errBytes, _ := json.Marshal(map[string]string{
				"error": fmt.Sprintf("looking up client ip %q", clientIP),
			})
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(errBytes)
			return
		}
		var jsonBytes []byte
		if r.URL.Query().Has("pretty") {
			jsonBytes, _ = json.MarshalIndent(location, "", "  ")
		} else {
			jsonBytes, _ = json.Marshal(location)
		}
		rw.Write(jsonBytes)
	}
}

func mmdbFromTarBall(mmdbBytes []byte) ([]byte, error) {
	gzf, err := gzip.NewReader(bytes.NewReader(mmdbBytes))
	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(gzf)
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag == tar.TypeReg && path.Base(header.Name) == "GeoLite2-City.mmdb" {
			data := make([]byte, header.Size)
			if _, err := io.ReadFull(tr, data); err != nil {
				return nil, err
			}
			return data[:], nil
		}
	}
	return nil, errors.New("mmdb not found in tarball")
}

// Location captures the relevant data from the GeoIP database lookup.
type Location struct {
	City struct {
		Names Names `maxminddb:"names" json:"names"`
	} `maxminddb:"city" json:"city"`
	Continent struct {
		Names Names `maxminddb:"names" json:"names"`
	} `maxminddb:"continent" json:"continent"`
	Country struct {
		IsoCode string `maxminddb:"iso_code"  json:"iso_code" `
	} `maxminddb:"country" json:"country"`
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius" json:"accuracy_radius"`
		Latitude       float64 `maxminddb:"latitude" json:"latitude"`
		Longitude      float64 `maxminddb:"longitude" json:"longitude"`
		MetroCode      uint    `maxminddb:"metro_code" json:"metro_code"`
		TimeZone       string  `maxminddb:"time_zone" json:"time_zone"`
	} `maxminddb:"location" json:"location"`
	Postal struct {
		Code string `maxminddb:"code" json:"code"`
	} `maxminddb:"postal" json:"postal"`
	Subdivisions []struct {
		IsoCode string `maxminddb:"iso_code" json:"iso_code"`
	} `maxminddb:"subdivisions" json:"subdivisions"`
}

// Names could be unmarshalled into a map[string]string, but only returning the en names for now.
type Names struct {
	EN string `maxminddb:"en" json:"en"`
}
