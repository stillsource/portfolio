package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"solenya-sync/internal/kdrive"
)

func main() {
	// Flags for kDrive configuration
	driveID := flag.String("drive-id", os.Getenv("KDRIVE_DRIVE_ID"), "kDrive Drive ID")
	folderID := flag.String("folder-id", os.Getenv("KDRIVE_FOLDER_ID"), "kDrive Root Folder ID")
	apiToken := flag.String("token", os.Getenv("KDRIVE_API_TOKEN"), "kDrive API Token")
	
	flag.Parse()

	// Validate inputs
	if *driveID == "" || *folderID == "" || *apiToken == "" {
		log.Fatalf("Missing required configuration. Provide flags or set environment variables: KDRIVE_DRIVE_ID, KDRIVE_FOLDER_ID, KDRIVE_API_TOKEN")
	}

	fmt.Printf("🚀 Initializing sync for Drive ID: %s, Folder ID: %s\n", *driveID, *folderID)

	// Create client
	client := kdrive.NewClient(*driveID, *apiToken)

	// Test fetch
	files, err := client.GetFiles(*folderID)
	if err != nil {
		log.Fatalf("❌ Failed to fetch files: %v", err)
	}

	fmt.Printf("📂 Found %d items in root folder.\n", len(files))
	for _, f := range files {
		fmt.Printf("   - [%s] %s (ID: %s)\n", f.Type, f.Name, f.ID)
	}

	fmt.Println("\n🎉 Initialization complete. Ready for full sync.")
}
