package can

import (
	"encoding/binary"
	"fmt"

	"go.einride.tech/can"
)

var servicePIDS = map[uint8]map[uint16]string{
	0x01: {
		0x0:  "PIDs supported [$01 - $20]",
		0x1:  "Monitor status since DTCs cleared.",
		0x2:  "DTC that caused freeze frame to be stored.",
		0x3:  "Fuel system status",
		0x4:  "Calculated engine load",
		0x5:  "Engine coolant temperature",
		0x6:  "Short term fuel trim (STFT)—Bank 1",
		0x7:  "Long term fuel trim (LTFT)—Bank 1",
		0x8:  "Short term fuel trim (STFT)—Bank 2",
		0x9:  "Long term fuel trim (LTFT)—Bank 2",
		0x0A: "Fuel pressure (gauge pressure)",
		0x0B: "Intake manifold absolute pressure",
		0x0C: "Engine speed",
		0x0D: "Vehicle speed",
		0x0E: "Timing advance",
		0x0F: "Intake air temperature",
		0x10: "Mass air flow sensor (MAF) air flow rate",
		0x11: "Throttle position",
		0x12: "Commanded secondary air status",
		0x13: "Oxygen sensors present",
		0x14: "Oxygen Sensor 1",
		0x15: "Oxygen Sensor 2",
		0x16: "Oxygen Sensor 3",
		0x17: "Oxygen Sensor 4",
		0x18: "Oxygen Sensor 5",
		0x19: "Oxygen Sensor 6",
		0x1A: "Oxygen Sensor 7",
		0x1B: "Oxygen Sensor 8",
		0x1C: "OBD standards this vehicle conforms to",
		0x1D: "Oxygen sensors present",
		0x1E: "Auxiliary input status",
		0x1F: "Run time since engine start",
		0x20: "PIDs supported [$21 - $40]",
		0x21: "Distance traveled with malfunction indicator lamp (MIL) on",
		0x22: "Fuel Rail Pressure (relative to manifold vacuum)",
		0x23: "Fuel Rail Gauge Pressure (diesel, or gasoline direct injection)",
		0x24: "Oxygen Sensor 1",
		0x25: "Oxygen Sensor 2",
		0x26: "Oxygen Sensor 3",
		0x27: "Oxygen Sensor 4",
		0x28: "Oxygen Sensor 5",
		0x29: "Oxygen Sensor 6",
		0x2A: "Oxygen Sensor 7",
		0x2B: "Oxygen Sensor 8",
		0x2C: "Commanded EGR",
		0x2D: "EGR Error",
		0x2E: "Commanded evaporative purge",
		0x2F: "Fuel Tank Level Input",
		0x30: "Warm-ups since codes cleared",
		0x31: "Distance traveled since codes cleared",
		0x32: "Evap. System Vapor Pressure",
		0x33: "Absolute Barometric Pressure",
		0x34: "Oxygen Sensor 1",
		0x35: "Oxygen Sensor 2",
		0x36: "Oxygen Sensor 3",
		0x37: "Oxygen Sensor 4",
		0x38: "Oxygen Sensor 5",
		0x39: "Oxygen Sensor 6",
		0x3A: "Oxygen Sensor 7",
		0x3B: "Oxygen Sensor 8",
		0x3C: "Catalyst Temperature: Bank 1, Sensor 1",
		0x3D: "Catalyst Temperature: Bank 2, Sensor 1",
		0x3E: "Catalyst Temperature: Bank 1, Sensor 2",
		0x3F: "Catalyst Temperature: Bank 2, Sensor 2",
		0x40: "PIDs supported [$41 - $60]",
		0x41: "Monitor status this drive cycle",
		0x42: "Control module voltage",
		0x43: "Absolute load value",
		0x44: "Commanded Air-Fuel Equivalence Ratio (lambda,λ)",
		0x45: "Relative throttle position",
		0x46: "Ambient air temperature",
		0x47: "Absolute throttle position B",
		0x48: "Absolute throttle position C",
		0x49: "Accelerator pedal position D",
		0x4A: "Accelerator pedal position E",
		0x4B: "Accelerator pedal position F",
		0x4C: "Commanded throttle actuator",
		0x4D: "Time run with MIL on",
		0x4E: "Time since trouble codes cleared",
		0x4F: "Maximum value for Fuel–Air equivalence ratio, oxygen sensor voltage, oxygen sensor current, and intake manifold absolute pressure",
		0x50: "Maximum value for air flow rate from mass air flow sensor",
		0x51: "Fuel Type",
		0x52: "Ethanol fuel %",
		0x53: "Absolute Evap system Vapor Pressure",
		0x54: "Evap system vapor pressure",
		0x55: "Short term secondary oxygen sensor trim, A: bank 1, B: bank 3",
		0x56: "Long term secondary oxygen sensor trim, A: bank 1, B: bank 3",
		0x57: "Short term secondary oxygen sensor trim, A: bank 2, B: bank 4",
		0x58: "Long term secondary oxygen sensor trim, A: bank 2, B: bank 4",
		0x59: "Fuel rail absolute pressure",
		0x5A: "Relative accelerator pedal position",
		0x5B: "Hybrid battery pack remaining life",
		0x5C: "Engine oil temperature",
		0x5D: "Fuel injection timing",
		0x5E: "Engine fuel rate",
		0x5F: "Emission requirements to which vehicle is designed",
		0x60: "PIDs supported [$61 - $80]",
		0x61: "Driver's demand engine - percent torque",
		0x62: "Actual engine - percent torque",
		0x63: "Engine reference torque",
		0x64: "Engine percent torque data",
		0x65: "Auxiliary input / output supported",
		0x66: "Mass air flow sensor",
		0x67: "Engine coolant temperature",
		0x68: "Intake air temperature sensor",
		0x69: "Actual EGR, Commanded EGR, and EGR Error",
		0x6A: "Commanded Diesel intake air flow control and relative intake air flow position",
		0x6B: "Exhaust gas recirculation temperature",
		0x6C: "Commanded throttle actuator control and relative throttle position",
		0x6D: "Fuel pressure control system",
		0x6E: "Injection pressure control system",
		0x6F: "Turbocharger compressor inlet pressure",
		0x70: "Boost pressure control",
		0x71: "Variable Geometry turbo (VGT) control",
		0x72: "Wastegate control",
		0x73: "Exhaust pressure",
		0x74: "Turbocharger RPM",
		0x75: "Turbocharger temperature",
		0x76: "Turbocharger temperature",
		0x77: "Charge air cooler temperature (CACT)",
		0x78: "Exhaust Gas temperature (EGT) Bank 1",
		0x79: "Exhaust Gas temperature (EGT) Bank 2",
		0x7A: "Diesel particulate filter (DPF)differential pressure",
		0x7B: "Diesel particulate filter (DPF)",
		0x7C: "Diesel Particulate filter (DPF) temperature",
		0x7D: "NOx NTE (Not-To-Exceed) control area status",
		0x7E: "PM NTE (Not-To-Exceed) control area status",
		0x7F: "Engine run time [b]",
		0x80: "PIDs supported [$81 - $A0]",
		0x81: "Engine run time for Auxiliary Emissions Control Device(AECD)",
		0x82: "Engine run time for Auxiliary Emissions Control Device(AECD)",
		0x83: "NOx sensor",
		0x84: "Manifold surface temperature",
		0x85: "NOx reagent system",
		0x86: "Particulate matter (PM) sensor",
		0x87: "Intake manifold absolute pressure",
		0x88: "SCR Induce System",
		0x89: "Run Time for AECD #11-#15",
		0x8A: "Run Time for AECD #16-#20",
		0x8B: "Diesel Aftertreatment",
		0x8C: "O2 Sensor (Wide Range)",
		0x8D: "Throttle Position G",
		0x8E: "Engine Friction - Percent Torque",
		0x8F: "PM Sensor Bank 1 & 2",
		0x90: "WWH-OBD Vehicle OBD System Information",
		0x91: "WWH-OBD Vehicle OBD System Information",
		0x92: "Fuel System Control",
		0x93: "WWH-OBD Vehicle OBD Counters support",
		0x94: "NOx Warning And Inducement System",
		0x98: "Exhaust Gas Temperature Sensor",
		0x99: "Exhaust Gas Temperature Sensor",
		0x9A: "Hybrid/EV Vehicle System Data, Battery, Voltage",
		0x9B: "Diesel Exhaust Fluid Sensor Data",
		0x9C: "O2 Sensor Data",
		0x9D: "Engine Fuel Rate",
		0x9E: "Engine Exhaust Flow Rate",
		0x9F: "Fuel System Percentage Use",
		0xA0: "PIDs supported [$A1 - $C0]",
		0xA1: "NOx Sensor Corrected Data",
		0xA2: "Cylinder Fuel Rate",
		0xA3: "Evap System Vapor Pressure",
		0xA4: "Transmission Actual Gear",
		0xA5: "Commanded Diesel Exhaust Fluid Dosing",
		0xA6: "Odometer [c]",
		0xA7: "NOx Sensor Concentration Sensors 3 and 4",
		0xA8: "NOx Sensor Corrected Concentration Sensors 3 and 4",
		0xA9: "ABS Disable Switch State",
		0xC0: "PIDs supported [$C1 - $E0]",
		0xC3: "Fuel Level Input A/B",
		0xC4: "Exhaust Particulate Control System Diagnostic Time/Count",
		0xC5: "Fuel Pressure A and B",
		0xC6: "Multiple system counters",
		0xC7: "Distance Since Reflash or Module Replacement",
		0xC8: "NOx Control Diagnostic (NCD) and Particulate Control Diagnostic (PCD) Warning Lamp status",
	},
}

