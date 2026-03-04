package types

import (
	"fmt"
	"io"
	math_bits "math/bits"

	proto "github.com/cosmos/gogoproto/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math_bits.Len64

// ---------------------------------------------------------------------------
// QueryGetLinkedIdentitiesRequest
// ---------------------------------------------------------------------------

type QueryGetLinkedIdentitiesRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Tld  string `protobuf:"bytes,2,opt,name=tld,proto3" json:"tld,omitempty"`
}

func (m *QueryGetLinkedIdentitiesRequest) Reset()         { *m = QueryGetLinkedIdentitiesRequest{} }
func (m *QueryGetLinkedIdentitiesRequest) String() string { return proto.CompactTextString(m) }
func (*QueryGetLinkedIdentitiesRequest) ProtoMessage()    {}

func (m *QueryGetLinkedIdentitiesRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *QueryGetLinkedIdentitiesRequest) GetTld() string {
	if m != nil {
		return m.Tld
	}
	return ""
}

func (m *QueryGetLinkedIdentitiesRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryGetLinkedIdentitiesRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryGetLinkedIdentitiesRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 2: tld (string)
	if len(m.Tld) > 0 {
		i -= len(m.Tld)
		copy(dAtA[i:], m.Tld)
		i = encodeVarintLinkedIdentities(dAtA, i, uint64(len(m.Tld)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: name (string)
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintLinkedIdentities(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryGetLinkedIdentitiesRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Name) > 0 {
		n += 1 + len(m.Name) + sovLinkedIdentities(uint64(len(m.Name)))
	}
	if len(m.Tld) > 0 {
		n += 1 + len(m.Tld) + sovLinkedIdentities(uint64(len(m.Tld)))
	}
	return n
}

func (m *QueryGetLinkedIdentitiesRequest) Unmarshal(dAtA []byte) error {
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
		case 1: // name
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
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
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // tld
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tld", wireType)
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
			m.Tld = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLinkedIdentities(dAtA[iNdEx:])
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

// ---------------------------------------------------------------------------
// QueryGetLinkedIdentitiesResponse
// ---------------------------------------------------------------------------

type QueryGetLinkedIdentitiesResponse struct {
	Identities []LinkedIdentity `protobuf:"bytes,1,rep,name=identities,proto3" json:"identities"`
}

func (m *QueryGetLinkedIdentitiesResponse) Reset()         { *m = QueryGetLinkedIdentitiesResponse{} }
func (m *QueryGetLinkedIdentitiesResponse) String() string { return proto.CompactTextString(m) }
func (*QueryGetLinkedIdentitiesResponse) ProtoMessage()    {}

func (m *QueryGetLinkedIdentitiesResponse) GetIdentities() []LinkedIdentity {
	if m != nil {
		return m.Identities
	}
	return nil
}

func (m *QueryGetLinkedIdentitiesResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryGetLinkedIdentitiesResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryGetLinkedIdentitiesResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 1: identities (repeated message)
	if len(m.Identities) > 0 {
		for iNdEx := len(m.Identities) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Identities[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintLinkedIdentities(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryGetLinkedIdentitiesResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Identities) > 0 {
		for _, e := range m.Identities {
			l := e.Size()
			n += 1 + l + sovLinkedIdentities(uint64(l))
		}
	}
	return n
}

func (m *QueryGetLinkedIdentitiesResponse) Unmarshal(dAtA []byte) error {
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
		case 1: // identities
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identities", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return fmt.Errorf("proto: negative length found during unmarshaling")
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Identities = append(m.Identities, LinkedIdentity{})
			if err := m.Identities[len(m.Identities)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLinkedIdentities(dAtA[iNdEx:])
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

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func encodeVarintLinkedIdentities(dAtA []byte, offset int, v uint64) int {
	offset -= sovLinkedIdentities(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func sovLinkedIdentities(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func skipLinkedIdentities(dAtA []byte) (n int, err error) {
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
	proto.RegisterType((*QueryGetLinkedIdentitiesRequest)(nil), "xid.v1.QueryGetLinkedIdentitiesRequest")
	proto.RegisterType((*QueryGetLinkedIdentitiesResponse)(nil), "xid.v1.QueryGetLinkedIdentitiesResponse")
}
