package main

import (
	"fmt"
	"log"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/gophon"
)

func main() {
	namespace := "github.com/hashicorp/terraform-provider-azurerm/internal"
	
	fmt.Printf("Testing ListSupportedTags for namespace: %s\n", namespace)
	
	tags, err := gophon.ListSupportedTags(namespace)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Printf("Found %d tags:\n", len(tags))
	
	// Show first 10 and last 10 tags
	if len(tags) > 20 {
		fmt.Println("First 10 tags:")
		for i := 0; i < 10; i++ {
			fmt.Printf("  %s\n", tags[i])
		}
		fmt.Println("...")
		fmt.Println("Last 10 tags:")
		for i := len(tags) - 10; i < len(tags); i++ {
			fmt.Printf("  %s\n", tags[i])
		}
	} else {
		fmt.Println("All tags:")
		for _, tag := range tags {
			fmt.Printf("  %s\n", tag)
		}
	}
}
