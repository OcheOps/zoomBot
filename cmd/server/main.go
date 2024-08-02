package main

import (
    "log"
    "os"
    "path/filepath"

    "zoombot/internal/api"
    "zoombot/internal/bot"
    "zoombot/internal/database"
)

func main() {
    db, err := database.InitDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("Failed to get user home directory: %v", err)
    }
    outputDir := filepath.Join(homeDir, "Downloads", "ZoomRecordings")
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        log.Fatalf("Failed to create output directory: %v", err)
    }

    zoomBot, err := bot.NewBot(db, outputDir)
    if err != nil {
        log.Fatalf("Failed to create bot: %v", err)
    }
    defer zoomBot.Close()

    server := api.NewServer(zoomBot)
    log.Printf("Server starting on :8080")
    log.Fatal(server.Run(":8080"))
}