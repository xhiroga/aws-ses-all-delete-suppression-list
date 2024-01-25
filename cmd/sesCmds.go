/*
Copyright © 2024 @xhiroga
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/spf13/cobra"
)

func getAllDestinations(svc *sesv2.Client) ([]types.SuppressedDestinationSummary, error) {
	var destinations []types.SuppressedDestinationSummary
	var nextToken *string
	for {
		input := &sesv2.ListSuppressedDestinationsInput{
			NextToken: nextToken,
		}
		resp, err := svc.ListSuppressedDestinations(context.TODO(), input)

		if err != nil {
			return nil, err
		}

		destinations = append(destinations, resp.SuppressedDestinationSummaries...)

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}
	return destinations, nil
}

func showSummary(destinations []types.SuppressedDestinationSummary) {
	reasonCount := make(map[string]int)
	for _, destination := range destinations {
		reasonCount[string(destination.Reason)]++
	}
	fmt.Println("Reason, Count")
	for reason, count := range reasonCount {
		fmt.Printf("%s, %d\n", reason, count)
	}
	totalDestinations := len(destinations)
	fmt.Printf("TOTAL, %d\n", totalDestinations)
}

func listAll(svc *sesv2.Client) {
	destinations, err := getAllDestinations(svc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Email Address,Reason,Last Update Time")
	for _, destination := range destinations {
		fmt.Printf("%s,%s,%v\n", *destination.EmailAddress, destination.Reason, *destination.LastUpdateTime)
	}
}

func summary(svc *sesv2.Client) {
	destinations, err := getAllDestinations(svc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	showSummary(destinations)
}

func deleteAll(svc *sesv2.Client) {
	destinations, err := getAllDestinations(svc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("You are about to delete the following suppressed email destinations:\n")
	showSummary(destinations)

	fmt.Printf("Do you really want to delete all suppressed email destinations? This action cannot be undone.\n")
	fmt.Printf("If you want to continue, type 'delete' to proceed: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "delete" {
		fmt.Println("Operation cancelled.")
		os.Exit(1)
	}

	fmt.Println("Deleting suppressed email destinations:")
	totalDestinations := len(destinations)
	for i, destination := range destinations {
		input := &sesv2.DeleteSuppressedDestinationInput{
			EmailAddress: destination.EmailAddress,
		}
		for j := 0; j < 3; j++ {
			_, err := svc.DeleteSuppressedDestination(context.TODO(), input)
			if err != nil {
				if strings.Contains(err.Error(), "TooManyRequestsException") {
					time.Sleep(time.Second * time.Duration(j+1))
					continue
				}
				fmt.Println(err)
				os.Exit(1)
			}
			break
		}
		fmt.Printf("\rProgress: %d/%d\n", i+1, totalDestinations)
	}
	fmt.Println("\nDeletion complete.")
}
func getSvc() *sesv2.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	if cfg.Region == "" {
		log.Fatalf("AWS region is not set")
	}
	return sesv2.NewFromConfig(cfg)
}

var listAllCmd = &cobra.Command{
	Use:   "listAll",
	Short: "Lists all suppressed email destinations",
	Long: `This command retrieves and lists all the suppressed email destinations
from your AWS SES account.`,
	Run: func(cmd *cobra.Command, args []string) {
		svc := getSvc()
		listAll(svc)
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarizes the suppressed email destinations by reason",
	Long: `This command summarizes all the suppressed email destinations by reason
from your AWS SES account.`,
	Run: func(cmd *cobra.Command, args []string) {
		svc := getSvc()
		summary(svc)
	},
}

var deleteAllCmd = &cobra.Command{
	Use:   "deleteAll",
	Short: "Deletes all suppressed email destinations",
	Long: `This command deletes all the suppressed email destinations
from your AWS SES account.`,
	Run: func(cmd *cobra.Command, args []string) {
		svc := getSvc()
		deleteAll(svc)
	},
}

func init() {
	rootCmd.AddCommand(listAllCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(deleteAllCmd)
}
