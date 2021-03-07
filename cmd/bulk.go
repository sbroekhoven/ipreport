package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/binaryfigments/ipreport/libs/ptr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type reportIP struct {
	IP  string `json:",omitempty"`
	PTR string `json:",omitempty"`
}

// bulkCmd represents the bulk command
var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Add information about an list of IP addresses and print CSV output.",
	Long:  `Add information about an list of IP addresses and print CSV output.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(logrus.Fields{}).Info("Starting")

		// get nameserver from flags
		nameserverFlag, _ := cmd.Flags().GetString("nameserver")
		log.WithFields(logrus.Fields{
			"nameserver": nameserverFlag,
		}).Info("Nameserver Flag")

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
		var ipaddresses []string
		for scanner.Scan() {
			ipaddresses = append(ipaddresses, scanner.Text())
		}
		file.Close()

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
	// start here

	ptrRecord, err := ptr.GetOne(ip, nameserverFlag)
	if err != nil {
		// TODO: build in a retry (in PTR function above)
		log.WithFields(logrus.Fields{
			"IP":  ip,
			"PTR": ptrRecord,
		}).Warn(err)
	}
	fmt.Printf("%s, %s\n", ip, ptrRecord)

}
