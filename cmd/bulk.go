package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/oschwald/geoip2-golang"
	"github.com/sbroekhoven/ipreport/libs/ptr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// bulkCmd represents the bulk command
var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Add information about an list of IP addresses and print CSV output.",
	Long:  `Add information about an list of IP addresses and print CSV output.`,
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
			ipaddresses = append(ipaddresses, scanner.Text())
		}
		file.Close()

		fmt.Printf("IP,PTR,ASN,ASN-Org,CC,Country,Sub-Division,City\n")

		var wg sync.WaitGroup
		for _, ipaddress := range ipaddresses {
			wg.Add(1)
			go transferWorker(ipaddress, nameserverFlag, &wg)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(bulkCmd)
	bulkCmd.Flags().String("file", "ips.txt", "File with IP addresses to use.")
}

func transferWorker(ip string, nameserverFlag string, wg *sync.WaitGroup) {
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
	fmt.Printf("%s,%s,%v,\"%s\",%s,\"%s\",\"%s\",\"%s\"\n", ip, ptrRecord, asnr.AutonomousSystemNumber, asnr.AutonomousSystemOrganization, cr.Country.IsoCode, cr.Country.Names["en"], subDivision, cr.City.Names["en"])
}
