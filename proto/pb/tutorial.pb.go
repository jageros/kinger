// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tutorial.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type GetCampIDArg struct {
}

func (m *GetCampIDArg) Reset()                    { *m = GetCampIDArg{} }
func (m *GetCampIDArg) String() string            { return proto.CompactTextString(m) }
func (*GetCampIDArg) ProtoMessage()               {}
func (*GetCampIDArg) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{0} }

type GetCampIDReply struct {
	CampID int32 `protobuf:"varint,1,opt,name=CampID,proto3" json:"CampID,omitempty"`
}

func (m *GetCampIDReply) Reset()                    { *m = GetCampIDReply{} }
func (m *GetCampIDReply) String() string            { return proto.CompactTextString(m) }
func (*GetCampIDReply) ProtoMessage()               {}
func (*GetCampIDReply) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{1} }

func (m *GetCampIDReply) GetCampID() int32 {
	if m != nil {
		return m.CampID
	}
	return 0
}

type SetCampIDArg struct {
	CampID int32 `protobuf:"varint,1,opt,name=CampID,proto3" json:"CampID,omitempty"`
}

func (m *SetCampIDArg) Reset()                    { *m = SetCampIDArg{} }
func (m *SetCampIDArg) String() string            { return proto.CompactTextString(m) }
func (*SetCampIDArg) ProtoMessage()               {}
func (*SetCampIDArg) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{2} }

func (m *SetCampIDArg) GetCampID() int32 {
	if m != nil {
		return m.CampID
	}
	return 0
}

type SetCampIDReply struct {
	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (m *SetCampIDReply) Reset()                    { *m = SetCampIDReply{} }
func (m *SetCampIDReply) String() string            { return proto.CompactTextString(m) }
func (*SetCampIDReply) ProtoMessage()               {}
func (*SetCampIDReply) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{3} }

func (m *SetCampIDReply) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

type StartTutorialBattleArg struct {
	CampID int32 `protobuf:"varint,1,opt,name=CampID,proto3" json:"CampID,omitempty"`
}

func (m *StartTutorialBattleArg) Reset()                    { *m = StartTutorialBattleArg{} }
func (m *StartTutorialBattleArg) String() string            { return proto.CompactTextString(m) }
func (*StartTutorialBattleArg) ProtoMessage()               {}
func (*StartTutorialBattleArg) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{4} }

func (m *StartTutorialBattleArg) GetCampID() int32 {
	if m != nil {
		return m.CampID
	}
	return 0
}

type StartTutorialBattleReply struct {
	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (m *StartTutorialBattleReply) Reset()         { *m = StartTutorialBattleReply{} }
func (m *StartTutorialBattleReply) String() string { return proto.CompactTextString(m) }
func (*StartTutorialBattleReply) ProtoMessage()    {}
func (*StartTutorialBattleReply) Descriptor() ([]byte, []int) {
	return fileDescriptorTutorial, []int{5}
}

func (m *StartTutorialBattleReply) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

type TutorialFightEnd struct {
}

func (m *TutorialFightEnd) Reset()                    { *m = TutorialFightEnd{} }
func (m *TutorialFightEnd) String() string            { return proto.CompactTextString(m) }
func (*TutorialFightEnd) ProtoMessage()               {}
func (*TutorialFightEnd) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{6} }

type GuideBattle struct {
	Desk          *FightDesk `protobuf:"bytes,1,opt,name=Desk" json:"Desk,omitempty"`
	GuideBattleID int32      `protobuf:"varint,2,opt,name=GuideBattleID,proto3" json:"GuideBattleID,omitempty"`
}

func (m *GuideBattle) Reset()                    { *m = GuideBattle{} }
func (m *GuideBattle) String() string            { return proto.CompactTextString(m) }
func (*GuideBattle) ProtoMessage()               {}
func (*GuideBattle) Descriptor() ([]byte, []int) { return fileDescriptorTutorial, []int{7} }

func (m *GuideBattle) GetDesk() *FightDesk {
	if m != nil {
		return m.Desk
	}
	return nil
}

