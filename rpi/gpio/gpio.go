/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE.md

	You may need to add your user to the gpio group
	sudo usermod -a -G gpio ${USER}
	then logout and log back in again
*/
package gpio

import (
	"os"
	"sync"
	"syscall"
	"github.com/djthorpe/gopi/rpi"
)



////////////////////////////////////////////////////////////////////////////////

type LogicalPin uint
type PhysicalPin uint

type Pin struct{
	Name string
	LogicalPin  LogicalPin
	PhysicalPin []PhysicalPin // there can be more than one physical pin number
	                          // in the case of voltage pins
}

type Pins struct{
	bytes []byte              // memory mapped byte array
	NumberOfPins uint         // number of physical pins for GPIO connector
}

////////////////////////////////////////////////////////////////////////////////

const (
	GPIO_DEV_GPIO = "/dev/gpiomem"
	GPIO_DEV_MEM = "/dev/mem"
	GPIO_SIZE_BYTES = 4096
	GPIO_OFFSET = 0x200000
)

// Pin modes
const (
	GPIO_PIN_MODE_UNKNOWN = iota
	GPIO_PIN_MODE_IN               // Input
	GPIO_PIN_MODE_OUT              // Output
	GPIO_PIN_MODE_ALT0             // Alternate function 0
	GPIO_PIN_MODE_ALT1             // Alternate function 1
	GPIO_PIN_MODE_ALT2             // Alternate function 2
	GPIO_PIN_MODE_ALT3             // Alternate function 3
	GPIO_PIN_MODE_ALT4             // Alternate function 4
	GPIO_PIN_MODE_ALT5             // Alternate function 5
)

// Logical GPIO pins
const (
	GPIO_PIN_UNKNOWN = iota
	GPIO_PIN_PWR3V3
	GPIO_PIN_PWR5V
	GPIO_PIN_GROUND
	GPIO_PIN_BCM0
	GPIO_PIN_BCM1
	GPIO_PIN_BCM2
	GPIO_PIN_BCM3
	GPIO_PIN_BCM4
	GPIO_PIN_BCM5
	GPIO_PIN_BCM6
	GPIO_PIN_BCM7
	GPIO_PIN_BCM8
	GPIO_PIN_BCM9
	GPIO_PIN_BCM10
	GPIO_PIN_BCM11
	GPIO_PIN_BCM12
	GPIO_PIN_BCM13
	GPIO_PIN_BCM14
	GPIO_PIN_BCM15
	GPIO_PIN_BCM16
	GPIO_PIN_BCM17
	GPIO_PIN_BCM18
	GPIO_PIN_BCM19
	GPIO_PIN_BCM20
	GPIO_PIN_BCM21
	GPIO_PIN_BCM22
	GPIO_PIN_BCM23
	GPIO_PIN_BCM24
	GPIO_PIN_BCM25
	GPIO_PIN_BCM26
	GPIO_PIN_BCM27
)

////////////////////////////////////////////////////////////////////////////////

// map product to number of pins
var productpinsmap = map[rpi.Product]uint{
    rpi.RPI_MODEL_UNKNOWN: 0,
	rpi.RPI_MODEL_A: 26,
    rpi.RPI_MODEL_B: 26,
    rpi.RPI_MODEL_A_PLUS: 40,
    rpi.RPI_MODEL_B_PLUS: 40,
    rpi.RPI_MODEL_B_PI_2: 40,
    rpi.RPI_MODEL_ZERO: 40,
}

