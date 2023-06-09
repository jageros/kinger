// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: reborn.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type BuyRebornGoodsArg_GoodsType int32

const (
	BuyRebornGoodsArg_Unknow    BuyRebornGoodsArg_GoodsType = 0
	BuyRebornGoodsArg_Card      BuyRebornGoodsArg_GoodsType = 1
	BuyRebornGoodsArg_Privilege BuyRebornGoodsArg_GoodsType = 2
	BuyRebornGoodsArg_CardSkin  BuyRebornGoodsArg_GoodsType = 3
	BuyRebornGoodsArg_Equip     BuyRebornGoodsArg_GoodsType = 4
)

var BuyRebornGoodsArg_GoodsType_name = map[int32]string{
	0: "Unknow",
	1: "Card",
	2: "Privilege",
	3: "CardSkin",
	4: "Equip",
}
var BuyRebornGoodsArg_GoodsType_value = map[string]int32{
	"Unknow":    0,
	"Card":      1,
	"Privilege": 2,
	"CardSkin":  3,
	"Equip":     4,
}

func (x BuyRebornGoodsArg_GoodsType) String() string {
	return proto.EnumName(BuyRebornGoodsArg_GoodsType_name, int32(x))
}
func (BuyRebornGoodsArg_GoodsType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptorReborn, []int{3, 0}
}

type RefineCardArg struct {
	CardIDs []uint32 `protobuf:"varint,1,rep,packed,name=CardIDs" json:"CardIDs,omitempty"`
}

func (m *RefineCardArg) Reset()                    { *m = RefineCardArg{} }
func (m *RefineCardArg) String() string            { return proto.CompactTextString(m) }
func (*RefineCardArg) ProtoMessage()               {}
func (*RefineCardArg) Descriptor() ([]byte, []int) { return fileDescriptorReborn, []int{0} }

func (m *RefineCardArg) GetCardIDs() []uint32 {
	if m != nil {
		return m.CardIDs
	}
	return nil
}

type RefineCardReply struct {
	Reputation int32 `protobuf:"varint,1,opt,name=Reputation,proto3" json:"Reputation,omitempty"`
}

func (m *RefineCardReply) Reset()                    { *m = RefineCardReply{} }
func (m *RefineCardReply) String() string            { return proto.CompactTextString(m) }
func (*RefineCardReply) ProtoMessage()               {}
func (*RefineCardReply) Descriptor() ([]byte, []int) { return fileDescriptorReborn, []int{1} }

func (m *RefineCardReply) GetReputation() int32 {
	if m != nil {
		return m.Reputation
	}
	return 0
}

type RebornReply struct {
	TreasureReward *OpenTreasureReply `protobuf:"bytes,1,opt,name=TreasureReward" json:"TreasureReward,omitempty"`
	Reputation     int32              `protobuf:"varint,2,opt,name=Reputation,proto3" json:"Reputation,omitempty"`
	NewName        string             `protobuf:"bytes,3,opt,name=NewName,proto3" json:"NewName,omitempty"`
	Gold           int32              `protobuf:"varint,4,opt,name=Gold,proto3" json:"Gold,omitempty"`
}

func (m *RebornReply) Reset()                    { *m = RebornReply{} }
func (m *RebornReply) String() string            { return proto.CompactTextString(m) }
func (*RebornReply) ProtoMessage()               {}
func (*RebornReply) Descriptor() ([]byte, []int) { return fileDescriptorReborn, []int{2} }

func (m *RebornReply) GetTreasureReward() *OpenTreasureReply {
	if m != nil {
		return m.TreasureReward
	}
	return nil
}

func (m *RebornReply) GetReputation() int32 {
	if m != nil {
		return m.Reputation
	}
	return 0
}

func (m *RebornReply) GetNewName() string {
	if m != nil {
		return m.NewName
	}
	return ""
}

func (m *RebornReply) GetGold() int32 {
	if m != nil {
		return m.Gold
	}
	return 0
}

type BuyRebornGoodsArg struct {
	Type    BuyRebornGoodsArg_GoodsType `protobuf:"varint,1,opt,name=Type,proto3,enum=pb.BuyRebornGoodsArg_GoodsType" json:"Type,omitempty"`
	GoodsID int32                       `protobuf:"varint,2,opt,name=GoodsID,proto3" json:"GoodsID,omitempty"`
}

