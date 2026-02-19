package types

import (
	"fmt"
	"io"
	math_bits "math/bits"

	proto "github.com/cosmos/gogoproto/proto"
)

// FullName returns the complete domain name (e.g. "alice.epix")
func (n NameRecord) FullName() string {
	return n.Name + "." + n.Tld
}

// DNS record type constants
const (
	DNSRecordTypeA     uint32 = 1
	DNSRecordTypeAAAA  uint32 = 28
	DNSRecordTypeCNAME uint32 = 5
	DNSRecordTypeTXT   uint32 = 16
	DNSRecordTypeMX    uint32 = 15
	DNSRecordTypeNS    uint32 = 2
	DNSRecordTypeSRV   uint32 = 33
)

// EpixNetPeer represents an EpixNet peer address attached to a name
type EpixNetPeer struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Label   string `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	AddedAt uint64 `protobuf:"varint,3,opt,name=added_at,json=addedAt,proto3" json:"added_at,omitempty"`
}

func (m *EpixNetPeer) Reset()         { *m = EpixNetPeer{} }
func (m *EpixNetPeer) String() string { return proto.CompactTextString(m) }
func (*EpixNetPeer) ProtoMessage()    {}
func (*EpixNetPeer) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bd3daa00c1847cd, []int{6}
}
func (m *EpixNetPeer) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EpixNetPeer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EpixNetPeer.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EpixNetPeer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EpixNetPeer.Merge(m, src)
}
func (m *EpixNetPeer) XXX_Size() int {
	return m.Size()
}
func (m *EpixNetPeer) XXX_DiscardUnknown() {
	xxx_messageInfo_EpixNetPeer.DiscardUnknown(m)
}

var xxx_messageInfo_EpixNetPeer proto.InternalMessageInfo

func (m *EpixNetPeer) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *EpixNetPeer) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *EpixNetPeer) GetAddedAt() uint64 {
	if m != nil {
		return m.AddedAt
	}
	return 0
}

func (m *EpixNetPeer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EpixNetPeer) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EpixNetPeer) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 3: added_at (uint64, varint)
	if m.AddedAt != 0 {
		i = encodeVarintTypes(dAtA, i, m.AddedAt)
		i--
		dAtA[i] = 0x18
	}
	// field 2: label (string)
	if len(m.Label) > 0 {
		i -= len(m.Label)
		copy(dAtA[i:], m.Label)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Label)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: address (string)
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EpixNetPeer) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Address) > 0 {
		n += 1 + len(m.Address) + sovTypes(uint64(len(m.Address)))
	}
	if len(m.Label) > 0 {
		n += 1 + len(m.Label) + sovTypes(uint64(len(m.Label)))
	}
	if m.AddedAt != 0 {
		n += 1 + sovTypes(m.AddedAt)
	}
	return n
}

func (m *EpixNetPeer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1: // address
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + int(stringLen)
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // label
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Label", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + int(stringLen)
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Label = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3: // added_at
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AddedAt", wireType)
			}
			m.AddedAt = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AddedAt |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return fmt.Errorf("proto: negative skip found during unmarshaling")
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func encodeVarintTypes(dAtA []byte, offset int, v uint64) int {
	offset -= sovTypes(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func sovTypes(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func skipTypes(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, fmt.Errorf("proto: negative length found during unmarshaling")
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, fmt.Errorf("proto: unexpected end group")
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, fmt.Errorf("proto: negative position after skip")
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

func init() {
	proto.RegisterType((*EpixNetPeer)(nil), "xid.v1.EpixNetPeer")
}