// map physical pin to PinNumber
var logicalpinmap = map[PhysicalPin]LogicalPin {
	1: GPIO_PIN_PWR3V3,
	2: GPIO_PIN_PWR5V,
	3: GPIO_PIN_BCM2,
	4: GPIO_PIN_PWR5V,
	5: GPIO_PIN_BCM3,
	6: GPIO_PIN_GROUND,
	7: GPIO_PIN_BCM4,
	8: GPIO_PIN_BCM14,
	9: GPIO_PIN_GROUND,
	10: GPIO_PIN_BCM15,
	11: GPIO_PIN_BCM17,
	12: GPIO_PIN_BCM18,
	13: GPIO_PIN_BCM27,
	14: GPIO_PIN_GROUND,
	15: GPIO_PIN_BCM22,
	16: GPIO_PIN_BCM23,
	17: GPIO_PIN_PWR3V3,
	18: GPIO_PIN_BCM24,
	19: GPIO_PIN_BCM10,
	20: GPIO_PIN_GROUND,
	21: GPIO_PIN_BCM9,
	22: GPIO_PIN_BCM25,
	23: GPIO_PIN_BCM11,
	24: GPIO_PIN_BCM8,
	25: GPIO_PIN_GROUND,
	26: GPIO_PIN_BCM7,
	27: GPIO_PIN_BCM0,
	28: GPIO_PIN_BCM1,
	29: GPIO_PIN_BCM5,
	30: GPIO_PIN_GROUND,
	31: GPIO_PIN_BCM6,
	32: GPIO_PIN_BCM12,
	33: GPIO_PIN_BCM13,
	34: GPIO_PIN_GROUND,
	35: GPIO_PIN_BCM19,
	36: GPIO_PIN_BCM16,
	37: GPIO_PIN_BCM26,
	38: GPIO_PIN_BCM20,
	39: GPIO_PIN_GROUND,
	40: GPIO_PIN_BCM21,
}

var logicalpinnamemap = map[LogicalPin]string {
	GPIO_PIN_UNKNOWN: "unknown",
	GPIO_PIN_PWR3V3: "3V3",
	GPIO_PIN_PWR5V: "5V",
	GPIO_PIN_GROUND: "GND",
	GPIO_PIN_BCM0: "GPIO0",
	GPIO_PIN_BCM1: "GPIO1",
	GPIO_PIN_BCM2: "GPIO2",
	GPIO_PIN_BCM3: "GPIO3",
	GPIO_PIN_BCM4: "GPIO4",
	GPIO_PIN_BCM5: "GPIO5",
	GPIO_PIN_BCM6: "GPIO6",
	GPIO_PIN_BCM7: "GPIO7",
	GPIO_PIN_BCM8: "GPIO8",
	GPIO_PIN_BCM9: "GPIO9",
	GPIO_PIN_BCM10: "GPIO10",
	GPIO_PIN_BCM11: "GPIO11",
	GPIO_PIN_BCM12: "GPIO12",
	GPIO_PIN_BCM13: "GPIO13",
	GPIO_PIN_BCM14: "GPIO14",
	GPIO_PIN_BCM15: "GPIO15",
	GPIO_PIN_BCM16: "GPIO16",
	GPIO_PIN_BCM17: "GPIO17",
	GPIO_PIN_BCM18: "GPIO18",
	GPIO_PIN_BCM19: "GPIO19",
	GPIO_PIN_BCM20: "GPIO20",
	GPIO_PIN_BCM21: "GPIO21",
	GPIO_PIN_BCM22: "GPIO22",
	GPIO_PIN_BCM23: "GPIO23",
	GPIO_PIN_BCM24: "GPIO24",
	GPIO_PIN_BCM25: "GPIO25",
	GPIO_PIN_BCM26: "GPIO26",
	GPIO_PIN_BCM27: "GPIO27",
}

////////////////////////////////////////////////////////////////////////////////

var (
	memlock sync.Mutex
)

////////////////////////////////////////////////////////////////////////////////