func (m *BuyRebornGoodsArg) Reset()                    { *m = BuyRebornGoodsArg{} }
func (m *BuyRebornGoodsArg) String() string            { return proto.CompactTextString(m) }
func (*BuyRebornGoodsArg) ProtoMessage()               {}
func (*BuyRebornGoodsArg) Descriptor() ([]byte, []int) { return fileDescriptorReborn, []int{3} }

func (m *BuyRebornGoodsArg) GetType() BuyRebornGoodsArg_GoodsType {
	if m != nil {
		return m.Type
	}
	return BuyRebornGoodsArg_Unknow
}

func (m *BuyRebornGoodsArg) GetGoodsID() int32 {
	if m != nil {
		return m.GoodsID
	}
	return 0
}

type RebornData struct {
	RemainDay int32 `protobuf:"varint,1,opt,name=RemainDay,proto3" json:"RemainDay,omitempty"`
	Prestige  int32 `protobuf:"varint,2,opt,name=Prestige,proto3" json:"Prestige,omitempty"`
	Cnt       int32 `protobuf:"varint,3,opt,name=Cnt,proto3" json:"Cnt,omitempty"`
}

func (m *RebornData) Reset()                    { *m = RebornData{} }
func (m *RebornData) String() string            { return proto.CompactTextString(m) }
func (*RebornData) ProtoMessage()               {}
func (*RebornData) Descriptor() ([]byte, []int) { return fileDescriptorReborn, []int{4} }

func (m *RebornData) GetRemainDay() int32 {
	if m != nil {
		return m.RemainDay
	}
	return 0
}

func (m *RebornData) GetPrestige() int32 {
	if m != nil {
		return m.Prestige
	}
	return 0
}

func (m *RebornData) GetCnt() int32 {
	if m != nil {
		return m.Cnt
	}
	return 0
}

func init() {
	proto.RegisterType((*RefineCardArg)(nil), "pb.RefineCardArg")
	proto.RegisterType((*RefineCardReply)(nil), "pb.RefineCardReply")
	proto.RegisterType((*RebornReply)(nil), "pb.RebornReply")
	proto.RegisterType((*BuyRebornGoodsArg)(nil), "pb.BuyRebornGoodsArg")
	proto.RegisterType((*RebornData)(nil), "pb.RebornData")
	proto.RegisterEnum("pb.BuyRebornGoodsArg_GoodsType", BuyRebornGoodsArg_GoodsType_name, BuyRebornGoodsArg_GoodsType_value)
}
func (m *RefineCardArg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RefineCardArg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.CardIDs) > 0 {
		dAtA2 := make([]byte, len(m.CardIDs)*10)
		var j1 int
		for _, num := range m.CardIDs {
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		dAtA[i] = 0xa
		i++
		i = encodeVarintReborn(dAtA, i, uint64(j1))
		i += copy(dAtA[i:], dAtA2[:j1])
	}
	return i, nil
}

func (m *RefineCardReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RefineCardReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Reputation != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.Reputation))
	}
	return i, nil
}

func (m *RebornReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RebornReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.TreasureReward != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.TreasureReward.Size()))
		n3, err := m.TreasureReward.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	if m.Reputation != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.Reputation))
	}
	if len(m.NewName) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintReborn(dAtA, i, uint64(len(m.NewName)))
		i += copy(dAtA[i:], m.NewName)
	}
	if m.Gold != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.Gold))
	}
	return i, nil
}

func (m *BuyRebornGoodsArg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BuyRebornGoodsArg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.Type))
	}
	if m.GoodsID != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.GoodsID))
	}
	return i, nil
}

func (m *RebornData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RebornData) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.RemainDay != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.RemainDay))
	}
	if m.Prestige != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.Prestige))
	}
	if m.Cnt != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintReborn(dAtA, i, uint64(m.Cnt))
	}
	return i, nil
}

func encodeVarintReborn(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *RefineCardArg) Size() (n int) {
	var l int
	_ = l
	if len(m.CardIDs) > 0 {
		l = 0
		for _, e := range m.CardIDs {
			l += sovReborn(uint64(e))
		}
		n += 1 + sovReborn(uint64(l)) + l
	}
	return n
}

func (m *RefineCardReply) Size() (n int) {
	var l int
	_ = l
	if m.Reputation != 0 {
		n += 1 + sovReborn(uint64(m.Reputation))
	}
	return n
}

