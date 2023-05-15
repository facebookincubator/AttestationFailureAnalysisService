// Code generated by Thrift Compiler (0.14.0). DO NOT EDIT.

package device

import (
	"bytes"
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"time"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = context.Background
var _ = time.Now
var _ = bytes.Equal

// Attributes:
//   - AssetID
//   - ModelID
//   - Hostname
type Device struct {
	AssetID  int64   `thrift:"AssetID,1" db:"AssetID" json:"AssetID"`
	ModelID  int64   `thrift:"ModelID,2" db:"ModelID" json:"ModelID"`
	Hostname *string `thrift:"Hostname,3" db:"Hostname" json:"Hostname,omitempty"`
}

func NewDevice() *Device {
	return &Device{}
}

func (p *Device) GetAssetID() int64 {
	return p.AssetID
}

func (p *Device) GetModelID() int64 {
	return p.ModelID
}

var Device_Hostname_DEFAULT string

func (p *Device) GetHostname() string {
	if !p.IsSetHostname() {
		return Device_Hostname_DEFAULT
	}
	return *p.Hostname
}
func (p *Device) IsSetHostname() bool {
	return p.Hostname != nil
}

func (p *Device) Read(ctx context.Context, iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(ctx); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if fieldTypeId == thrift.I64 {
				if err := p.ReadField1(ctx, iprot); err != nil {
					return err
				}
			} else {
				if err := iprot.Skip(ctx, fieldTypeId); err != nil {
					return err
				}
			}
		case 2:
			if fieldTypeId == thrift.I64 {
				if err := p.ReadField2(ctx, iprot); err != nil {
					return err
				}
			} else {
				if err := iprot.Skip(ctx, fieldTypeId); err != nil {
					return err
				}
			}
		case 3:
			if fieldTypeId == thrift.STRING {
				if err := p.ReadField3(ctx, iprot); err != nil {
					return err
				}
			} else {
				if err := iprot.Skip(ctx, fieldTypeId); err != nil {
					return err
				}
			}
		default:
			if err := iprot.Skip(ctx, fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(ctx); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(ctx); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *Device) ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(ctx); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.AssetID = v
	}
	return nil
}

func (p *Device) ReadField2(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(ctx); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.ModelID = v
	}
	return nil
}

func (p *Device) ReadField3(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(ctx); err != nil {
		return thrift.PrependError("error reading field 3: ", err)
	} else {
		p.Hostname = &v
	}
	return nil
}

func (p *Device) Write(ctx context.Context, oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin(ctx, "Device"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if p != nil {
		if err := p.writeField1(ctx, oprot); err != nil {
			return err
		}
		if err := p.writeField2(ctx, oprot); err != nil {
			return err
		}
		if err := p.writeField3(ctx, oprot); err != nil {
			return err
		}
	}
	if err := oprot.WriteFieldStop(ctx); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(ctx); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *Device) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin(ctx, "AssetID", thrift.I64, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:AssetID: ", p), err)
	}
	if err := oprot.WriteI64(ctx, int64(p.AssetID)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.AssetID (1) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(ctx); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:AssetID: ", p), err)
	}
	return err
}

func (p *Device) writeField2(ctx context.Context, oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin(ctx, "ModelID", thrift.I64, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:ModelID: ", p), err)
	}
	if err := oprot.WriteI64(ctx, int64(p.ModelID)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.ModelID (2) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(ctx); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:ModelID: ", p), err)
	}
	return err
}

func (p *Device) writeField3(ctx context.Context, oprot thrift.TProtocol) (err error) {
	if p.IsSetHostname() {
		if err := oprot.WriteFieldBegin(ctx, "Hostname", thrift.STRING, 3); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 3:Hostname: ", p), err)
		}
		if err := oprot.WriteString(ctx, string(*p.Hostname)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.Hostname (3) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(ctx); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 3:Hostname: ", p), err)
		}
	}
	return err
}

func (p *Device) Equals(other *Device) bool {
	if p == other {
		return true
	} else if p == nil || other == nil {
		return false
	}
	if p.AssetID != other.AssetID {
		return false
	}
	if p.ModelID != other.ModelID {
		return false
	}
	if p.Hostname != other.Hostname {
		if p.Hostname == nil || other.Hostname == nil {
			return false
		}
		if (*p.Hostname) != (*other.Hostname) {
			return false
		}
	}
	return true
}

func (p *Device) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Device(%+v)", *p)
}