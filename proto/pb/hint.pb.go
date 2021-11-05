// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: hint.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type HintType int32

const (
	HintType_HtUnknow        HintType = 0
	HintType_HtLimitGift     HintType = 1
	HintType_HtShopTreasure  HintType = 2
	HintType_HtFreeAds       HintType = 3
	HintType_HtGaCha         HintType = 4
	HintType_HtFirstRecharge HintType = 5
	HintType_HtRandomFree    HintType = 6
	HintType_HtRecruitReward HintType = 500
	HintType_HtActivity      HintType = 1000
)

var HintType_name = map[int32]string{
	0:    "HtUnknow",
	1:    "HtLimitGift",
	2:    "HtShopTreasure",
	3:    "HtFreeAds",
	4:    "HtGaCha",
	5:    "HtFirstRecharge",
	6:    "HtRandomFree",
	500:  "HtRecruitReward",
	1000: "HtActivity",
}
var HintType_value = map[string]int32{
	"HtUnknow":        0,
	"HtLimitGift":     1,
	"HtShopTreasure":  2,
	"HtFreeAds":       3,
	"HtGaCha":         4,
	"HtFirstRecharge": 5,
	"HtRandomFree":    6,
	"HtRecruitReward": 500,
	"HtActivity":      1000,
}

func (x HintType) String() string {
	return proto.EnumName(HintType_name, int32(x))
}
func (HintType) EnumDescriptor() ([]byte, []int) { return fileDescriptorHint, []int{0} }

type Hint struct {
	Type  HintType `protobuf:"varint,1,opt,name=Type,proto3,enum=pb.HintType" json:"Type,omitempty"`
	Count int32    `protobuf:"varint,2,opt,name=Count,proto3" json:"Count,omitempty"`
}

func (m *Hint) Reset()                    { *m = Hint{} }
func (m *Hint) String() string            { return proto.CompactTextString(m) }
func (*Hint) ProtoMessage()               {}
func (*Hint) Descriptor() ([]byte, []int) { return fileDescriptorHint, []int{0} }

func (m *Hint) GetType() HintType {
	if m != nil {
		return m.Type
	}
	return HintType_HtUnknow
}

func (m *Hint) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func init() {
	proto.RegisterType((*Hint)(nil), "pb.Hint")
	proto.RegisterEnum("pb.HintType", HintType_name, HintType_value)
}
func (m *Hint) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Hint) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintHint(dAtA, i, uint64(m.Type))
	}
	if m.Count != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintHint(dAtA, i, uint64(m.Count))
	}
	return i, nil
}

func encodeVarintHint(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Hint) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovHint(uint64(m.Type))
	}
	if m.Count != 0 {
		n += 1 + sovHint(uint64(m.Count))
	}
	return n
}

func sovHint(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozHint(x uint64) (n int) {
	return sovHint(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Hint) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHint
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Hint: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Hint: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (HintType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Count", wireType)
			}
			m.Count = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Count |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipHint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthHint
			}
			if (iNdEx + skippy) > l {
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
func skipHint(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowHint
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHint
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHint
				}
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthHint
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowHint
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipHint(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthHint = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowHint   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("hint.proto", fileDescriptorHint) }

var fileDescriptorHint = []byte{
	// 268 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x90, 0x31, 0x4e, 0xf3, 0x40,
	0x14, 0x84, 0xb3, 0x49, 0x9c, 0xe4, 0x7f, 0xf1, 0x1f, 0xaf, 0x1e, 0x29, 0x52, 0x59, 0x16, 0x55,
	0x44, 0xe1, 0x02, 0x7a, 0xa4, 0x10, 0x29, 0xd9, 0x82, 0x6a, 0x09, 0x07, 0x58, 0xdb, 0x0b, 0x5e,
	0xa1, 0xec, 0x5a, 0xeb, 0x67, 0xa2, 0xdc, 0x84, 0x1b, 0x70, 0x15, 0x4a, 0x8e, 0x80, 0x4c, 0xc3,
	0x01, 0x38, 0x00, 0x72, 0x10, 0xe5, 0x7c, 0xf3, 0x4d, 0x33, 0x00, 0xa5, 0xb1, 0x94, 0x56, 0xde,
	0x91, 0xc3, 0x7e, 0x95, 0x9d, 0x5f, 0xc3, 0x50, 0x18, 0x4b, 0x98, 0xc0, 0x70, 0x77, 0xac, 0xf4,
	0x82, 0x25, 0x6c, 0x39, 0xbb, 0x0c, 0xd3, 0x2a, 0x4b, 0x3b, 0xde, 0x31, 0x79, 0x6a, 0x70, 0x0e,
	0xc1, 0xda, 0x35, 0x96, 0x16, 0xfd, 0x84, 0x2d, 0x03, 0xf9, 0x1b, 0x2e, 0x5e, 0x19, 0x4c, 0xfe,
	0x44, 0x0c, 0x61, 0x22, 0xe8, 0xde, 0x3e, 0x59, 0x77, 0xe0, 0x3d, 0x8c, 0x60, 0x2a, 0xe8, 0xd6,
	0xec, 0x0d, 0x6d, 0xcd, 0x03, 0x71, 0x86, 0x08, 0x33, 0x41, 0x77, 0xa5, 0xab, 0x76, 0x5e, 0xab,
	0xba, 0xf1, 0x9a, 0xf7, 0xf1, 0x3f, 0xfc, 0x13, 0xb4, 0xf1, 0x5a, 0xaf, 0x8a, 0x9a, 0x0f, 0x70,
	0x0a, 0x63, 0x41, 0x5b, 0xb5, 0x2e, 0x15, 0x1f, 0xe2, 0x19, 0x44, 0x82, 0x36, 0xc6, 0xd7, 0x24,
	0x75, 0x5e, 0x2a, 0xff, 0xa8, 0x79, 0x80, 0x1c, 0x42, 0x41, 0x52, 0xd9, 0xc2, 0xed, 0xbb, 0x19,
	0x1f, 0xe1, 0xbc, 0xd3, 0xa4, 0xce, 0x7d, 0x63, 0x48, 0xea, 0x83, 0xf2, 0x05, 0xff, 0x1e, 0x60,
	0x04, 0x20, 0x68, 0x95, 0x93, 0x79, 0x36, 0x74, 0xe4, 0x5f, 0xe3, 0x1b, 0xfe, 0xd6, 0xc6, 0xec,
	0xbd, 0x8d, 0xd9, 0x47, 0x1b, 0xb3, 0x97, 0xcf, 0xb8, 0x97, 0x8d, 0x4e, 0x37, 0x5c, 0xfd, 0x04,
	0x00, 0x00, 0xff, 0xff, 0x69, 0x21, 0x75, 0x4a, 0x14, 0x01, 0x00, 0x00,
}