func (m *GuideBattle) GetGuideBattleID() int32 {
	if m != nil {
		return m.GuideBattleID
	}
	return 0
}

func init() {
	proto.RegisterType((*GetCampIDArg)(nil), "pb.GetCampIDArg")
	proto.RegisterType((*GetCampIDReply)(nil), "pb.GetCampIDReply")
	proto.RegisterType((*SetCampIDArg)(nil), "pb.SetCampIDArg")
	proto.RegisterType((*SetCampIDReply)(nil), "pb.SetCampIDReply")
	proto.RegisterType((*StartTutorialBattleArg)(nil), "pb.StartTutorialBattleArg")
	proto.RegisterType((*StartTutorialBattleReply)(nil), "pb.StartTutorialBattleReply")
	proto.RegisterType((*TutorialFightEnd)(nil), "pb.TutorialFightEnd")
	proto.RegisterType((*GuideBattle)(nil), "pb.GuideBattle")
}
func (m *GetCampIDArg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetCampIDArg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *GetCampIDReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetCampIDReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.CampID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintTutorial(dAtA, i, uint64(m.CampID))
	}
	return i, nil
}

func (m *SetCampIDArg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetCampIDArg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.CampID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintTutorial(dAtA, i, uint64(m.CampID))
	}
	return i, nil
}

func (m *SetCampIDReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetCampIDReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Ok {
		dAtA[i] = 0x8
		i++
		if m.Ok {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	return i, nil
}

func (m *StartTutorialBattleArg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StartTutorialBattleArg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.CampID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintTutorial(dAtA, i, uint64(m.CampID))
	}
	return i, nil
}

func (m *StartTutorialBattleReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StartTutorialBattleReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Ok {
		dAtA[i] = 0x8
		i++
		if m.Ok {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	return i, nil
}

func (m *TutorialFightEnd) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TutorialFightEnd) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *GuideBattle) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GuideBattle) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Desk != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTutorial(dAtA, i, uint64(m.Desk.Size()))
		n1, err := m.Desk.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.GuideBattleID != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintTutorial(dAtA, i, uint64(m.GuideBattleID))
	}
	return i, nil
}

func encodeVarintTutorial(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *GetCampIDArg) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *GetCampIDReply) Size() (n int) {
	var l int
	_ = l
	if m.CampID != 0 {
		n += 1 + sovTutorial(uint64(m.CampID))
	}
	return n
}

func (m *SetCampIDArg) Size() (n int) {
	var l int
	_ = l
	if m.CampID != 0 {
		n += 1 + sovTutorial(uint64(m.CampID))
	}
	return n
}

func (m *SetCampIDReply) Size() (n int) {
	var l int
	_ = l
	if m.Ok {
		n += 2
	}
	return n
}

func (m *StartTutorialBattleArg) Size() (n int) {
	var l int
	_ = l
	if m.CampID != 0 {
		n += 1 + sovTutorial(uint64(m.CampID))
	}
	return n
}

func (m *StartTutorialBattleReply) Size() (n int) {
	var l int
	_ = l
	if m.Ok {
		n += 2
	}
	return n
}

func (m *TutorialFightEnd) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *GuideBattle) Size() (n int) {
	var l int
	_ = l
	if m.Desk != nil {
		l = m.Desk.Size()
		n += 1 + l + sovTutorial(uint64(l))
	}
	if m.GuideBattleID != 0 {
		n += 1 + sovTutorial(uint64(m.GuideBattleID))
	}
	return n
}

