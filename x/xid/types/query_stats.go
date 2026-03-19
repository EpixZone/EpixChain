package types

import (
	"fmt"
	"io"

	proto "github.com/cosmos/gogoproto/proto"
)

// TLDStats contains statistics for a single TLD.
type TLDStats struct {
	Tld        string `protobuf:"bytes,1,opt,name=tld,proto3" json:"tld,omitempty"`
	NameCount  uint64 `protobuf:"varint,2,opt,name=name_count,json=nameCount,proto3" json:"name_count,omitempty"`
	FeesBurned string `protobuf:"bytes,3,opt,name=fees_burned,json=feesBurned,proto3" json:"fees_burned,omitempty"`
	Enabled    bool   `protobuf:"varint,4,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (m *TLDStats) Reset()         { *m = TLDStats{} }
func (m *TLDStats) String() string { return proto.CompactTextString(m) }
func (*TLDStats) ProtoMessage()    {}

func (m *TLDStats) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TLDStats) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TLDStats) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 4: enabled (bool)
	if m.Enabled {
		i--
		if m.Enabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x20
	}
	// field 3: fees_burned (string)
	if len(m.FeesBurned) > 0 {
		i -= len(m.FeesBurned)
		copy(dAtA[i:], m.FeesBurned)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.FeesBurned)))
		i--
		dAtA[i] = 0x1a
	}
	// field 2: name_count (uint64)
	if m.NameCount != 0 {
		i = encodeVarintQuery(dAtA, i, m.NameCount)
		i--
		dAtA[i] = 0x10
	}
	// field 1: tld (string)
	if len(m.Tld) > 0 {
		i -= len(m.Tld)
		copy(dAtA[i:], m.Tld)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Tld)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *TLDStats) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Tld) > 0 {
		n += 1 + len(m.Tld) + sovQuery(uint64(len(m.Tld)))
	}
	if m.NameCount != 0 {
		n += 1 + sovQuery(m.NameCount)
	}
	if len(m.FeesBurned) > 0 {
		n += 1 + len(m.FeesBurned) + sovQuery(uint64(len(m.FeesBurned)))
	}
	if m.Enabled {
		n += 1 + 1
	}
	return n
}

func (m *TLDStats) Unmarshal(dAtA []byte) error {
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
		case 1: // tld
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
		case 2: // name_count
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NameCount", wireType)
			}
			m.NameCount = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NameCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3: // fees_burned
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeesBurned", wireType)
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
			m.FeesBurned = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4: // enabled
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
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
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
// QueryGetStatsRequest — empty request
// ---------------------------------------------------------------------------

type QueryGetStatsRequest struct{}

func (m *QueryGetStatsRequest) Reset()         { *m = QueryGetStatsRequest{} }
func (m *QueryGetStatsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryGetStatsRequest) ProtoMessage()    {}

func (m *QueryGetStatsRequest) Marshal() (dAtA []byte, err error) {
	return nil, nil
}

func (m *QueryGetStatsRequest) MarshalTo(dAtA []byte) (int, error) {
	return 0, nil
}

func (m *QueryGetStatsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	return len(dAtA), nil
}

func (m *QueryGetStatsRequest) Size() int {
	return 0
}

func (m *QueryGetStatsRequest) Unmarshal(dAtA []byte) error {
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
		_ = fieldNum
		iNdEx = preIndex
		skippy, err := skipQuery(dAtA[iNdEx:])
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
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// ---------------------------------------------------------------------------
// QueryGetStatsResponse
// ---------------------------------------------------------------------------

type QueryGetStatsResponse struct {
	TotalNames      uint64     `protobuf:"varint,1,opt,name=total_names,json=totalNames,proto3" json:"total_names,omitempty"`
	TotalFeesBurned string     `protobuf:"bytes,2,opt,name=total_fees_burned,json=totalFeesBurned,proto3" json:"total_fees_burned,omitempty"`
	TldStats        []TLDStats `protobuf:"bytes,3,rep,name=tld_stats,json=tldStats,proto3" json:"tld_stats"`
}

func (m *QueryGetStatsResponse) Reset()         { *m = QueryGetStatsResponse{} }
func (m *QueryGetStatsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryGetStatsResponse) ProtoMessage()    {}

func (m *QueryGetStatsResponse) GetTotalNames() uint64 {
	if m != nil {
		return m.TotalNames
	}
	return 0
}

func (m *QueryGetStatsResponse) GetTotalFeesBurned() string {
	if m != nil {
		return m.TotalFeesBurned
	}
	return ""
}

func (m *QueryGetStatsResponse) GetTldStats() []TLDStats {
	if m != nil {
		return m.TldStats
	}
	return nil
}

func (m *QueryGetStatsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryGetStatsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryGetStatsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 3: tld_stats (repeated message)
	if len(m.TldStats) > 0 {
		for iNdEx := len(m.TldStats) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TldStats[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	// field 2: total_fees_burned (string)
	if len(m.TotalFeesBurned) > 0 {
		i -= len(m.TotalFeesBurned)
		copy(dAtA[i:], m.TotalFeesBurned)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.TotalFeesBurned)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: total_names (uint64)
	if m.TotalNames != 0 {
		i = encodeVarintQuery(dAtA, i, m.TotalNames)
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryGetStatsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	if m.TotalNames != 0 {
		n += 1 + sovQuery(m.TotalNames)
	}
	if len(m.TotalFeesBurned) > 0 {
		n += 1 + len(m.TotalFeesBurned) + sovQuery(uint64(len(m.TotalFeesBurned)))
	}
	if len(m.TldStats) > 0 {
		for _, e := range m.TldStats {
			l := e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryGetStatsResponse) Unmarshal(dAtA []byte) error {
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
		case 1: // total_names
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalNames", wireType)
			}
			m.TotalNames = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalNames |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2: // total_fees_burned
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalFeesBurned", wireType)
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
			m.TotalFeesBurned = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3: // tld_stats
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TldStats", wireType)
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
			m.TldStats = append(m.TldStats, TLDStats{})
			if err := m.TldStats[len(m.TldStats)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
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

func init() {
	proto.RegisterType((*TLDStats)(nil), "xid.v1.TLDStats")
	proto.RegisterType((*QueryGetStatsRequest)(nil), "xid.v1.QueryGetStatsRequest")
	proto.RegisterType((*QueryGetStatsResponse)(nil), "xid.v1.QueryGetStatsResponse")
}
