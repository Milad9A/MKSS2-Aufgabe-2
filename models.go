package main

import "time"

// Position represents the robot's coordinates
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Action represents an activity performed by a robot
type Action struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

// Robot represents a robot in the system
type Robot struct {
	ID        string   `json:"id"`
	Position  Position `json:"position"`
	Direction string   `json:"direction"` // "north", "east", "south", "west"
	Energy    int      `json:"energy"`
	Inventory []string `json:"inventory"`
	Actions   []Action `json:"actions"`
}

// MoveRequest is the payload for the move endpoint
type MoveRequest struct {
	Direction string `json:"direction"` // "up", "down", "left", "right"
}

// StateUpdateRequest is the payload for the state update endpoint
type StateUpdateRequest struct {
	Energy   *int      `json:"energy,omitempty"`
	Position *Position `json:"position,omitempty"`
}

// Link represents a HATEOAS link
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// PageInfo contains pagination information
type PageInfo struct {
	Number        int  `json:"number"`
	Size          int  `json:"size"`
	TotalElements int  `json:"totalElements"`
	TotalPages    int  `json:"totalPages"`
	HasNext       bool `json:"hasNext"`
	HasPrevious   bool `json:"hasPrevious"`
}

// ActionWithLinks represents an action with HATEOAS links
type ActionWithLinks struct {
	Action
	Links []Link `json:"links"`
}

// PaginatedActions represents a paginated list of actions with navigation links
type PaginatedActions struct {
	Page    PageInfo          `json:"page"`
	Actions []ActionWithLinks `json:"actions"`
	Links   []Link            `json:"links"`
}
