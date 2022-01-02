package cmd

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/binaryfigments/ipreport/libs/ptr"
	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// nucleiCmd represents the bulk command
var nucleiCmd = &cobra.Command{
	Use:   "nuclei",
	Short: "Get IP addresses from nuclei output and add information to it.",
	Long:  `Get IP addresses from nuclei output and add information to it.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.SetOutput(os.Stderr)
		log.SetLevel(logrus.InfoLevel)

		// get nameserver from flags
		nameserverFlag, _ := cmd.Flags().GetString("nameserver")
		log.WithFields(logrus.Fields{
			"nameserver": nameserverFlag,
		}).Info("Nameserver flag")

		fileFlag, _ := cmd.Flags().GetString("file")

		log.WithFields(logrus.Fields{
			"info": "MaxMind",
		}).Info("This product includes GeoLite2 data created by MaxMind, available from www.maxmind.com")

		// open file
		file, err := os.Open(fileFlag)
		if err != nil {
			log.WithFields(logrus.Fields{
				"file": fileFlag,
			}).Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		var ipaddresses []string
		for scanner.Scan() {
			words := strings.Fields(scanner.Text())
			u, err := url.Parse(words[5])
			if err != nil {
				log.WithFields(logrus.Fields{
					"urlparse": words[5],
				}).Fatal(err)
			}
			host, _, _ := net.SplitHostPort(u.Host)
			ipaddresses = append(ipaddresses, host)
		}
		file.Close()

		fmt.Printf("IP,ASN,ASN-Org,CC,Country,Sub-Division,City\n")

		var wg sync.WaitGroup
		for _, ipaddress := range ipaddresses {
			wg.Add(1)
			go transferWorkerNuclei(ipaddress, nameserverFlag, &wg)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(nucleiCmd)
	nucleiCmd.Flags().String("file", "nuclei.txt", "File with nuclei output.")
}

func transferWorkerNuclei(ip string, nameserverFlag string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Get PTR record
	ptrRecord, err := ptr.GetOne(ip, nameserverFlag)
	if err != nil {
		log.WithFields(logrus.Fields{
			"IP":  ip,
			"PTR": ptrRecord,
		}).Warn(err)
	}

	// Get GeoIP ASN information
	asnbd, err := geoip2.Open("geoip/GeoLite2-ASN.mmdb")
	if err != nil {
		log.WithFields(logrus.Fields{
			"ASN1": "error",
		}).Fatal(err)
	}
	defer asnbd.Close()
	asnr, err := asnbd.ASN(net.ParseIP(ip))
	if err != nil {
		log.WithFields(logrus.Fields{
			"ASN2": "error",
		}).Fatal(err)
	}

	// Get GeoIP City information
	citydb, err := geoip2.Open("geoip/GeoLite2-City.mmdb")
	if err != nil {
		log.WithFields(logrus.Fields{
			"City1": "error",
		}).Fatal(err)
	}
	defer citydb.Close()
	cr, err := citydb.City(net.ParseIP(ip))
	if err != nil {
		log.WithFields(logrus.Fields{
			"City2": "error",
		}).Fatal(err)
	}
	var subDivision string
	if len(cr.Subdivisions) > 0 {
		subDivision = cr.Subdivisions[0].Names["en"]
	}

	// Return a CSV line here.
	fmt.Printf("%s,%v,\"%s\",%s,\"%s\",\"%s\",\"%s\"\n", ip, asnr.AutonomousSystemNumber, asnr.AutonomousSystemOrganization, cr.Country.IsoCode, cr.Country.Names["en"], subDivision, cr.City.Names["en"])
}
