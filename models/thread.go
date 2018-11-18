package models

import (
	"time"
)

//easyjson:json
type Thread struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created,omitempty"`
	Forum   string    `json:"forum,omitempty"`
	ID      int32     `json:"id,omitempty"`
	Message string    `json:"message"`
	Slug    string    `json:"slug,omitempty"`
	Title   string    `json:"title"`
	Votes   int32     `json:"votes,omitempty"`
}

//easyjson:json
type ThreadUpdate struct {
	Message string `json:"message,omitempty"`
	Title   string `json:"title,omitempty"`
}

//easyjson:json
type Threads []*Thread
