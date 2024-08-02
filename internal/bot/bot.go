package bot

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    "zoombot/internal/models"
    "zoombot/internal/scraper"
)

type Bot struct {
    DB      *sql.DB
    Scraper *scraper.Scraper
}

func NewBot(db *sql.DB, outputDir string) (*Bot, error) {
    scraper, err := scraper.NewScraper(outputDir)
    if err != nil {
        return nil, fmt.Errorf("failed to create scraper: %w", err)
    }

    return &Bot{
        DB:      db,
        Scraper: scraper,
    }, nil
}

func (b *Bot) Close() {
    b.Scraper.Close()
}

func (b *Bot) AddMeeting(meeting *models.Meeting) error {
    _, err := b.DB.Exec("INSERT INTO meetings (link, password, start_time) VALUES (?, ?, ?)",
        meeting.Link, meeting.Password, meeting.StartTime)
    if err != nil {
        return fmt.Errorf("failed to add meeting: %w", err)
    }
    return nil
}

func (b *Bot) ListMeetings() ([]models.Meeting, error) {
    rows, err := b.DB.Query("SELECT id, link, password, start_time FROM meetings ORDER BY start_time")
    if err != nil {
        return nil, fmt.Errorf("failed to query meetings: %w", err)
    }
    defer rows.Close()

    var meetings []models.Meeting
    for rows.Next() {
        var m models.Meeting
        if err := rows.Scan(&m.ID, &m.Link, &m.Password, &m.StartTime); err != nil {
            return nil, fmt.Errorf("failed to scan meeting row: %w", err)
        }
        meetings = append(meetings, m)
    }
    return meetings, nil
}

func (b *Bot) JoinMeeting(id string) {
    var meeting models.Meeting
    err := b.DB.QueryRow("SELECT id, link, password, start_time FROM meetings WHERE id = ?", id).Scan(&meeting.ID, &meeting.Link, &meeting.Password, &meeting.StartTime)
    if err != nil {
        log.Printf("Error fetching meeting: %v", err)
        return
    }

    if err := b.Scraper.JoinMeeting(meeting.Link, meeting.Password); err != nil {
        log.Printf("Error joining meeting: %v", err)
        return
    }

    if err := b.Scraper.StartRecording(meeting.ID); err != nil {
        log.Printf("Error starting recording: %v", err)
        return
    }

    // Wait for the scheduled duration or a maximum of 2 hours
    duration := time.Until(meeting.StartTime.Add(2 * time.Hour))
    if duration > 2*time.Hour {
        duration = 2 * time.Hour
    }
    time.Sleep(duration)

    if err := b.Scraper.StopRecording(); err != nil {
        log.Printf("Error stopping recording: %v", err)
    }

    if err := b.Scraper.LeaveMeeting(); err != nil {
        log.Printf("Error leaving meeting: %v", err)
    }

    if err := b.UpdateMeetingStatus(meeting.ID, "completed"); err != nil {
        log.Printf("Error updating meeting status: %v", err)
    }
}

func (b *Bot) UpdateMeetingStatus(id string, status string) error {
    _, err := b.DB.Exec("UPDATE meetings SET status = ? WHERE id = ?", status, id)
    if err != nil {
        return fmt.Errorf("failed to update meeting status: %w", err)
    }
    return nil
}