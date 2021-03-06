/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2017
	All Rights Reserved

    Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

// This example is the OpenVG tiger example, which draws a tiger using
// commands from a data file (tiger_data.txt) which should be in the same
// folder as this example code
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

import (
	app "github.com/djthorpe/gopi/app"
	khronos "github.com/djthorpe/gopi/khronos"
)

////////////////////////////////////////////////////////////////////////////////

type Operation struct {
	fill   khronos.VGPaint
	stroke khronos.VGPaint
	path   khronos.VGPath
}

////////////////////////////////////////////////////////////////////////////////

var (
	opcode_r = regexp.MustCompile("'(\\w)'")
	value_r  = regexp.MustCompile("([0-9\\.]*[0-9]+)f?")
)

////////////////////////////////////////////////////////////////////////////////

// Return the opcodes, values and error
func ReadData(filename string) ([]string, []float32, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	// Read opcodes and values
	opcodes := opcode_r.FindAllSubmatch(data, -1)
	if opcodes == nil {
		return nil, nil, errors.New("Invalid data file, no opcodes")
	}
	values := value_r.FindAllSubmatch(data, -1)
	if values == nil {
		return nil, nil, errors.New("Invalid data file, no values")
	}

	opcodes2 := make([]string, len(opcodes))
	values2 := make([]float32, len(values))

	// Convert opcodes to string
	for i, opcode := range opcodes {
		opcodes2[i] = string(opcode[1])
	}

	// Convert values to float32
	for i, value := range values {
		value64, err := strconv.ParseFloat(string(value[1]), 32)
		if err != nil {
			return nil, nil, err
		}
		values2[i] = float32(value64)
	}

	// Success
	return opcodes2, values2, nil
}

////////////////////////////////////////////////////////////////////////////////

func (this *Operation) ParseFillOpcode(vg khronos.VGDriver, code string) error {
	var err error
	switch code {
	case "N":
		this.fill = nil
	case "F":
		this.fill, err = vg.CreatePaint(khronos.VGColorWhite)
		if err != nil {
			return err
		}
		if err := this.fill.SetFillRule(khronos.VG_STYLE_FILL_NONZERO); err != nil {
			return err
		}
	case "E":
		this.fill, err = vg.CreatePaint(khronos.VGColorWhite)
		if err != nil {
			return err
		}
		if err := this.fill.SetFillRule(khronos.VG_STYLE_FILL_EVENODD); err != nil {
			return err
		}
	default:
		return errors.New("Invalid ParseFillOpcode value")
	}
	return nil
}

func (this *Operation) ParseStrokeOpcode(vg khronos.VGDriver, code string) error {
	var err error
	switch code {
	case "N":
		this.stroke = nil
	case "S":
		// TODO
		this.stroke, err = vg.CreatePaint(khronos.VGColorWhite)
		if err != nil {
			return err
		}
	default:
		return errors.New("Invalid ParseStrokeOpcode value")
	}
	return nil
}

func (this *Operation) ParseLineCapOpcode(vg khronos.VGDriver, code string) error {
	if this.stroke == nil {
		return nil
	}
	switch code {
	case "B":
		return this.stroke.SetStrokeStyle(khronos.VG_STYLE_JOIN_NONE, khronos.VG_STYLE_CAP_BUTT)
	case "R":
		return this.stroke.SetStrokeStyle(khronos.VG_STYLE_JOIN_NONE, khronos.VG_STYLE_CAP_ROUND)
	case "S":
		return this.stroke.SetStrokeStyle(khronos.VG_STYLE_JOIN_NONE, khronos.VG_STYLE_CAP_SQUARE)
	default:
		return errors.New("Invalid ParseLineCapOpcode value")
	}
	return nil
}

func (this *Operation) ParseLineJoinOpcode(vg khronos.VGDriver, code string) error {
	if this.stroke == nil {
		return nil
	}
	switch code {
	case "M":
		return this.stroke.SetStrokeStyle(khronos.VG_STYLE_JOIN_MITER, khronos.VG_STYLE_CAP_NONE)
	case "R":
		return this.stroke.SetStrokeStyle(khronos.VG_STYLE_JOIN_ROUND, khronos.VG_STYLE_CAP_NONE)
	case "B":
		return this.stroke.SetStrokeStyle(khronos.VG_STYLE_JOIN_BEVEL, khronos.VG_STYLE_CAP_NONE)
	default:
		return errors.New("Invalid ParseLineJoinOpcode value")
	}
	return nil
}

func (this *Operation) ParseMiterLimit(vg khronos.VGDriver, limit float32) error {
	if this.stroke == nil {
		return nil
	}
	return this.stroke.SetMiterLimit(limit)
}

func (this *Operation) ParseStrokeWidth(vg khronos.VGDriver, width float32) error {
	if this.stroke == nil {
		return nil
	}
	return this.stroke.SetStrokeWidth(width)
}

func (this *Operation) ParseStrokeColor(vg khronos.VGDriver, r, g, b float32) error {
	if this.stroke == nil {
		return nil
	}
	return this.stroke.SetColor(khronos.VGColor{r, g, b, 1.0})
}