func New(model *rpi.Model) (*Pins,error) {

	// check model
	if model == nil || model.Product == rpi.RPI_MODEL_UNKNOWN {
		return nil,rpi.ErrorUnknownProduct
	}

	// create 'this' object
	this := new(Pins)
	this.bytes = nil
	this.NumberOfPins = uint(productpinsmap[model.Product])
	if this.NumberOfPins == 0 {
		return nil,rpi.ErrorUnknownProduct
	}

	// initialize
	var file *os.File
	var err error
	if file, err = os.OpenFile(GPIO_DEV_GPIO,os.O_RDWR|os.O_SYNC,0); err != nil {
		if os.IsNotExist(err) {
			file, err = os.OpenFile(GPIO_DEV_MEM,os.O_RDWR|os.O_SYNC,0)
		}
	}

	if err != nil {
		return nil,err
	}

	// file descriptor can be closed after memory mapping
	defer file.Close()

	// lock for memmap
	memlock.Lock()
	defer memlock.Unlock()

	// memory map GPIO registers to byte array
	base := model.PeripheralBase + GPIO_OFFSET
	this.bytes, err = syscall.Mmap(int(file.Fd()),int64(base),GPIO_SIZE_BYTES,syscall.PROT_READ|syscall.PROT_WRITE,syscall.MAP_SHARED)
	if err != nil {
		return nil,err
	}

	// Return this
	return this,nil
}

func (this *Pins) Terminate() {
	// lock for memmap
	memlock.Lock()
	defer memlock.Unlock()
	// unmap memory
	if this.bytes != nil {
		syscall.Munmap(this.bytes)
		this.bytes = nil
	}
}

// Return pin from logical pin name
func (this *Pins) getLogicalPin(pin LogicalPin) (*Pin,error) {
/*	// create the pins from the physical pin connections
	for p := uint(1); p <= this.NumberOfPins; p++ {
		l := LogicalPin(logicalpinmap[PhysicalPin(p)])
		n := logicalpinnamemap[l]
		this.pin[l] = Pin{ n,l,PhysicalPin(p) }
	}
*/
	// TODO
	return nil,nil
}

// Return pin from physical pin location
func (this *Pins) getPhysicalPin(pin PhysicalPin) (*Pin,error) {
	// Sanity check the physical pin
	if uint(pin) < 1 || uint(pin) > this.NumberOfPins {
		return nil,rpi.ErrorUnknownPin
	}
	logicalPin := logicalpinmap[pin]
	// Sanity check the logical pin
	if logicalPin == GPIO_PIN_UNKNOWN {
		return nil,rpi.ErrorUnknownPin
	}
	// Return the pin
	return &Pin{ logicalpinnamemap[logicalPin],logicalPin,[]PhysicalPin{ pin } },nil
}

// Return pin for Logical or Physical pin
func (this *Pins) GetPin(pin interface{}) (*Pin,error) {
	switch pin.(type) {
    case LogicalPin:
        return this.getLogicalPin(pin.(LogicalPin))
    case PhysicalPin:
        return this.getPhysicalPin(pin.(PhysicalPin))
    default:
        return nil,rpi.ErrorUnknownPin
    }
}