type OBD2PID struct {
	ID   uint16
	Name string
}

func (p OBD2PID) String() string {
	if p.Name != "" {
		return p.Name
	}
	return fmt.Sprintf("pid 0x%d", p.ID)
}

func lookupPID(svcID uint8, data []uint8) OBD2PID {
	if len(data) == 1 {
		data = []byte{
			0x00,
			data[0],
		}
	}

	pid := OBD2PID{
		ID: binary.BigEndian.Uint16(data),
	}

	// resolve service
	if svc, found := servicePIDS[svcID]; found {
		// resolve PID name
		if name, found := svc[pid.ID]; found {
			pid.Name = name
		}
	}

	return pid
}

func (msg *OBD2Message) ParseRequest(frame can.Frame) bool {
	svcID := frame.Data[1]
	// validate service / mode
	if svcID > 0x0a {
		return false
	}

	msgSize := frame.Data[0]
	// validate data size
	if msgSize > 6 {
		return false
	}

	data := frame.Data[2 : 1+msgSize]

	msg.PID = lookupPID(svcID, data)
	msg.Type = OBD2MessageTypeRequest
	msg.ECU = 0xff // broadcast
	msg.Size = msgSize - 1
	msg.Service = OBD2Service(svcID)
	msg.Data = data

	return true
}
