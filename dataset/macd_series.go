package dataset

import (
	"fmt"

	"github.com/unidoc/unichart/dataset/sequence"
	"github.com/unidoc/unichart/render"
)

const (
	// defaultMACDPeriodPrimary is the long window.
	defaultMACDPeriodPrimary = 26

	// defaultMACDPeriodSecondary is the short window.
	defaultMACDPeriodSecondary = 12

	// defaultMACDSignalPeriod is the signal period to compute for the MACD.
	defaultMACDSignalPeriod = 9
)

// MACDSeries computes the difference between the MACD line and the MACD Signal line.
// It is used in technical analysis and gives a lagging indicator of momentum.
type MACDSeries struct {
	Name        string
	Style       render.Style
	YAxis       YAxisType
	InnerSeries ValuesProvider

	PrimaryPeriod   int
	SecondaryPeriod int
	SignalPeriod    int

	signal *MACDSignalSeries
	macdl  *MACDLineSeries
}

// Validate validates the series.
func (macd MACDSeries) Validate() error {
	var err error
	if macd.signal != nil {
		err = macd.signal.Validate()
	}
	if err != nil {
		return err
	}
	if macd.macdl != nil {
		err = macd.macdl.Validate()
	}
	if err != nil {
		return err
	}
	return nil
}

// GetPeriods returns the primary and secondary periods.
func (macd MACDSeries) GetPeriods() (w1, w2, sig int) {
	if macd.PrimaryPeriod == 0 {
		w1 = defaultMACDPeriodPrimary
	} else {
		w1 = macd.PrimaryPeriod
	}
	if macd.SecondaryPeriod == 0 {
		w2 = defaultMACDPeriodSecondary
	} else {
		w2 = macd.SecondaryPeriod
	}
	if macd.SignalPeriod == 0 {
		sig = defaultMACDSignalPeriod
	} else {
		sig = macd.SignalPeriod
	}
	return
}

// GetName returns the name of the time series.
func (macd MACDSeries) GetName() string {
	return macd.Name
}

// GetStyle returns the line style.
func (macd MACDSeries) GetStyle() render.Style {
	return macd.Style
}

// GetYAxis returns which YAxis the series draws on.
func (macd MACDSeries) GetYAxis() YAxisType {
	return macd.YAxis
}

// Len returns the number of elements in the series.
func (macd MACDSeries) Len() int {
	if macd.InnerSeries == nil {
		return 0
	}

	return macd.InnerSeries.Len()
}

// GetValues gets a value at a given index. For MACD it is the signal value.
func (macd *MACDSeries) GetValues(index int) (x float64, y float64) {
	if macd.InnerSeries == nil {
		return
	}

	if macd.signal == nil || macd.macdl == nil {
		macd.ensureChildSeries()
	}

	_, lv := macd.macdl.GetValues(index)
	_, sv := macd.signal.GetValues(index)

	x, _ = macd.InnerSeries.GetValues(index)
	y = lv - sv

	return
}

func (macd *MACDSeries) ensureChildSeries() {
	w1, w2, sig := macd.GetPeriods()

	macd.signal = &MACDSignalSeries{
		InnerSeries:     macd.InnerSeries,
		PrimaryPeriod:   w1,
		SecondaryPeriod: w2,
		SignalPeriod:    sig,
	}

	macd.macdl = &MACDLineSeries{
		InnerSeries:     macd.InnerSeries,
		PrimaryPeriod:   w1,
		SecondaryPeriod: w2,
	}
}

// MACDSignalSeries computes the EMA of the MACDLineSeries.
type MACDSignalSeries struct {
	Name        string
	Style       render.Style
	YAxis       YAxisType
	InnerSeries ValuesProvider

	PrimaryPeriod   int
	SecondaryPeriod int
	SignalPeriod    int

	signal *EMASeries
}

// Validate validates the series.
func (macds MACDSignalSeries) Validate() error {
	if macds.signal != nil {
		return macds.signal.Validate()
	}
	return nil
}

// GetPeriods returns the primary and secondary periods.
func (macds MACDSignalSeries) GetPeriods() (w1, w2, sig int) {
	if macds.PrimaryPeriod == 0 {
		w1 = defaultMACDPeriodPrimary
	} else {
		w1 = macds.PrimaryPeriod
	}
	if macds.SecondaryPeriod == 0 {
		w2 = defaultMACDPeriodSecondary
	} else {
		w2 = macds.SecondaryPeriod
	}
	if macds.SignalPeriod == 0 {
		sig = defaultMACDSignalPeriod
	} else {
		sig = macds.SignalPeriod
	}
	return
}

// GetName returns the name of the time series.
func (macds MACDSignalSeries) GetName() string {
	return macds.Name
}

