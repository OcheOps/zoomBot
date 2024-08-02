package models

import "time"

type Meeting struct {
    ID        string    `json:"id"`
    Link      string    `json:"link"`
    Password  string    `json:"password"`
    StartTime time.Time `json:"start_time"`
    Status    string    `json:"status"`
}