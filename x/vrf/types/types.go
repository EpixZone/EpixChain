package types

import (
	"fmt"
	"io"
	math_bits "math/bits"

	proto "github.com/cosmos/gogoproto/proto"
)

// RandomBeacon represents a verifiable random beacon produced at a given block height
type RandomBeacon struct {
	Height    uint64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Beacon    string `protobuf:"bytes,2,opt,name=beacon,proto3" json:"beacon,omitempty"`
	Proposer  string `protobuf:"bytes,3,opt,name=proposer,proto3" json:"proposer,omitempty"`
	Timestamp int64  `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (m *RandomBeacon) Reset()         { *m = RandomBeacon{} }
func (m *RandomBeacon) String() string { return proto.CompactTextString(m) }
func (*RandomBeacon) ProtoMessage()    {}
func (m *RandomBeacon) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *RandomBeacon) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RandomBeacon.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *RandomBeacon) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RandomBeacon.Merge(m, src)
}

func (m *RandomBeacon) XXX_Size() int {
	return m.Size()
}

func (m *RandomBeacon) XXX_DiscardUnknown() {
	xxx_messageInfo_RandomBeacon.DiscardUnknown(m)
}

var xxx_messageInfo_RandomBeacon proto.InternalMessageInfo

func (m *RandomBeacon) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *RandomBeacon) GetBeacon() string {
	if m != nil {
		return m.Beacon
	}
	return ""
}

func (m *RandomBeacon) GetProposer() string {
	if m != nil {
		return m.Proposer
	}
	return ""
}

func (m *RandomBeacon) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *RandomBeacon) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RandomBeacon) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RandomBeacon) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 4: timestamp (int64, varint)
	if m.Timestamp != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Timestamp))
		i--
		dAtA[i] = 0x20
	}
	// field 3: proposer (string)
	if len(m.Proposer) > 0 {
		i -= len(m.Proposer)
		copy(dAtA[i:], m.Proposer)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Proposer)))
		i--
		dAtA[i] = 0x1a
	}
	// field 2: beacon (string)
	if len(m.Beacon) > 0 {
		i -= len(m.Beacon)
		copy(dAtA[i:], m.Beacon)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Beacon)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: height (uint64, varint)
	if m.Height != 0 {
		i = encodeVarintTypes(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *RandomBeacon) Size() (n int) {
	if m == nil {
		return 0
	}
	if m.Height != 0 {
		n += 1 + sovTypes(m.Height)
	}
	if len(m.Beacon) > 0 {
		n += 1 + len(m.Beacon) + sovTypes(uint64(len(m.Beacon)))
	}
	if len(m.Proposer) > 0 {
		n += 1 + len(m.Proposer) + sovTypes(uint64(len(m.Proposer)))
	}
	if m.Timestamp != 0 {
		n += 1 + sovTypes(uint64(m.Timestamp))
	}
	return n
}

func (m *RandomBeacon) Unmarshal(dAtA []byte) error {
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
		case 1: // height
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2: // beacon
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Beacon", wireType)
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
			m.Beacon = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3: // proposer
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proposer", wireType)
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
			m.Proposer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4: // timestamp
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= int64(b&0x7F) << shift
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

// Params defines the parameters for the VRF module
type Params struct {
	Enabled        bool   `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	LookbackBlocks uint64 `protobuf:"varint,2,opt,name=lookback_blocks,json=lookbackBlocks,proto3" json:"lookback_blocks,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}

func (m *Params) XXX_Size() int {
	return m.Size()
}

func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetEnabled() bool {
	if m != nil {
		return m.Enabled
	}
	return false
}

func (m *Params) GetLookbackBlocks() uint64 {
	if m != nil {
		return m.LookbackBlocks
	}
	return 0
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 2: lookback_blocks (uint64, varint)
	if m.LookbackBlocks != 0 {
		i = encodeVarintTypes(dAtA, i, m.LookbackBlocks)
		i--
		dAtA[i] = 0x10
	}
	// field 1: enabled (bool, varint)
	if m.Enabled {
		i--
		dAtA[i] = 1
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	if m.Enabled {
		n += 2
	}
	if m.LookbackBlocks != 0 {
		n += 1 + sovTypes(m.LookbackBlocks)
	}
	return n
}

func (m *Params) Unmarshal(dAtA []byte) error {
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
		case 1: // enabled
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Enabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Enabled = v != 0
		case 2: // lookback_blocks
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LookbackBlocks", wireType)
			}
			m.LookbackBlocks = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LookbackBlocks |= uint64(b&0x7F) << shift
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

// GenesisState defines the VRF module's genesis state
type GenesisState struct {
	Params  Params         `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Beacons []RandomBeacon `protobuf:"bytes,2,rep,name=beacons,proto3" json:"beacons"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}

func (m *GenesisState) XXX_Size() int {
	return m.Size()
}

func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetBeacons() []RandomBeacon {
	if m != nil {
		return m.Beacons
	}
	return nil
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 2: beacons (repeated message)
	for iNdEx := len(m.Beacons) - 1; iNdEx >= 0; iNdEx-- {
		{
			size, err := m.Beacons[iNdEx].MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	// field 1: params (message)
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTypes(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	l := m.Params.Size()
	n += 1 + l + sovTypes(uint64(l))
	for _, e := range m.Beacons {
		l = e.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func (m *GenesisState) Unmarshal(dAtA []byte) error {
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
		case 1: // params
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
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
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2: // beacons
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Beacons", wireType)
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
			m.Beacons = append(m.Beacons, RandomBeacon{})
			if err := m.Beacons[len(m.Beacons)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
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
	proto.RegisterFile("vrf/v1/vrf.proto", fileDescriptor_vrf_v1_vrf)
	proto.RegisterType((*RandomBeacon)(nil), "vrf.v1.RandomBeacon")
	proto.RegisterType((*Params)(nil), "vrf.v1.Params")
	proto.RegisterType((*GenesisState)(nil), "vrf.v1.GenesisState")
}

func (*RandomBeacon) Descriptor() ([]byte, []int) {
	return fileDescriptor_vrf_v1_vrf, []int{0}
}

func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_vrf_v1_vrf, []int{1}
}

func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_vrf_v1_vrf, []int{2}
}

// 273 bytes of a gzipped FileDescriptorProto for vrf/v1/vrf.proto
var fileDescriptor_vrf_v1_vrf = []byte{
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x4c, 0x50, 0xc1, 0x4a, 0xc3, 0x40,
	0x14, 0x24, 0xb6, 0xa4, 0xed, 0xb6, 0x54, 0x59, 0x44, 0x82, 0x78, 0xa8, 0x11, 0x34, 0xa7, 0x2c,
	0xad, 0x7f, 0x90, 0x8b, 0x07, 0x2f, 0xb2, 0xde, 0xbc, 0xc8, 0x6e, 0xfa, 0xd2, 0x84, 0x64, 0xf3,
	0x96, 0xdd, 0x35, 0xd4, 0xbf, 0x97, 0x6c, 0x12, 0xf5, 0xf6, 0x66, 0xde, 0x30, 0xc3, 0x0c, 0xb9,
	0xea, 0x4c, 0xc1, 0xba, 0x3d, 0xeb, 0x4c, 0x91, 0x6a, 0x83, 0x0e, 0x69, 0xd8, 0x9f, 0xdd, 0x3e,
	0x3e, 0x93, 0x0d, 0x17, 0xed, 0x11, 0x55, 0x06, 0x22, 0xc7, 0x96, 0xde, 0x90, 0xb0, 0x84, 0xea,
	0x54, 0xba, 0x28, 0xd8, 0x05, 0xc9, 0x9c, 0x8f, 0xa8, 0xe7, 0xa5, 0x57, 0x44, 0x17, 0xbb, 0x20,
	0x59, 0xf1, 0x11, 0xd1, 0x5b, 0xb2, 0xd4, 0x06, 0x35, 0x5a, 0x30, 0xd1, 0xcc, 0x7f, 0x7e, 0x31,
	0xbd, 0x23, 0x2b, 0x57, 0x29, 0xb0, 0x4e, 0x28, 0x1d, 0xcd, 0x77, 0x41, 0x32, 0xe3, 0x7f, 0x44,
	0xfc, 0x4a, 0xc2, 0x37, 0x61, 0x84, 0xb2, 0x34, 0x22, 0x0b, 0x68, 0x85, 0x6c, 0xe0, 0xe8, 0x43,
	0x97, 0x7c, 0x82, 0xf4, 0x89, 0x5c, 0x36, 0x88, 0xb5, 0x14, 0x79, 0xfd, 0x29, 0x1b, 0xcc, 0x6b,
	0xeb, 0xe3, 0xe7, 0x7c, 0x3b, 0xd1, 0x99, 0x67, 0xe3, 0x82, 0x6c, 0x5e, 0xa0, 0x05, 0x5b, 0xd9,
	0x77, 0x27, 0x1c, 0xd0, 0x47, 0x12, 0x6a, 0x6f, 0xee, 0x1d, 0xd7, 0x87, 0x6d, 0x3a, 0xf4, 0x4d,
	0x87, 0x48, 0x3e, 0x7e, 0x69, 0x4a, 0x16, 0x43, 0x91, 0xde, 0x78, 0x96, 0xac, 0x0f, 0xd7, 0x93,
	0xf0, 0xff, 0x2a, 0x7c, 0x12, 0x65, 0x0f, 0x1f, 0xf7, 0xa7, 0xca, 0x95, 0x5f, 0x32, 0xcd, 0x51,
	0xb1, 0x1c, 0xad, 0x42, 0xcb, 0xa0, 0x53, 0xec, 0xdc, 0x6f, 0xcb, 0xdc, 0xb7, 0x06, 0x2b, 0x43,
	0x3f, 0xf1, 0xf3, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x8c, 0x2e, 0x78, 0xe1, 0x76, 0x01, 0x00,
	0x00,
}
