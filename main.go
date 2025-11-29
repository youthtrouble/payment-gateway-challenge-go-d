package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

//	@title			Payment Gateway API
//	@version		1.0
//	@description	A payment gateway API that allows merchants to process card payments and retrieve payment details.
//	@description	The gateway validates requests, communicates with an acquiring bank, and stores payment information.
//	@description
//	@description	## Payment Status
//	@description	- **Authorized**: Payment was approved by the bank
//	@description	- **Declined**: Payment was declined by the bank
//	@description	- **Rejected**: Payment was rejected due to validation errors (never sent to bank)
//	@description
//	@description	## Security
//	@description	- Only the last 4 digits of card numbers are stored and returned
//	@description	- CVV is never stored, only sent to the bank
//	@description
//	@description	## Supported Currencies
//	@description	USD, GBP, EUR

//	@contact.name	API Support
//	@contact.url	https://github.com/cko-recruitment/payment-gateway-challenge-go

//	@host		localhost:8090
//	@BasePath	/

//	@schemes	http
func main() {
	fmt.Printf("version %s, commit %s, built at %s\n", version, commit, date)
	docs.SwaggerInfo.Version = version

	err := run()
	if err != nil {
		fmt.Printf("fatal API error: %v\n", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Printf("sigterm/interrupt signal\n")
		cancel()
	}()

	defer func() {
		// recover after panic
		if x := recover(); x != nil {
			fmt.Printf("run time panic:\n%v\n", x)
			panic(x)
		}
	}()

	api := api.New()
	if err := api.Run(ctx, ":8090"); err != nil {
		return err
	}

	return nil
}
