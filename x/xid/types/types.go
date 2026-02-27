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
	DNSRecordTypeSRV      uint32 = 33
	DNSRecordTypeEpixNet  uint32 = 65280 // Private-use range (RFC 6895) for EpixNet peer discovery
)

// ContentRoot represents the auto-computed Merkle root of active peers for an xID
type ContentRoot struct {
	Root      string `protobuf:"bytes,1,opt,name=root,proto3" json:"root,omitempty"`
	UpdatedAt uint64 `protobuf:"varint,2,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (m *ContentRoot) Reset()         { *m = ContentRoot{} }
func (m *ContentRoot) String() string { return proto.CompactTextString(m) }
func (*ContentRoot) ProtoMessage()    {}
func (*ContentRoot) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bd3daa00c1847cd, []int{6}
}
func (m *ContentRoot) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ContentRoot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ContentRoot.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ContentRoot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContentRoot.Merge(m, src)
}
func (m *ContentRoot) XXX_Size() int {
	return m.Size()
}
func (m *ContentRoot) XXX_DiscardUnknown() {
	xxx_messageInfo_ContentRoot.DiscardUnknown(m)
}

var xxx_messageInfo_ContentRoot proto.InternalMessageInfo

func (m *ContentRoot) GetRoot() string {
	if m != nil {
		return m.Root
	}
	return ""
}

func (m *ContentRoot) GetUpdatedAt() uint64 {
	if m != nil {
		return m.UpdatedAt
	}
	return 0
}

func (m *ContentRoot) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ContentRoot) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ContentRoot) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 2: updated_at (uint64, varint)
	if m.UpdatedAt != 0 {
		i = encodeVarintTypes(dAtA, i, m.UpdatedAt)
		i--
		dAtA[i] = 0x10
	}
	// field 1: root (string)
	if len(m.Root) > 0 {
		i -= len(m.Root)
		copy(dAtA[i:], m.Root)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Root)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ContentRoot) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Root) > 0 {
		n += 1 + len(m.Root) + sovTypes(uint64(len(m.Root)))
	}
	if m.UpdatedAt != 0 {
		n += 1 + sovTypes(m.UpdatedAt)
	}
	return n
}

func (m *ContentRoot) Unmarshal(dAtA []byte) error {
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
		case 1: // root
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Root", wireType)
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
			m.Root = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // updated_at
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdatedAt", wireType)
			}
			m.UpdatedAt = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdatedAt |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3: // submitter
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Submitter", wireType)
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
			// Skip legacy submitter field (backward compat)
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

// EpixNetPeer represents an EpixNet peer address attached to a name
type EpixNetPeer struct {
	Address   string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Label     string `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	AddedAt   uint64 `protobuf:"varint,3,opt,name=added_at,json=addedAt,proto3" json:"added_at,omitempty"`
	Active    bool   `protobuf:"varint,4,opt,name=active,proto3" json:"active,omitempty"`
	RevokedAt uint64 `protobuf:"varint,5,opt,name=revoked_at,json=revokedAt,proto3" json:"revoked_at,omitempty"`
}

func (m *EpixNetPeer) Reset()         { *m = EpixNetPeer{} }
func (m *EpixNetPeer) String() string { return proto.CompactTextString(m) }
func (*EpixNetPeer) ProtoMessage()    {}
func (*EpixNetPeer) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bd3daa00c1847cd, []int{7}
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

func (m *EpixNetPeer) GetActive() bool {
	if m != nil {
		return m.Active
	}
	return false
}

