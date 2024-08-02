package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

type Scraper struct {
    ctx            context.Context
    cancel         context.CancelFunc
    recordingCmd   *exec.Cmd
    outputDir      string
}

func NewScraper(outputDir string) (*Scraper, error) {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", false),
        chromedp.Flag("use-fake-ui-for-media-stream", true),
        chromedp.Flag("use-fake-device-for-media-stream", true),
    )
    allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
    ctx, cancel := chromedp.NewContext(allocCtx)

    return &Scraper{
        ctx:       ctx,
        cancel:    cancel,
        outputDir: outputDir,
    }, nil
}

func (s *Scraper) Close() {
    s.StopRecording()
    s.cancel()
}

func (s *Scraper) JoinMeeting(link, password string) error {
    var err error
    err = chromedp.Run(s.ctx,
        chromedp.Navigate(link),
        chromedp.WaitVisible(`#join-form`, chromedp.ByID),
        chromedp.SendKeys(`input[type="text"]`, password, chromedp.ByQuery),
        chromedp.Click(`#joinBtn`, chromedp.ByID),
        chromedp.WaitVisible(`#wc-footer`, chromedp.ByID),
    )
    if err != nil {
        return fmt.Errorf("failed to join meeting: %w", err)
    }
    log.Println("Successfully joined meeting")
    return nil
}

func (s *Scraper) StartRecording(meetingID string) error {
    outputFile := filepath.Join(s.outputDir, fmt.Sprintf("meeting_%s_%s.mp4", meetingID, time.Now().Format("20060102150405")))
    cmd := exec.Command("ffmpeg",
        "-f", "gdigrab",
        "-framerate", "30",
        "-i", "desktop",
        "-f", "dshow",
        "-i", "audio=virtual-audio-capturer",
        "-c:v", "libx264",
        "-preset", "ultrafast",
        "-c:a", "aac",
        outputFile)

    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("failed to start recording: %w", err)
    }

    s.recordingCmd = cmd
    log.Printf("Started recording to %s", outputFile)
    return nil
}

func (s *Scraper) StopRecording() error {
    if s.recordingCmd == nil || s.recordingCmd.Process == nil {
        return nil
    }

    err := s.recordingCmd.Process.Signal(os.Interrupt)
    if err != nil {
        return fmt.Errorf("failed to stop recording: %w", err)
    }

    err = s.recordingCmd.Wait()
    if err != nil {
        log.Printf("Recording command exited with error: %v", err)
    }

    s.recordingCmd = nil
    log.Println("Stopped recording")
    return nil
}

func (s *Scraper) LeaveMeeting() error {
    err := chromedp.Run(s.ctx,
        chromedp.Click(`#wc-footer-left > div.footer-button-base__button-group > button:nth-child(1)`, chromedp.ByQuery),
        chromedp.Click(`#wc-container-right > div > div.leave-meeting-options__inner-frame > div.leave-meeting-options__buttons > button.zm-btn.zm-btn-legacy.zm-btn--primary.zm-btn__outline--blue`, chromedp.ByQuery),
    )
    if err != nil {
        return fmt.Errorf("failed to leave meeting: %w", err)
    }
    log.Println("Successfully left meeting")
    return nil
}