func (m *RebornReply) Size() (n int) {
	var l int
	_ = l
	if m.TreasureReward != nil {
		l = m.TreasureReward.Size()
		n += 1 + l + sovReborn(uint64(l))
	}
	if m.Reputation != 0 {
		n += 1 + sovReborn(uint64(m.Reputation))
	}
	l = len(m.NewName)
	if l > 0 {
		n += 1 + l + sovReborn(uint64(l))
	}
	if m.Gold != 0 {
		n += 1 + sovReborn(uint64(m.Gold))
	}
	return n
}

func (m *BuyRebornGoodsArg) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovReborn(uint64(m.Type))
	}
	if m.GoodsID != 0 {
		n += 1 + sovReborn(uint64(m.GoodsID))
	}
	return n
}

func (m *RebornData) Size() (n int) {
	var l int
	_ = l
	if m.RemainDay != 0 {
		n += 1 + sovReborn(uint64(m.RemainDay))
	}
	if m.Prestige != 0 {
		n += 1 + sovReborn(uint64(m.Prestige))
	}
	if m.Cnt != 0 {
		n += 1 + sovReborn(uint64(m.Cnt))
	}
	return n
}

func sovReborn(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozReborn(x uint64) (n int) {
	return sovReborn(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *RefineCardArg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReborn
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
			return fmt.Errorf("proto: RefineCardArg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RefineCardArg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v uint32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowReborn
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= (uint32(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.CardIDs = append(m.CardIDs, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowReborn
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthReborn
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				for iNdEx < postIndex {
					var v uint32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowReborn
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= (uint32(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.CardIDs = append(m.CardIDs, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field CardIDs", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipReborn(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthReborn
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
func (m *RefineCardReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReborn
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
			return fmt.Errorf("proto: RefineCardReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RefineCardReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reputation", wireType)
			}
			m.Reputation = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Reputation |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipReborn(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthReborn
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
func (m *RebornReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReborn
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
			return fmt.Errorf("proto: RebornReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RebornReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreasureReward", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthReborn
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.TreasureReward == nil {
				m.TreasureReward = &OpenTreasureReply{}
			}
			if err := m.TreasureReward.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reputation", wireType)
			}
			m.Reputation = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Reputation |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NewName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthReborn
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NewName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gold", wireType)
			}
			m.Gold = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Gold |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipReborn(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthReborn
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
func (m *BuyRebornGoodsArg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReborn
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
			return fmt.Errorf("proto: BuyRebornGoodsArg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BuyRebornGoodsArg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (BuyRebornGoodsArg_GoodsType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GoodsID", wireType)
			}
			m.GoodsID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GoodsID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipReborn(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthReborn
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
func (m *RebornData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReborn
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
			return fmt.Errorf("proto: RebornData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RebornData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainDay", wireType)
			}
			m.RemainDay = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RemainDay |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Prestige", wireType)
			}
			m.Prestige = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Prestige |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cnt", wireType)
			}
			m.Cnt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReborn
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Cnt |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipReborn(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthReborn
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
func skipReborn(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowReborn
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
					return 0, ErrIntOverflowReborn
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
					return 0, ErrIntOverflowReborn
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
				return 0, ErrInvalidLengthReborn
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowReborn
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
				next, err := skipReborn(dAtA[start:])
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
	ErrInvalidLengthReborn = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowReborn   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("reborn.proto", fileDescriptorReborn) }

var fileDescriptorReborn = []byte{
	// 381 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x92, 0x41, 0x8e, 0xd3, 0x30,
	0x14, 0x86, 0xc7, 0x4d, 0x3a, 0x34, 0x6f, 0xa6, 0x25, 0x58, 0x42, 0x8a, 0x46, 0x28, 0x54, 0x59,
	0x95, 0x4d, 0x24, 0x66, 0xd6, 0x2c, 0x98, 0x06, 0x55, 0xdd, 0x94, 0xca, 0x14, 0x89, 0xad, 0xa3,
	0x3c, 0x2a, 0xab, 0xa9, 0x6d, 0x9c, 0x84, 0x2a, 0x37, 0x41, 0xdc, 0x81, 0x7b, 0xb0, 0xe4, 0x08,
	0xa8, 0x5c, 0x04, 0xd9, 0x4d, 0x5b, 0x54, 0x76, 0xef, 0xff, 0xfd, 0x5b, 0xff, 0xe7, 0x27, 0xc3,
	0xad, 0xc1, 0x5c, 0x19, 0x99, 0x6a, 0xa3, 0x6a, 0x45, 0x7b, 0x3a, 0xbf, 0x1b, 0xd5, 0x06, 0x79,
	0xd5, 0x18, 0x3c, 0x78, 0xc9, 0x2b, 0x18, 0x32, 0xfc, 0x2c, 0x24, 0x4e, 0xb9, 0x29, 0xde, 0x9a,
	0x35, 0x8d, 0xe0, 0x89, 0x1d, 0xe7, 0x59, 0x15, 0x91, 0xb1, 0x37, 0x19, 0xb2, 0xa3, 0x4c, 0x5e,
	0xc3, 0xd3, 0x73, 0x94, 0xa1, 0x2e, 0x5b, 0x1a, 0x03, 0x30, 0xd4, 0x4d, 0xcd, 0x6b, 0xa1, 0x64,
	0x44, 0xc6, 0x64, 0xd2, 0x67, 0xff, 0x38, 0xc9, 0x77, 0x02, 0x37, 0xcc, 0x21, 0x1c, 0xf2, 0x6f,
	0x60, 0xb4, 0xea, 0xfa, 0x19, 0xee, 0xb8, 0x29, 0xdc, 0x9d, 0x9b, 0xfb, 0xe7, 0xa9, 0xce, 0xd3,
	0xf7, 0x1a, 0xe5, 0xf9, 0x54, 0x97, 0x2d, 0xbb, 0x08, 0x5f, 0xd4, 0xf5, 0x2e, 0xeb, 0x2c, 0xfb,
	0x02, 0x77, 0x0b, 0xbe, 0xc5, 0xc8, 0x1b, 0x93, 0x49, 0xc0, 0x8e, 0x92, 0x52, 0xf0, 0x67, 0xaa,
	0x2c, 0x22, 0xdf, 0xdd, 0x71, 0x73, 0xf2, 0x83, 0xc0, 0xb3, 0xc7, 0xa6, 0x3d, 0xf0, 0xcd, 0x94,
	0x2a, 0x2a, 0xfb, 0xfe, 0x07, 0xf0, 0x57, 0xad, 0x46, 0x07, 0x36, 0xba, 0x7f, 0x69, 0xc1, 0xfe,
	0x0b, 0xa5, 0x6e, 0xb0, 0x31, 0xe6, 0xc2, 0xb6, 0xd8, 0x59, 0xf3, 0xac, 0xa3, 0x3a, 0xca, 0x64,
	0x0e, 0xc1, 0x29, 0x4c, 0x01, 0xae, 0x3f, 0xca, 0x8d, 0x54, 0xbb, 0xf0, 0x8a, 0x0e, 0xc0, 0xb7,
	0x7b, 0x0c, 0x09, 0x1d, 0x42, 0xb0, 0x34, 0xe2, 0xab, 0x28, 0x71, 0x8d, 0x61, 0x8f, 0xde, 0xc2,
	0xc0, 0x1e, 0x7c, 0xd8, 0x08, 0x19, 0x7a, 0x34, 0x80, 0xfe, 0xbb, 0x2f, 0x8d, 0xd0, 0xa1, 0x9f,
	0x7c, 0xb2, 0xaf, 0xb7, 0x18, 0x19, 0xaf, 0x39, 0x7d, 0x01, 0x01, 0xc3, 0x2d, 0x17, 0x32, 0xe3,
	0x6d, 0xb7, 0xf9, 0xb3, 0x41, 0xef, 0x60, 0xb0, 0x34, 0x58, 0xd5, 0x62, 0x8d, 0x1d, 0xd1, 0x49,
	0xd3, 0x10, 0xbc, 0xa9, 0xac, 0xdd, 0x86, 0xfa, 0xcc, 0x8e, 0x8f, 0xe1, 0xcf, 0x7d, 0x4c, 0x7e,
	0xed, 0x63, 0xf2, 0x7b, 0x1f, 0x93, 0x6f, 0x7f, 0xe2, 0xab, 0xfc, 0xda, 0xfd, 0x8e, 0x87, 0xbf,
	0x01, 0x00, 0x00, 0xff, 0xff, 0x72, 0xf5, 0x4c, 0x75, 0x41, 0x02, 0x00, 0x00,
}
