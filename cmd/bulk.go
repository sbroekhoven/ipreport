package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/binaryfigments/dnstransfer/libs/axfr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// bulkCmd represents the bulk command
var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Add information about an list of IP addresses and print CSV output.",
	Long:  `Add information about an list of IP addresses and print CSV output.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bulk called")

		// get nameserver from flags
		nameserverFlag, _ := cmd.Flags().GetString("nameserver")
		fileFlag, _ := cmd.Flags().GetString("file")

		// open file
		file, err := os.Open(fileFlag)
		if err != nil {
			log.WithFields(logrus.Fields{
				"file": fileFlag,
			}).Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		var domains []string
		for scanner.Scan() {
			domains = append(domains, scanner.Text())
		}
		file.Close()

		var wg sync.WaitGroup
		for _, domain := range domains {
			wg.Add(1)
			go transferWorker(domain, nameserverFlag, &wg)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(bulkCmd)
	bulkCmd.Flags().String("file", "ips.txt", "File with IP addresses to use.")
	// bulkCmd.PersistentFlags().String("foo", "", "A help for foo")
	// bulkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func transferWorker(domain string, nameserverFlag string, wg *sync.WaitGroup) {
	defer wg.Done()
	// start here
	// Get nameservers
	var domainnameservers string
	err := retry(2, 2*time.Second, func() (err error) {
		domainnameservers, err = ptr.getOne(domain, nameserverFlag)
		return
	})
	// domainnameservers, err := ns.Get(domain, nameserverFlag)
	if err != nil {
		// TODO: build in a retry (in ns function above)
		log.WithFields(logrus.Fields{
			"error":        err,
			"domain":       domain,
			"nameserver":   nameserverFlag,
			"transferable": false,
		}).Warn("Get Nameservers")
	}

	for _, domainnameserver := range domainnameservers {
		// DNS RCODEs: http://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-6
		axfrdata, err := axfr.Get(domain, domainnameserver)
		if err != nil {
			// Check if bad xfr code
			if strings.Contains(err.Error(), "bad xfr rcode") {
				log.WithFields(logrus.Fields{
					"error":        err,
					"domain":       domain,
					"nameserver":   domainnameserver,
					"transferable": false,
				}).Info("Zone transfer failed")
			}
			// strings.Contains(err, "tcp"):
			if strings.Contains(err.Error(), "red tcp") {
				log.WithFields(logrus.Fields{
					"error":        err,
					"domain":       domain,
					"nameserver":   domainnameserver,
					"transferable": false,
				}).Info("TCP timeout on port 53")
			}
		}
		if len(axfrdata.Records) > 0 {
			log.WithFields(logrus.Fields{
				"error":        err,
				"domain":       domain,
				"nameserver":   domainnameserver,
				"transferable": true,
			}).Error("Zone can be transfered!")
		}
	}
}