// GetStyle returns the line style.
func (macds MACDSignalSeries) GetStyle() render.Style {
	return macds.Style
}

// GetYAxis returns which YAxis the series draws on.
func (macds MACDSignalSeries) GetYAxis() YAxisType {
	return macds.YAxis
}

// Len returns the number of elements in the series.
func (macds *MACDSignalSeries) Len() int {
	if macds.InnerSeries == nil {
		return 0
	}

	return macds.InnerSeries.Len()
}

// GetValues gets a value at a given index. For MACD it is the signal value.
func (macds *MACDSignalSeries) GetValues(index int) (x float64, y float64) {
	if macds.InnerSeries == nil {
		return
	}

	if macds.signal == nil {
		macds.ensureSignal()
	}
	x, _ = macds.InnerSeries.GetValues(index)
	_, y = macds.signal.GetValues(index)
	return
}

func (macds *MACDSignalSeries) ensureSignal() {
	w1, w2, sig := macds.GetPeriods()

	macds.signal = &EMASeries{
		InnerSeries: &MACDLineSeries{
			InnerSeries:     macds.InnerSeries,
			PrimaryPeriod:   w1,
			SecondaryPeriod: w2,
		},
		Period: sig,
	}
}

// Render renders the series.
func (macds *MACDSignalSeries) Render(r render.Renderer, canvasBox render.Box, xrange, yrange sequence.Range, defaults render.Style) {
	style := macds.Style.InheritFrom(defaults)
	drawLineSeries(r, canvasBox, xrange, yrange, style, macds)
}

// MACDLineSeries is a series that computes the inner ema1-ema2 value as a series.
type MACDLineSeries struct {
	Name        string
	Style       render.Style
	YAxis       YAxisType
	InnerSeries ValuesProvider

	PrimaryPeriod   int
	SecondaryPeriod int

	ema1 *EMASeries
	ema2 *EMASeries

	Sigma float64
}

// Validate validates the series.
func (macdl MACDLineSeries) Validate() error {
	var err error
	if macdl.ema1 != nil {
		err = macdl.ema1.Validate()
	}
	if err != nil {
		return err
	}
	if macdl.ema2 != nil {
		err = macdl.ema2.Validate()
	}
	if err != nil {
		return err
	}
	if macdl.InnerSeries == nil {
		return fmt.Errorf("MACDLineSeries: must provide an inner series")
	}
	return nil
}

// GetName returns the name of the time series.
func (macdl MACDLineSeries) GetName() string {
	return macdl.Name
}

// GetStyle returns the line style.
func (macdl MACDLineSeries) GetStyle() render.Style {
	return macdl.Style
}

// GetYAxis returns which YAxis the series draws on.
func (macdl MACDLineSeries) GetYAxis() YAxisType {
	return macdl.YAxis
}

// GetPeriods returns the primary and secondary periods.
func (macdl MACDLineSeries) GetPeriods() (w1, w2 int) {
	if macdl.PrimaryPeriod == 0 {
		w1 = defaultMACDPeriodPrimary
	} else {
		w1 = macdl.PrimaryPeriod
	}
	if macdl.SecondaryPeriod == 0 {
		w2 = defaultMACDPeriodSecondary
	} else {
		w2 = macdl.SecondaryPeriod
	}
	return
}

// Len returns the number of elements in the series.
func (macdl *MACDLineSeries) Len() int {
	if macdl.InnerSeries == nil {
		return 0
	}

	return macdl.InnerSeries.Len()
}

// GetValues gets a value at a given index. For MACD it is the signal value.
func (macdl *MACDLineSeries) GetValues(index int) (x float64, y float64) {
	if macdl.InnerSeries == nil {
		return
	}
	if macdl.ema1 == nil && macdl.ema2 == nil {
		macdl.ensureEMASeries()
	}

	x, _ = macdl.InnerSeries.GetValues(index)

	_, emav1 := macdl.ema1.GetValues(index)
	_, emav2 := macdl.ema2.GetValues(index)

	y = emav2 - emav1
	return
}

func (macdl *MACDLineSeries) ensureEMASeries() {
	w1, w2 := macdl.GetPeriods()

	macdl.ema1 = &EMASeries{
		InnerSeries: macdl.InnerSeries,
		Period:      w1,
	}
	macdl.ema2 = &EMASeries{
		InnerSeries: macdl.InnerSeries,
		Period:      w2,
	}
}

// Render renders the series.
func (macdl *MACDLineSeries) Render(r render.Renderer, canvasBox render.Box, xrange, yrange sequence.Range, defaults render.Style) {
	style := macdl.Style.InheritFrom(defaults)
	drawLineSeries(r, canvasBox, xrange, yrange, style, macdl)
}
