package state

import (
	"fmt"
	"github.com/goburrow/modbus"
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
	"time"
)

var (
	Modbus = appdef.NewQName(appdef.SysPackage, "Modbus")
)

// TemperatureSensor represents a Modbus-connected temperature sensor.
type modbusStorage struct {
	modbus modbusClient
}

type modbusClient struct {
	handler *modbus.TCPClientHandler
}

type modBusKeyBuilder struct {
	*keyBuilder
}

type modbusValue struct {
	istructs.IStateValue
	data map[string]interface{}
}

type IModbusClient interface {
	Connect(ip string, port string, slaveID byte) error
	ReadData() (float64, float64, error)
	Close()
}

func (ms *modbusStorage) NewKeyBuilder(appdef.QName, istructs.IStateKeyBuilder) istructs.IStateKeyBuilder {
	return &modBusKeyBuilder{
		keyBuilder: newKeyBuilder(Modbus, appdef.NullQName),
	}
}

func (mkb *modBusKeyBuilder) host() string {
	if val, ok := mkb.data["Host"]; ok {
		return val.(string)
	}
	panic("Host not set")
}

func (mkb *modBusKeyBuilder) port() string {
	if val, ok := mkb.data["Port"]; ok {
		return val.(string)
	}
	panic("Port not set")
}

func (mkb *modBusKeyBuilder) slaveID() byte {
	if val, ok := mkb.data["SlaveID"]; ok {
		return byte(val.(int32))
	}
	panic("SlaveID not set")
}

func (ms *modbusStorage) Read(key istructs.IStateKeyBuilder, callback istructs.ValueCallback) (err error) {
	var (
		t, h float64
	)
	kb := key.(*modBusKeyBuilder)
	err = ms.modbus.Connect(kb.host(), kb.port(), kb.slaveID())
	if err != nil {
		return err
	}
	defer ms.modbus.Close()
	t, h, err = ms.modbus.ReadData()
	if err != nil {
		return err
	}
	return callback(nil,
		&modbusValue{data: map[string]interface{}{
			"T": t,
			"H": h,
		},
		},
	)
}

// Connect establishes a connection to the Modbus device.
func (mc *modbusClient) Connect(ip string, port string, slaveID byte) error {
	mc.handler = modbus.NewTCPClientHandler(fmt.Sprintf("%s:%s", ip, port))
	mc.handler.Timeout = 10 * time.Second
	mc.handler.SlaveId = slaveID
	return mc.handler.Connect()
}

// ReadData reads the temperature from the sensor.
func (mc *modbusClient) ReadData() (float64, float64, error) {
	if mc.handler == nil {
		return 0, 0, fmt.Errorf("device not connected")
	}
	client := modbus.NewClient(mc.handler)

	//	results, err := client.ReadHoldingRegisters(0, 1)
	results, err := client.ReadInputRegisters(1, 2)
	if err != nil {
		return 0, 0, err
	}

	temperature := float64(int16(results[0])<<8 | int16(results[1]))
	humidity := float64(int16(results[2])<<8 | int16(results[3]))
	return temperature, humidity, nil
}

// Close closes the connection to the Modbus device.
func (mc *modbusClient) Close() {
	if mc.handler != nil {
		mc.handler.Close()
	}
}
