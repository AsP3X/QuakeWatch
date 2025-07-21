package models

import (
	"fmt"
	"time"
)

// USGSResponse represents the top-level response from USGS API
type USGSResponse struct {
	Type     string       `json:"type"`
	Metadata Metadata     `json:"metadata"`
	Features []Earthquake `json:"features"`
}

// Metadata contains information about the API response
type Metadata struct {
	Generated int64  `json:"generated"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	API       string `json:"api"`
	Count     int    `json:"count"`
}

// Earthquake represents a single earthquake event
type Earthquake struct {
	Type       string               `json:"type"`
	Properties EarthquakeProperties `json:"properties"`
	Geometry   Geometry             `json:"geometry"`
	ID         string               `json:"id"`
}

// EarthquakeProperties contains the properties of an earthquake
type EarthquakeProperties struct {
	Mag     float64  `json:"mag"`
	Place   string   `json:"place"`
	Time    int64    `json:"time"`
	Updated int64    `json:"updated"`
	URL     string   `json:"url"`
	Detail  string   `json:"detail"`
	Felt    *int     `json:"felt,omitempty"`
	CDI     *float64 `json:"cdi,omitempty"`
	MMI     *float64 `json:"mmi,omitempty"`
	Alert   string   `json:"alert,omitempty"`
	Status  string   `json:"status"`
	Tsunami int      `json:"tsunami"`
	Sig     int      `json:"sig"`
	Net     string   `json:"net"`
	Code    string   `json:"code"`
	IDs     string   `json:"ids"`
	Sources string   `json:"sources"`
	Types   string   `json:"types"`
	Nst     *int     `json:"nst,omitempty"`
	Dmin    *float64 `json:"dmin,omitempty"`
	RMS     *float64 `json:"rms,omitempty"`
	Gap     *float64 `json:"gap,omitempty"`
	MagType string   `json:"magType,omitempty"`
	Type    string   `json:"type"`
	Title   string   `json:"title"`
}

// Geometry represents the geographical location of an earthquake
type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// GetTime returns the earthquake time as a time.Time
func (e *EarthquakeProperties) GetTime() time.Time {
	return time.Unix(e.Time/1000, 0)
}

// GetUpdated returns the earthquake updated time as a time.Time
func (e *EarthquakeProperties) GetUpdated() time.Time {
	return time.Unix(e.Updated/1000, 0)
}

// IsSignificant returns true if the earthquake magnitude is 4.5 or greater
func (e *EarthquakeProperties) IsSignificant() bool {
	return e.Mag >= 4.5
}

// GetMagnitude returns the magnitude as a string with type
func (e *EarthquakeProperties) GetMagnitude() string {
	if e.MagType != "" {
		return fmt.Sprintf("%.1f %s", e.Mag, e.MagType)
	}
	return fmt.Sprintf("%.1f", e.Mag)
}