func sovTutorial(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozTutorial(x uint64) (n int) {
	return sovTutorial(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GetCampIDArg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: GetCampIDArg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetCampIDArg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *GetCampIDReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: GetCampIDReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetCampIDReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CampID", wireType)
			}
			m.CampID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CampID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *SetCampIDArg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: SetCampIDArg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetCampIDArg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CampID", wireType)
			}
			m.CampID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CampID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *SetCampIDReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: SetCampIDReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetCampIDReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ok", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Ok = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *StartTutorialBattleArg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: StartTutorialBattleArg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StartTutorialBattleArg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CampID", wireType)
			}
			m.CampID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CampID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *StartTutorialBattleReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: StartTutorialBattleReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StartTutorialBattleReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ok", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Ok = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *TutorialFightEnd) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: TutorialFightEnd: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TutorialFightEnd: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func (m *GuideBattle) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTutorial
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
			return fmt.Errorf("proto: GuideBattle: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GuideBattle: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Desk", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
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
				return ErrInvalidLengthTutorial
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Desk == nil {
				m.Desk = &FightDesk{}
			}
			if err := m.Desk.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GuideBattleID", wireType)
			}
			m.GuideBattleID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTutorial
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GuideBattleID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTutorial(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTutorial
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
func skipTutorial(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTutorial
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
					return 0, ErrIntOverflowTutorial
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
					return 0, ErrIntOverflowTutorial
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
				return 0, ErrInvalidLengthTutorial
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowTutorial
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
				next, err := skipTutorial(dAtA[start:])
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
	ErrInvalidLengthTutorial = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTutorial   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("tutorial.proto", fileDescriptorTutorial) }

var fileDescriptorTutorial = []byte{
	// 234 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2b, 0x29, 0x2d, 0xc9,
	0x2f, 0xca, 0x4c, 0xcc, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x92, 0xe2,
	0x49, 0x4a, 0x2c, 0x29, 0xc9, 0x49, 0x85, 0x88, 0x28, 0xf1, 0x71, 0xf1, 0xb8, 0xa7, 0x96, 0x38,
	0x27, 0xe6, 0x16, 0x78, 0xba, 0x38, 0x16, 0xa5, 0x2b, 0x69, 0x70, 0xf1, 0xc1, 0xf9, 0x41, 0xa9,
	0x05, 0x39, 0x95, 0x42, 0x62, 0x5c, 0x6c, 0x10, 0xae, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x6b, 0x10,
	0x94, 0xa7, 0xa4, 0xc6, 0xc5, 0x13, 0x8c, 0xa4, 0x13, 0xa7, 0x3a, 0x05, 0x2e, 0xbe, 0x60, 0x54,
	0x13, 0xf9, 0xb8, 0x98, 0xf2, 0xb3, 0xc1, 0xaa, 0x38, 0x82, 0x98, 0xf2, 0xb3, 0x95, 0x0c, 0xb8,
	0xc4, 0x82, 0x4b, 0x12, 0x8b, 0x4a, 0x42, 0xa0, 0x8e, 0x75, 0x02, 0x3b, 0x10, 0x9f, 0x99, 0x5a,
	0x5c, 0x12, 0x58, 0x74, 0x60, 0x37, 0x5d, 0x88, 0x4b, 0x00, 0xa6, 0xcc, 0x2d, 0x33, 0x3d, 0xa3,
	0xc4, 0x35, 0x2f, 0x45, 0x29, 0x8c, 0x8b, 0xdb, 0xbd, 0x34, 0x33, 0x25, 0x15, 0xa2, 0x4f, 0x48,
	0x91, 0x8b, 0xc5, 0x25, 0xb5, 0x18, 0xa2, 0x89, 0xdb, 0x88, 0x57, 0xaf, 0x20, 0x49, 0x0f, 0xac,
	0x14, 0x24, 0x18, 0x04, 0x96, 0x12, 0x52, 0xe1, 0xe2, 0x45, 0xd2, 0xe1, 0xe9, 0x22, 0xc1, 0x04,
	0x76, 0x10, 0xaa, 0xa0, 0x93, 0xc0, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78,
	0x24, 0xc7, 0x38, 0xe3, 0xb1, 0x1c, 0x43, 0x12, 0x1b, 0x38, 0x98, 0x8d, 0x01, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x0c, 0x3a, 0x7a, 0x1f, 0x8a, 0x01, 0x00, 0x00,
}