/*

const (
	// Physical addresses for various peripheral register sets

	// Base Physical Address of the BCM 2835 peripheral registers
	BCM2835_PERI_BASE = 0x20000000

	// Base Physical Address of the System Timer registers
	BCM2835_ST_BASE = BCM2835_PERI_BASE + 0x3000

	// Base Physical Address of the Pads registers
	BCM2835_GPIO_PADS = BCM2835_PERI_BASE + 0x100000

	// Base Physical Address of the Clock/timer registers
	BCM2835_CLOCK_BASE = BCM2835_PERI_BASE + 0x101000

	// Base Physical Address of the GPIO registers
	BCM2835_GPIO_BASE = BCM2835_PERI_BASE + 0x200000

	// Base Physical Address of the SPI0 registers
	BCM2835_SPI0_BASE = BCM2835_PERI_BASE + 0x204000

	// Base Physical Address of the BSC0 registers
	BCM2835_BSC0_BASE = BCM2835_PERI_BASE + 0x205000

	// Base Physical Address of the PWM registers
	BCM2835_GPIO_PWM = BCM2835_PERI_BASE + 0x20C000

	// Base Physical Address of the BSC1 registers
	BCM2835_BSC1_BASE = BCM2835_PERI_BASE + 0x804000

	// Size of memory page on RPi
	BCM2835_PAGE_SIZE = 4 * 1024

	// Size of memory block on RPi
	BCM2835_BLOCK_SIZE = 4 * 1024

	BCM2835_GPFSEL0   = 0x0000 // GPIO Function Select 0
	BCM2835_GPFSEL1   = 0x0004 // GPIO Function Select 1
	BCM2835_GPFSEL2   = 0x0008 // GPIO Function Select 2
	BCM2835_GPFSEL3   = 0x000c // GPIO Function Select 3
	BCM2835_GPFSEL4   = 0x0010 // GPIO Function Select 4
	BCM2835_GPFSEL5   = 0x0014 // GPIO Function Select 5
	BCM2835_GPSET0    = 0x001c // GPIO Pin Output Set 0
	BCM2835_GPSET1    = 0x0020 // GPIO Pin Output Set 1
	BCM2835_GPCLR0    = 0x0028 // GPIO Pin Output Clear 0
	BCM2835_GPCLR1    = 0x002c // GPIO Pin Output Clear 1
	BCM2835_GPLEV0    = 0x0034 // GPIO Pin Level 0
	BCM2835_GPLEV1    = 0x0038 // GPIO Pin Level 1
	BCM2835_GPEDS0    = 0x0040 // GPIO Pin Event Detect Status 0
	BCM2835_GPEDS1    = 0x0044 // GPIO Pin Event Detect Status 1
	BCM2835_GPREN0    = 0x004c // GPIO Pin Rising Edge Detect Enable 0
	BCM2835_GPREN1    = 0x0050 // GPIO Pin Rising Edge Detect Enable 1
	BCM2835_GPFEN0    = 0x0048 // GPIO Pin Falling Edge Detect Enable 0
	BCM2835_GPFEN1    = 0x005c // GPIO Pin Falling Edge Detect Enable 1
	BCM2835_GPHEN0    = 0x0064 // GPIO Pin High Detect Enable 0
	BCM2835_GPHEN1    = 0x0068 // GPIO Pin High Detect Enable 1
	BCM2835_GPLEN0    = 0x0070 // GPIO Pin Low Detect Enable 0
	BCM2835_GPLEN1    = 0x0074 // GPIO Pin Low Detect Enable 1
	BCM2835_GPAREN0   = 0x007c // GPIO Pin Async. Rising Edge Detect 0
	BCM2835_GPAREN1   = 0x0080 // GPIO Pin Async. Rising Edge Detect 1
	BCM2835_GPAFEN0   = 0x0088 // GPIO Pin Async. Falling Edge Detect 0
	BCM2835_GPAFEN1   = 0x008c // GPIO Pin Async. Falling Edge Detect 1
	BCM2835_GPPUD     = 0x0094 // GPIO Pin Pull-up/down Enable
	BCM2835_GPPUDCLK0 = 0x0098 // GPIO Pin Pull-up/down Enable Clock 0
	BCM2835_GPPUDCLK1 = 0x009c // GPIO Pin Pull-up/down Enable Clock 1


	GPIO_P1_12 = 18
	GPIO_P1_13 = 27
	GPIO_P1_15 = 22
	GPIO_P1_18 = 24
	GPIO_P1_22 = 25
	GPIO_P1_16 = 23
	GPIO_P1_11 = 17

	GPIO21 = GPIO_P1_13
	GPIO22 = GPIO_P1_15
	GPIO23 = GPIO_P1_16
	GPIO25 = GPIO_P1_22
	GPIO24 = GPIO_P1_18
	GPIO27 = GPIO_P1_13
	GPIO17 = GPIO_P1_11
)

*/


