package linkclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Printer struct {
	// Printer state e.g. "Ready"/"Idle"/"Error"
	State PrinterState
	// Loaded material as reported by the printer.
	// e.g. "PETG", "ABS".
	Material string
	// Print speed in percent 0-100.
	PrintSpeed int
	// Number of hot ends.
	ToolCount int
	// Bed temperature.
	BedTemperature Temperature
	// Individual tool temperatures.
	ToolTemperatures map[string]Temperature

	// "raw" api response data.
	Response PrinterResponse
}

type Temperature struct {
	Actual, Target float64
}

func PrinterFromPrinterResponse(res PrinterResponse) Printer {
	fmt.Println(res)
	p := Printer{
		State:      printerStateFromStateResponse(res.State),
		Material:   res.Telemetry.Material,
		PrintSpeed: res.Telemetry.PrintSpeed,
		Response:   res,
	}

	p.ToolTemperatures = map[string]Temperature{}
	for sensor, reading := range res.Temperature {
		if strings.HasPrefix(sensor, "tool") {
			p.ToolCount++
			p.ToolTemperatures[sensor] = Temperature{
				Actual: reading.Actual,
				Target: reading.Target,
			}
			continue
		}
		if sensor == "bed" {
			p.BedTemperature = Temperature{
				Actual: reading.Actual,
				Target: reading.Target,
			}
		}
	}
	return p
}

type PrinterState string

const (
	PrinterStateUnknown   PrinterState = "Unknown"
	PrinterStateIdle      PrinterState = "Idle"
	PrinterStateReady     PrinterState = "Ready"
	PrinterStateBusy      PrinterState = "Busy"
	PrinterStatePaused    PrinterState = "Paused"
	PrinterStatePrinting  PrinterState = "Printing"
	PrinterStateFinished  PrinterState = "Finished"
	PrinterStateStopped   PrinterState = "Stopped"
	PrinterStateError     PrinterState = "Error"
	PrinterStateAttention PrinterState = "Attention"
)

type PrinterResponse struct {
	State       PrinterStateResponse                  `json:"state"`
	Telemetry   PrinterTelemetryResponse              `json:"telemetry"`
	Temperature map[string]PrinterTemperatureResponse `json:"temperature"`
}

type PrinterTelemetryResponse struct {
	TempBed    float64 `json:"temp-bed"`
	TempNozzle float64 `json:"temp-nozzle"`
	PrintSpeed int     `json:"print-speed"`
	ZHeight    float64 `json:"z-height"`
	Material   string  `json:"material"`
}

type PrinterTemperatureResponse struct {
	Actual  float64 `json:"actual"`
	Target  float64 `json:"target"`
	Display float64 `json:"display"`
	Offset  float64 `json:"offset"`
}

type PrinterStateResponse struct {
	Text  string                    `json:"text"`
	Flags PrinterStateFlagsResponse `json:"flags"`
}

type PrinterStateFlagsResponse struct {
	Operational   bool   `json:"operational"`
	Paused        bool   `json:"paused"`
	Printing      bool   `json:"printing"`
	Cancelling    bool   `json:"cancelling"`
	Pausing       bool   `json:"pausing"`
	SDReady       bool   `json:"sdReady"`
	Error         bool   `json:"error"`
	ClosedOnError bool   `json:"closedOnError"`
	Ready         bool   `json:"ready"`
	Busy          bool   `json:"busy"`
	Finished      bool   `json:"finished"`
	LinkState     string `json:"link_state"`
}

func (c *Client) GetPrinter(
	ctx context.Context,
) (res Printer, err error) {
	urlParams := url.Values{}
	response := PrinterResponse{}
	err = c.do(
		ctx, http.MethodGet, "/printer", urlParams, nil, &response,
	)
	if err != nil {
		return res, err
	}

	return PrinterFromPrinterResponse(response), nil
}

// Logic taken from:
// https://github.com/prusa3d/Prusa-Link-Web/blob/15b3af78f2c78fbed358e411e7ff1ccb2d9e4b27/src/state.js
func printerStateFromStateResponse(res PrinterStateResponse) PrinterState {
	// First try to get state from .flags.link_state, if provided.
	switch res.Flags.LinkState {
	case "IDLE":
		return PrinterStateIdle
	case "READY":
		return PrinterStateReady
	case "BUSY":
		return PrinterStateBusy
	case "PRINTING":
		return PrinterStatePrinting
	case "PAUSED":
		return PrinterStatePaused
	case "FINISHED":
		return PrinterStateFinished
	case "STOPPED":
		return PrinterStateStopped
	case "ERROR":
		return PrinterStateError
	case "ATTENTION":
		return PrinterStateAttention
	case "":
		// link_state missing, fallback to discover state from flags.
	default:
		return PrinterStateUnknown
	}

	// Discover state from flags.
	if res.Flags.Error {
		return PrinterStateError
	}
	if strings.ToUpper(res.Text) == "BUSY" {
		return PrinterStateBusy
	}
	if res.Flags.Finished {
		return PrinterStateFinished
	}
	if res.Flags.Pausing || res.Flags.Paused {
		return PrinterStatePaused
	}
	if res.Flags.Ready && res.Flags.Operational {
		return PrinterStateReady
	}
	return PrinterStateIdle
}
