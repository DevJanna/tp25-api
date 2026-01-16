package interpolation

import (
	"math"
	"sort"
)

// Point represents a data point for interpolation
type Point struct {
	X float64
	Y float64
}

// Curve represents an interpolation curve with data points
type Curve struct {
	Points []Point
}

// NewCurve creates a new interpolation curve from points
func NewCurve(points []Point) *Curve {
	// Sort points by X value
	sort.Slice(points, func(i, j int) bool {
		return points[i].X < points[j].X
	})
	return &Curve{Points: points}
}

// Interpolate performs linear interpolation to find Y value for given X
func (c *Curve) Interpolate(x float64) float64 {
	if len(c.Points) == 0 {
		return 0
	}

	if len(c.Points) == 1 {
		return c.Points[0].Y
	}

	// If x is below the first point, return first Y
	if x <= c.Points[0].X {
		return c.Points[0].Y
	}

	// If x is above the last point, return last Y
	if x >= c.Points[len(c.Points)-1].X {
		return c.Points[len(c.Points)-1].Y
	}

	// Find the two points to interpolate between
	for i := 0; i < len(c.Points)-1; i++ {
		x1, y1 := c.Points[i].X, c.Points[i].Y
		x2, y2 := c.Points[i+1].X, c.Points[i+1].Y

		if x >= x1 && x <= x2 {
			// Linear interpolation formula: y = y1 + (x - x1) * (y2 - y1) / (x2 - x1)
			return y1 + (x-x1)*(y2-y1)/(x2-x1)
		}
	}

	return 0
}

// HydraulicCalculator provides hydraulic engineering calculations
type HydraulicCalculator struct {
	// Water level to volume curve (WAU -> Volume in 10^6 m³)
	VolumeCurve *Curve
	// Water level and discharge to flow rate curve
	FlowCurve *Curve
	// Overflow parameters
	OverflowM float64 // Coefficient m
	OverflowB float64 // Width B in meters
	OverflowG float64 // Gravity constant (9.81 m/s²)
}

// NewHydraulicCalculator creates a new hydraulic calculator with default curves
func NewHydraulicCalculator() *HydraulicCalculator {
	// Default volume curve (WAU -> Volume)
	// These are example points, should be configured per reservoir
	volumePoints := []Point{
		{X: 0, Y: 0},
		{X: 1, Y: 0.5},
		{X: 2, Y: 1.2},
		{X: 3, Y: 2.1},
		{X: 4, Y: 3.2},
		{X: 5, Y: 4.5},
	}

	return &HydraulicCalculator{
		VolumeCurve: NewCurve(volumePoints),
		OverflowM:   0.49, // Typical overflow coefficient
		OverflowB:   10.0, // 10 meters width (example)
		OverflowG:   9.81, // Gravity constant
	}
}

// CalculateWaterIndex calculates water volume (V) from water level (WAU)
// Returns volume in 10^6 m³
func (h *HydraulicCalculator) CalculateWaterIndex(wau float64) float64 {
	return h.VolumeCurve.Interpolate(wau)
}

// CalculateWaterFlow calculates water flow rate (Q) from water level (WAU) and discharge (DR)
// WAU: Water level
// DR: Discharge valve opening or similar parameter
// Returns flow rate in m³/s
func (h *HydraulicCalculator) CalculateWaterFlow(wau, dr float64) float64 {
	if h.FlowCurve != nil {
		// Use curve if available
		// Combine WAU and DR for lookup (this depends on specific curve structure)
		return h.FlowCurve.Interpolate(wau)
	}

	// Simplified calculation based on water level
	// Q = C * A * sqrt(2 * g * h)
	// Where C is discharge coefficient, A is area, g is gravity, h is head
	// This is a simplified model
	if wau <= 0 {
		return 0
	}

	// Simple linear model: Q increases with water level and discharge
	return dr * math.Sqrt(2*h.OverflowG*wau)
}

// CalculateWaterOverFlow calculates overflow discharge (Q_of) from water level (WAU)
// Formula: Q_of = m × B × √(2g) × WAU^1.5
// m: overflow coefficient
// B: width of overflow weir (m)
// g: gravity constant (9.81 m/s²)
// WAU: water level above overflow crest (m)
// Returns overflow discharge in m³/s
func (h *HydraulicCalculator) CalculateWaterOverFlow(wau float64) float64 {
	if wau <= 0 {
		return 0
	}

	// Q_of = m × B × √(2g) × WAU^1.5
	sqrt2g := math.Sqrt(2 * h.OverflowG)
	wauPower := math.Pow(wau, 1.5)

	return h.OverflowM * h.OverflowB * sqrt2g * wauPower
}

// SetVolumeCurve sets a custom volume curve
func (h *HydraulicCalculator) SetVolumeCurve(points []Point) {
	h.VolumeCurve = NewCurve(points)
}

// SetFlowCurve sets a custom flow curve
func (h *HydraulicCalculator) SetFlowCurve(points []Point) {
	h.FlowCurve = NewCurve(points)
}

// SetOverflowParams sets overflow calculation parameters
func (h *HydraulicCalculator) SetOverflowParams(m, b float64) {
	h.OverflowM = m
	h.OverflowB = b
}
