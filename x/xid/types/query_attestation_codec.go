package types

import (
	"fmt"
	"io"

	query "github.com/cosmos/cosmos-sdk/types/query"
)

// --- QueryStateDigestResponse ---

func (m *QueryStateDigestResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryStateDigestResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryStateDigestResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 3: num_names (uint64)
	if m.NumNames != 0 {
		i = encodeVarintQuery(dAtA, i, m.NumNames)
		i--
		dAtA[i] = 0x18
	}
	// field 2: height (uint64)
	if m.Height != 0 {
		i = encodeVarintQuery(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x10
	}
	// field 1: digest (string)
	if len(m.Digest) > 0 {
		i -= len(m.Digest)
		copy(dAtA[i:], m.Digest)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Digest)))
		i--
		dAtA[i] = 0x0a
	}
	return len(dAtA) - i, nil
}

func (m *QueryStateDigestResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Digest) > 0 {
		n += 1 + len(m.Digest) + sovQuery(uint64(len(m.Digest)))
	}
	if m.Height != 0 {
		n += 1 + sovQuery(m.Height)
	}
	if m.NumNames != 0 {
		n += 1 + sovQuery(m.NumNames)
	}
	return n
}

func (m *QueryStateDigestResponse) Unmarshal(dAtA []byte) error {
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
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return fmt.Errorf("proto: illegal negative length")
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	return nil
}

// --- QueryAttestationsRequest ---

func (m *QueryAttestationsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryAttestationsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryAttestationsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if len(m.Digest) > 0 {
		i -= len(m.Digest)
		copy(dAtA[i:], m.Digest)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Digest)))
		i--
		dAtA[i] = 0x0a
	}
	return len(dAtA) - i, nil
}

func (m *QueryAttestationsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Digest) > 0 {
		n += 1 + len(m.Digest) + sovQuery(uint64(len(m.Digest)))
	}
	return n
}

func (m *QueryAttestationsRequest) Unmarshal(dAtA []byte) error {
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
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return fmt.Errorf("proto: illegal negative length")
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	return nil
}

// --- QueryAttestationsResponse ---

func (m *QueryAttestationsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryAttestationsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryAttestationsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 2: finalized (bool)
	if m.Finalized {
		i--
		if m.Finalized {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x10
	}
	// field 1: attestations (repeated message)
	if len(m.Attestations) > 0 {
		for iNdEx := len(m.Attestations) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Attestations[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x0a
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryAttestationsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Attestations) > 0 {
		for _, e := range m.Attestations {
			l := e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if m.Finalized {
		n += 1 + 1
	}
	return n
}

func (m *QueryAttestationsResponse) Unmarshal(dAtA []byte) error {
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
		case 1: // attestations
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Attestations", wireType)
			}
			var msgLen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msgLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + msgLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Attestations = append(m.Attestations, Attestation{})
			if err := m.Attestations[len(m.Attestations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2: // finalized
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Finalized", wireType)
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
			m.Finalized = v != 0
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return fmt.Errorf("proto: illegal negative length")
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	return nil
}

// --- QueryStateSnapshotRequest ---

func (m *QueryStateSnapshotRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryStateSnapshotRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryStateSnapshotRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Pagination != nil {
		{
			size, err := m.Pagination.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x0a
	}
	return len(dAtA) - i, nil
}

func (m *QueryStateSnapshotRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	if m.Pagination != nil {
		l := m.Pagination.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryStateSnapshotRequest) Unmarshal(dAtA []byte) error {
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
		case 1: // pagination
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pagination", wireType)
			}
			var msgLen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msgLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + msgLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Pagination == nil {
				m.Pagination = &query.PageRequest{}
			}
			if err := m.Pagination.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
				return fmt.Errorf("proto: illegal negative length")
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	return nil
}

// --- DomainSnapshot ---

func (m *DomainSnapshot) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DomainSnapshot) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DomainSnapshot) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 5: content_root (string)
	if len(m.ContentRoot) > 0 {
		i -= len(m.ContentRoot)
		copy(dAtA[i:], m.ContentRoot)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ContentRoot)))
		i--
		dAtA[i] = 0x2a
	}
	// field 4: peers (repeated message)
	if len(m.Peers) > 0 {
		for iNdEx := len(m.Peers) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Peers[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	// field 3: dns_records (repeated message)
	if len(m.DnsRecords) > 0 {
		for iNdEx := len(m.DnsRecords) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DnsRecords[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	// field 2: profile (message, optional)
	if m.Profile != nil {
		{
			size, err := m.Profile.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	// field 1: record (message)
	{
		size, err := m.Record.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x0a
	return len(dAtA) - i, nil
}

func (m *DomainSnapshot) Size() (n int) {
	if m == nil {
		return 0
	}
	l := m.Record.Size()
	n += 1 + l + sovQuery(uint64(l))
	if m.Profile != nil {
		l = m.Profile.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	if len(m.DnsRecords) > 0 {
		for _, e := range m.DnsRecords {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if len(m.Peers) > 0 {
		for _, e := range m.Peers {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if len(m.ContentRoot) > 0 {
		n += 1 + len(m.ContentRoot) + sovQuery(uint64(len(m.ContentRoot)))
	}
	return n
}

func (m *DomainSnapshot) Unmarshal(dAtA []byte) error {
	// DomainSnapshot is only used in responses, unmarshal not needed for REST
	return nil
}

// --- QueryStateSnapshotResponse ---

func (m *QueryStateSnapshotResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryStateSnapshotResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryStateSnapshotResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	// field 4: pagination (message, optional)
	if m.Pagination != nil {
		{
			size, err := m.Pagination.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	// field 3: height (uint64)
	if m.Height != 0 {
		i = encodeVarintQuery(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x18
	}
	// field 2: digest (string)
	if len(m.Digest) > 0 {
		i -= len(m.Digest)
		copy(dAtA[i:], m.Digest)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Digest)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: domains (repeated message)
	if len(m.Domains) > 0 {
		for iNdEx := len(m.Domains) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Domains[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x0a
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryStateSnapshotResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	if len(m.Domains) > 0 {
		for _, e := range m.Domains {
			l := e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if len(m.Digest) > 0 {
		n += 1 + len(m.Digest) + sovQuery(uint64(len(m.Digest)))
	}
	if m.Height != 0 {
		n += 1 + sovQuery(m.Height)
	}
	if m.Pagination != nil {
		l := m.Pagination.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryStateSnapshotResponse) Unmarshal(dAtA []byte) error {
	// QueryStateSnapshotResponse is only used in responses, unmarshal not needed for REST
	return nil
}
