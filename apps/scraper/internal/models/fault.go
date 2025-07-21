package models

import "fmt"

// Fault represents the top-level fault data structure from EMSC
type Fault struct {
	Type     string         `json:"type"`
	Features []FaultFeature `json:"features"`
}

// FaultFeature represents a single fault feature
type FaultFeature struct {
	Type       string          `json:"type"`
	Properties FaultProperties `json:"properties"`
	Geometry   FaultGeometry   `json:"geometry"`
	ID         string          `json:"id,omitempty"`
}

// FaultProperties contains the properties of a fault
type FaultProperties struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	SlipRate     *float64 `json:"slip_rate,omitempty"`
	SlipType     string   `json:"slip_type,omitempty"`
	Dip          *float64 `json:"dip,omitempty"`
	Rake         *float64 `json:"rake,omitempty"`
	Length       *float64 `json:"length,omitempty"`
	Width        *float64 `json:"width,omitempty"`
	MaxMagnitude *float64 `json:"max_magnitude,omitempty"`
	Description  string   `json:"description,omitempty"`
	Source       string   `json:"source,omitempty"`
}

// FaultGeometry represents the geographical geometry of a fault
type FaultGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// GetSlipRate returns the slip rate as a string with units
func (f *FaultProperties) GetSlipRate() string {
	if f.SlipRate != nil {
		return fmt.Sprintf("%.2f mm/year", *f.SlipRate)
	}
	return "Unknown"
}

// GetMaxMagnitude returns the maximum magnitude as a string
func (f *FaultProperties) GetMaxMagnitude() string {
	if f.MaxMagnitude != nil {
		return fmt.Sprintf("%.1f", *f.MaxMagnitude)
	}
	return "Unknown"
}

// GetLength returns the fault length as a string with units
func (f *FaultProperties) GetLength() string {
	if f.Length != nil {
		return fmt.Sprintf("%.1f km", *f.Length)
	}
	return "Unknown"
}

// GetWidth returns the fault width as a string with units
func (f *FaultProperties) GetWidth() string {
	if f.Width != nil {
		return fmt.Sprintf("%.1f km", *f.Width)
	}
	return "Unknown"
}