func (this *Operation) ParseFillColor(vg khronos.VGDriver, r, g, b float32) error {
	if this.fill == nil {
		return nil
	}
	return this.fill.SetColor(khronos.VGColor{r, g, b, 1.0})
}

func (this *Operation) ParsePathPoint(vg khronos.VGDriver, opcode string, points []float32, i int) (int, error) {
	switch opcode {
	case "M":
		if err := this.path.MoveTo(khronos.VGPoint{points[i], points[i+1]}); err != nil {
			return 0, err
		}
		return 2, nil
	case "L":
		if err := this.path.LineTo(khronos.VGPoint{points[i], points[i+1]}); err != nil {
			return 0, err
		}
		return 2, nil
	case "C":
		if err := this.path.CubicTo(khronos.VGPoint{points[i], points[i+1]}, khronos.VGPoint{points[i+2], points[i+3]}, khronos.VGPoint{points[i+4], points[i+5]}); err != nil {
			return 0, err
		}
		return 6, nil
	case "E":
		if err := this.path.Close(); err != nil {
			return 0, err
		}
		return 0, nil
	default:
		return 0, errors.New("Invalid ParsePathPoint opcode value")
	}
}

////////////////////////////////////////////////////////////////////////////////

func ProcessDataFromFile(app *app.App, filename string) ([]*Operation, error) {

	// Read data from file
	opcodes, values, err := ReadData(filename)
	if err != nil {
		return nil, err
	}

	// Create operations
	operations := make([]*Operation, 0, len(opcodes))
	c := 0
	v := 0
	for c < len(opcodes) && v < len(values) {
		app.Logger.Debug("=> Opcode %v", c)

		op := new(Operation)
		// Fill opcode
		if err := op.ParseFillOpcode(app.OpenVG, opcodes[c]); err != nil {
			return nil, err
		}
		c += 1

		// Stroke opcode
		if err := op.ParseStrokeOpcode(app.OpenVG, opcodes[c]); err != nil {
			return nil, err
		}
		c += 1

		// Line Cap
		if err := op.ParseLineCapOpcode(app.OpenVG, opcodes[c]); err != nil {
			return nil, err
		}
		c += 1

		// Line Join
		if err := op.ParseLineJoinOpcode(app.OpenVG, opcodes[c]); err != nil {
			return nil, err
		}
		c += 1

		// Miter Limit
		if err := op.ParseMiterLimit(app.OpenVG, values[v]); err != nil {
			return nil, err
		}
		v += 1

		// Stroke Width
		if err := op.ParseStrokeWidth(app.OpenVG, values[v]); err != nil {
			return nil, err
		}
		v += 1

		// Colors
		if err := op.ParseStrokeColor(app.OpenVG, values[v], values[v+1], values[v+2]); err != nil {
			return nil, err
		}
		v += 3
		if err := op.ParseFillColor(app.OpenVG, values[v], values[v+1], values[v+2]); err != nil {
			return nil, err
		}
		v += 3

		// Path elements
		elements := int(values[v])
		if op.path, err = app.OpenVG.CreatePath(); err != nil {
			return nil, err
		}
		v += 1

		for i := 0; i < elements; i++ {
			vinc, err := op.ParsePathPoint(app.OpenVG, opcodes[c], values, v)
			if err != nil {
				return nil, err
			}
			c += 1
			v += vinc
		}

		// Append the OP into the array of ops
		operations = append(operations, op)
	}

	return operations, nil
}

////////////////////////////////////////////////////////////////////////////////

func MyRunLoop(app *app.App) error {
	args := app.FlagSet.Args()
	if len(args) != 1 {
		return app.Logger.Error("Missing data filename")
	}

	// Create background surface
	surface, err := app.EGL.CreateBackground("OpenVG", 1.0)
	if err != nil {
		return err
	}
	defer app.EGL.DestroySurface(surface)

	// Read operations
	operations, err := ProcessDataFromFile(app, args[0])
	if err != nil {
		return err
	}

	r := float32(0)
	for {
		// Draw
		app.OpenVG.Do(surface, func() error {
			app.OpenVG.Clear(surface, khronos.VGColorLightGrey)
			app.OpenVG.Rotate(r)
			app.OpenVG.Scale(0.5, 0.5)
			for _, op := range operations {
				if err := op.path.Draw(op.stroke, op.fill); err != nil {
					return err
				}
			}
			return nil
		})
		r = r + 0.5
	}

	// Wait until interrupted
	app.WaitUntilDone()

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the config
	config := app.Config(app.APP_EGL | app.APP_OPENVG)
	config.FlagSet.FlagFloat64("opacity", 1.0, "Image opacity, 0.0 -> 1.0")

	// Create the application
	myapp, err := app.NewApp(config)
	if err == app.ErrHelp {
		return
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
	defer myapp.Close()

	// Run the application
	if err := myapp.Run(MyRunLoop); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}
}