func (m *EpixNetPeer) GetRevokedAt() uint64 {
	if m != nil {
		return m.RevokedAt
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
	// field 5: revoked_at (uint64, varint)
	if m.RevokedAt != 0 {
		i = encodeVarintTypes(dAtA, i, m.RevokedAt)
		i--
		dAtA[i] = 0x28
	}
	// field 4: active (bool, varint)
	if m.Active {
		i--
		dAtA[i] = 1
		i--
		dAtA[i] = 0x20
	}
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
	if m.Active {
		n += 2
	}
	if m.RevokedAt != 0 {
		n += 1 + sovTypes(m.RevokedAt)
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
		case 4: // active
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Active", wireType)
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
			m.Active = v != 0
		case 5: // revoked_at
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RevokedAt", wireType)
			}
			m.RevokedAt = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RevokedAt |= uint64(b&0x7F) << shift
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

// StateDigest represents a deterministic hash of all xID state
type StateDigest struct {
	Digest   string `protobuf:"bytes,1,opt,name=digest,proto3" json:"digest,omitempty"`
	Height   uint64 `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	NumNames uint64 `protobuf:"varint,3,opt,name=num_names,json=numNames,proto3" json:"num_names,omitempty"`
}

func (m *StateDigest) Reset()         { *m = StateDigest{} }
func (m *StateDigest) String() string { return proto.CompactTextString(m) }
func (*StateDigest) ProtoMessage()    {}
func (*StateDigest) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bd3daa00c1847cd, []int{8}
}
func (m *StateDigest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StateDigest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StateDigest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StateDigest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateDigest.Merge(m, src)
}
func (m *StateDigest) XXX_Size() int {
	return m.Size()
}
func (m *StateDigest) XXX_DiscardUnknown() {
	xxx_messageInfo_StateDigest.DiscardUnknown(m)
}

var xxx_messageInfo_StateDigest proto.InternalMessageInfo

func (m *StateDigest) GetDigest() string {
	if m != nil {
		return m.Digest
	}
	return ""
}

func (m *StateDigest) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *StateDigest) GetNumNames() uint64 {
	if m != nil {
		return m.NumNames
	}
	return 0
}

func (m *StateDigest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StateDigest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StateDigest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.NumNames != 0 {
		i = encodeVarintTypes(dAtA, i, m.NumNames)
		i--
		dAtA[i] = 0x18
	}
	if m.Height != 0 {
		i = encodeVarintTypes(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x10
	}
	if len(m.Digest) > 0 {
		i -= len(m.Digest)
		copy(dAtA[i:], m.Digest)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Digest)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *StateDigest) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Digest) > 0 {
		n += 1 + len(m.Digest) + sovTypes(uint64(len(m.Digest)))
	}
	if m.Height != 0 {
		n += 1 + sovTypes(m.Height)
	}
	if m.NumNames != 0 {
		n += 1 + sovTypes(m.NumNames)
	}
	return n
}

func (m *StateDigest) Unmarshal(dAtA []byte) error {
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
		case 1: // digest
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Digest", wireType)
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
			m.Digest = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // height
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
		case 3: // num_names
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumNames", wireType)
			}
			m.NumNames = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumNames |= uint64(b&0x7F) << shift
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

// Attestation represents a validator's attestation of a state digest
type Attestation struct {
	ValidatorAddr string `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
	Digest        string `protobuf:"bytes,2,opt,name=digest,proto3" json:"digest,omitempty"`
	Signature     string `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	Height        uint64 `protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *Attestation) Reset()         { *m = Attestation{} }
func (m *Attestation) String() string { return proto.CompactTextString(m) }
func (*Attestation) ProtoMessage()    {}
func (*Attestation) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bd3daa00c1847cd, []int{9}
}
func (m *Attestation) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Attestation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Attestation.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Attestation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Attestation.Merge(m, src)
}
func (m *Attestation) XXX_Size() int {
	return m.Size()
}
func (m *Attestation) XXX_DiscardUnknown() {
	xxx_messageInfo_Attestation.DiscardUnknown(m)
}

var xxx_messageInfo_Attestation proto.InternalMessageInfo

func (m *Attestation) GetValidatorAddr() string {
	if m != nil {
		return m.ValidatorAddr
	}
	return ""
}

func (m *Attestation) GetDigest() string {
	if m != nil {
		return m.Digest
	}
	return ""
}

func (m *Attestation) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

func (m *Attestation) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Attestation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Attestation) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Attestation) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Height != 0 {
		i = encodeVarintTypes(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x20
	}
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Digest) > 0 {
		i -= len(m.Digest)
		copy(dAtA[i:], m.Digest)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Digest)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ValidatorAddr) > 0 {
		i -= len(m.ValidatorAddr)
		copy(dAtA[i:], m.ValidatorAddr)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.ValidatorAddr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Attestation) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.ValidatorAddr) > 0 {
		n += 1 + len(m.ValidatorAddr) + sovTypes(uint64(len(m.ValidatorAddr)))
	}
	if len(m.Digest) > 0 {
		n += 1 + len(m.Digest) + sovTypes(uint64(len(m.Digest)))
	}
	if len(m.Signature) > 0 {
		n += 1 + len(m.Signature) + sovTypes(uint64(len(m.Signature)))
	}
	if m.Height != 0 {
		n += 1 + sovTypes(m.Height)
	}
	return n
}

func (m *Attestation) Unmarshal(dAtA []byte) error {
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
		case 1: // validator_addr
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddr", wireType)
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
			m.ValidatorAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // digest
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Digest", wireType)
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
			m.Digest = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3: // signature
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
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
			m.Signature = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4: // height
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

func init() {
	proto.RegisterType((*ContentRoot)(nil), "xid.v1.ContentRoot")
	proto.RegisterType((*EpixNetPeer)(nil), "xid.v1.EpixNetPeer")
	proto.RegisterType((*StateDigest)(nil), "xid.v1.StateDigest")
	proto.RegisterType((*Attestation)(nil), "xid.v1.Attestation")
}
