package cmd

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Retrying after error")

		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
