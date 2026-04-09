package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"solenya-sync/internal/kdrive"
	"solenya-sync/internal/markdown"
	"solenya-sync/internal/metadata"
	"solenya-sync/internal/utils"
	"time"
)

func main() {
	// Flags for kDrive configuration
	driveID := flag.String("drive-id", os.Getenv("KDRIVE_DRIVE_ID"), "kDrive Drive ID")
	folderID := flag.String("folder-id", os.Getenv("KDRIVE_FOLDER_ID"), "kDrive Root Folder ID")
	apiToken := flag.String("token", os.Getenv("KDRIVE_API_TOKEN"), "kDrive API Token")
	outDir := flag.String("out", "src/content/rolls/synced", "Output directory for Markdown files")
	
	flag.Parse()

	if *driveID == "" || *folderID == "" || *apiToken == "" {
		log.Fatalf("Missing required configuration. Provide flags or set environment variables: KDRIVE_DRIVE_ID, KDRIVE_FOLDER_ID, KDRIVE_API_TOKEN")
	}

	fmt.Printf("🚀 Initializing Solenya Sync (Go Edition)\n")
	fmt.Printf("📂 Target: %s\n", *outDir)

	client := kdrive.NewClient(*driveID, *apiToken)

	// Fetch rolls (folders in the root)
	items, err := client.GetFiles(*folderID)
	if err != nil {
		log.Fatalf("❌ Failed to fetch rolls: %v", err)
	}

	for _, item := range items {
		if item.Type != "dir" {
			continue
		}

		fmt.Printf("\n📸 Processing Roll: %s\n", item.Name)
		
		// In a real implementation, we would fetch files in this folder,
		// download them, extract EXIF/Palettes, and then write the Roll.
		// For now, we'll demonstrate the structure using the implemented packages.
		
		rollData := &markdown.RollData{
			Title: item.Name,
			Date:  time.Unix(item.CreatedAt, 0).Format("2006-01-02"),
			Tags:  []string{"synced"}, // Placeholder tags
		}

		// Example image data (this would be populated by the kdrive client + processor)
		rollData.Images = []markdown.ImageData{
			{
				URL: "https://example.com/placeholder.jpg",
				Exif: &metadata.ExifData{
					Body: "Solenya-MK1",
				},
			},
		}

		slug := utils.Slugify(item.Name)
		if err := markdown.WriteRoll(*outDir, slug, rollData); err != nil {
			fmt.Printf("   ❌ Error writing roll %s: %v\n", item.Name, err)
		} else {
			fmt.Printf("   ✅ Generated %s.md\n", slug)
		}
	}

	fmt.Println("\n🎉 Sync complete. Go home, Jerry.")
}
