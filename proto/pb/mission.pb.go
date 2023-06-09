// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: mission.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Mission struct {
	MissionID int32 `protobuf:"varint,1,opt,name=MissionID,proto3" json:"MissionID,omitempty"`
	CurCnt    int32 `protobuf:"varint,2,opt,name=CurCnt,proto3" json:"CurCnt,omitempty"`
	IsReward  bool  `protobuf:"varint,3,opt,name=IsReward,proto3" json:"IsReward,omitempty"`
	ID        int32 `protobuf:"varint,4,opt,name=ID,proto3" json:"ID,omitempty"`
}

func (m *Mission) Reset()                    { *m = Mission{} }
func (m *Mission) String() string            { return proto.CompactTextString(m) }
func (*Mission) ProtoMessage()               {}
func (*Mission) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{0} }

func (m *Mission) GetMissionID() int32 {
	if m != nil {
		return m.MissionID
	}
	return 0
}

func (m *Mission) GetCurCnt() int32 {
	if m != nil {
		return m.CurCnt
	}
	return 0
}

func (m *Mission) GetIsReward() bool {
	if m != nil {
		return m.IsReward
	}
	return false
}

func (m *Mission) GetID() int32 {
	if m != nil {
		return m.ID
	}
	return 0
}

type MissionTreasure struct {
	TreasureModelID string `protobuf:"bytes,1,opt,name=TreasureModelID,proto3" json:"TreasureModelID,omitempty"`
	CurCnt          int32  `protobuf:"varint,2,opt,name=CurCnt,proto3" json:"CurCnt,omitempty"`
}

func (m *MissionTreasure) Reset()                    { *m = MissionTreasure{} }
func (m *MissionTreasure) String() string            { return proto.CompactTextString(m) }
func (*MissionTreasure) ProtoMessage()               {}
func (*MissionTreasure) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{1} }

func (m *MissionTreasure) GetTreasureModelID() string {
	if m != nil {
		return m.TreasureModelID
	}
	return ""
}

func (m *MissionTreasure) GetCurCnt() int32 {
	if m != nil {
		return m.CurCnt
	}
	return 0
}

type MissionInfo struct {
	Missions          []*Mission       `protobuf:"bytes,1,rep,name=Missions" json:"Missions,omitempty"`
	Treasure          *MissionTreasure `protobuf:"bytes,2,opt,name=Treasure" json:"Treasure,omitempty"`
	CanRefresh        bool             `protobuf:"varint,3,opt,name=CanRefresh,proto3" json:"CanRefresh,omitempty"`
	RefreshRemainTime int32            `protobuf:"varint,4,opt,name=RefreshRemainTime,proto3" json:"RefreshRemainTime,omitempty"`
}

func (m *MissionInfo) Reset()                    { *m = MissionInfo{} }
func (m *MissionInfo) String() string            { return proto.CompactTextString(m) }
func (*MissionInfo) ProtoMessage()               {}
func (*MissionInfo) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{2} }

func (m *MissionInfo) GetMissions() []*Mission {
	if m != nil {
		return m.Missions
	}
	return nil
}

func (m *MissionInfo) GetTreasure() *MissionTreasure {
	if m != nil {
		return m.Treasure
	}
	return nil
}

func (m *MissionInfo) GetCanRefresh() bool {
	if m != nil {
		return m.CanRefresh
	}
	return false
}

func (m *MissionInfo) GetRefreshRemainTime() int32 {
	if m != nil {
		return m.RefreshRemainTime
	}
	return 0
}

type TargetMission struct {
	ID int32 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
}

func (m *TargetMission) Reset()                    { *m = TargetMission{} }
func (m *TargetMission) String() string            { return proto.CompactTextString(m) }
func (*TargetMission) ProtoMessage()               {}
func (*TargetMission) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{3} }

func (m *TargetMission) GetID() int32 {
	if m != nil {
		return m.ID
	}
	return 0
}

type MissionReward struct {
	Jade        int32    `protobuf:"varint,1,opt,name=Jade,proto3" json:"Jade,omitempty"`
	Gold        int32    `protobuf:"varint,2,opt,name=Gold,proto3" json:"Gold,omitempty"`
	Bowlder     int32    `protobuf:"varint,3,opt,name=Bowlder,proto3" json:"Bowlder,omitempty"`
	NextMission *Mission `protobuf:"bytes,4,opt,name=NextMission" json:"NextMission,omitempty"`
}

func (m *MissionReward) Reset()                    { *m = MissionReward{} }
func (m *MissionReward) String() string            { return proto.CompactTextString(m) }
func (*MissionReward) ProtoMessage()               {}
func (*MissionReward) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{4} }

func (m *MissionReward) GetJade() int32 {
	if m != nil {
		return m.Jade
	}
	return 0
}

func (m *MissionReward) GetGold() int32 {
	if m != nil {
		return m.Gold
	}
	return 0
}

func (m *MissionReward) GetBowlder() int32 {
	if m != nil {
		return m.Bowlder
	}
	return 0
}

func (m *MissionReward) GetNextMission() *Mission {
	if m != nil {
		return m.NextMission
	}
	return nil
}

type OpenMissionTreasureReply struct {
	TreasureReward *OpenTreasureReply `protobuf:"bytes,1,opt,name=TreasureReward" json:"TreasureReward,omitempty"`
	NextTreasure   *MissionTreasure   `protobuf:"bytes,2,opt,name=NextTreasure" json:"NextTreasure,omitempty"`
}

func (m *OpenMissionTreasureReply) Reset()                    { *m = OpenMissionTreasureReply{} }
func (m *OpenMissionTreasureReply) String() string            { return proto.CompactTextString(m) }
func (*OpenMissionTreasureReply) ProtoMessage()               {}
func (*OpenMissionTreasureReply) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{5} }

func (m *OpenMissionTreasureReply) GetTreasureReward() *OpenTreasureReply {
	if m != nil {
		return m.TreasureReward
	}
	return nil
}

func (m *OpenMissionTreasureReply) GetNextTreasure() *MissionTreasure {
	if m != nil {
		return m.NextTreasure
	}
	return nil
}

type UpdateMissionProcessArg struct {
	Missions []*Mission `protobuf:"bytes,1,rep,name=Missions" json:"Missions,omitempty"`
}

func (m *UpdateMissionProcessArg) Reset()                    { *m = UpdateMissionProcessArg{} }
func (m *UpdateMissionProcessArg) String() string            { return proto.CompactTextString(m) }
func (*UpdateMissionProcessArg) ProtoMessage()               {}
func (*UpdateMissionProcessArg) Descriptor() ([]byte, []int) { return fileDescriptorMission, []int{6} }

func (m *UpdateMissionProcessArg) GetMissions() []*Mission {
	if m != nil {
		return m.Missions
	}
	return nil
}

func init() {
	proto.RegisterType((*Mission)(nil), "pb.Mission")
	proto.RegisterType((*MissionTreasure)(nil), "pb.MissionTreasure")
	proto.RegisterType((*MissionInfo)(nil), "pb.MissionInfo")
	proto.RegisterType((*TargetMission)(nil), "pb.TargetMission")
	proto.RegisterType((*MissionReward)(nil), "pb.MissionReward")
	proto.RegisterType((*OpenMissionTreasureReply)(nil), "pb.OpenMissionTreasureReply")
	proto.RegisterType((*UpdateMissionProcessArg)(nil), "pb.UpdateMissionProcessArg")
}
func (m *Mission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Mission) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.MissionID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.MissionID))
	}
	if m.CurCnt != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.CurCnt))
	}
	if m.IsReward {
		dAtA[i] = 0x18
		i++
		if m.IsReward {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.ID != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.ID))
	}
	return i, nil
}

func (m *MissionTreasure) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MissionTreasure) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.TreasureModelID) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMission(dAtA, i, uint64(len(m.TreasureModelID)))
		i += copy(dAtA[i:], m.TreasureModelID)
	}
	if m.CurCnt != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.CurCnt))
	}
	return i, nil
}

func (m *MissionInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MissionInfo) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Missions) > 0 {
		for _, msg := range m.Missions {
			dAtA[i] = 0xa
			i++
			i = encodeVarintMission(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.Treasure != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.Treasure.Size()))
		n1, err := m.Treasure.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.CanRefresh {
		dAtA[i] = 0x18
		i++
		if m.CanRefresh {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.RefreshRemainTime != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.RefreshRemainTime))
	}
	return i, nil
}

func (m *TargetMission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TargetMission) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.ID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.ID))
	}
	return i, nil
}

func (m *MissionReward) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MissionReward) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Jade != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.Jade))
	}
	if m.Gold != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.Gold))
	}
	if m.Bowlder != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.Bowlder))
	}
	if m.NextMission != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.NextMission.Size()))
		n2, err := m.NextMission.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *OpenMissionTreasureReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OpenMissionTreasureReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.TreasureReward != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.TreasureReward.Size()))
		n3, err := m.TreasureReward.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	if m.NextTreasure != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMission(dAtA, i, uint64(m.NextTreasure.Size()))
		n4, err := m.NextTreasure.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n4
	}
	return i, nil
}

func (m *UpdateMissionProcessArg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UpdateMissionProcessArg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Missions) > 0 {
		for _, msg := range m.Missions {
			dAtA[i] = 0xa
			i++
			i = encodeVarintMission(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func encodeVarintMission(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Mission) Size() (n int) {
	var l int
	_ = l
	if m.MissionID != 0 {
		n += 1 + sovMission(uint64(m.MissionID))
	}
	if m.CurCnt != 0 {
		n += 1 + sovMission(uint64(m.CurCnt))
	}
	if m.IsReward {
		n += 2
	}
	if m.ID != 0 {
		n += 1 + sovMission(uint64(m.ID))
	}
	return n
}

func (m *MissionTreasure) Size() (n int) {
	var l int
	_ = l
	l = len(m.TreasureModelID)
	if l > 0 {
		n += 1 + l + sovMission(uint64(l))
	}
	if m.CurCnt != 0 {
		n += 1 + sovMission(uint64(m.CurCnt))
	}
	return n
}

func (m *MissionInfo) Size() (n int) {
	var l int
	_ = l
	if len(m.Missions) > 0 {
		for _, e := range m.Missions {
			l = e.Size()
			n += 1 + l + sovMission(uint64(l))
		}
	}
	if m.Treasure != nil {
		l = m.Treasure.Size()
		n += 1 + l + sovMission(uint64(l))
	}
	if m.CanRefresh {
		n += 2
	}
	if m.RefreshRemainTime != 0 {
		n += 1 + sovMission(uint64(m.RefreshRemainTime))
	}
	return n
}

func (m *TargetMission) Size() (n int) {
	var l int
	_ = l
	if m.ID != 0 {
		n += 1 + sovMission(uint64(m.ID))
	}
	return n
}

func (m *MissionReward) Size() (n int) {
	var l int
	_ = l
	if m.Jade != 0 {
		n += 1 + sovMission(uint64(m.Jade))
	}
	if m.Gold != 0 {
		n += 1 + sovMission(uint64(m.Gold))
	}
	if m.Bowlder != 0 {
		n += 1 + sovMission(uint64(m.Bowlder))
	}
	if m.NextMission != nil {
		l = m.NextMission.Size()
		n += 1 + l + sovMission(uint64(l))
	}
	return n
}

func (m *OpenMissionTreasureReply) Size() (n int) {
	var l int
	_ = l
	if m.TreasureReward != nil {
		l = m.TreasureReward.Size()
		n += 1 + l + sovMission(uint64(l))
	}
	if m.NextTreasure != nil {
		l = m.NextTreasure.Size()
		n += 1 + l + sovMission(uint64(l))
	}
	return n
}

func (m *UpdateMissionProcessArg) Size() (n int) {
	var l int
	_ = l
	if len(m.Missions) > 0 {
		for _, e := range m.Missions {
			l = e.Size()
			n += 1 + l + sovMission(uint64(l))
		}
	}
	return n
}

func sovMission(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMission(x uint64) (n int) {
	return sovMission(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Mission) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: Mission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Mission: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MissionID", wireType)
			}
			m.MissionID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MissionID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurCnt", wireType)
			}
			m.CurCnt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CurCnt |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsReward", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
			m.IsReward = bool(v != 0)
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func (m *MissionTreasure) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: MissionTreasure: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MissionTreasure: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreasureModelID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TreasureModelID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurCnt", wireType)
			}
			m.CurCnt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CurCnt |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func (m *MissionInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: MissionInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MissionInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Missions", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Missions = append(m.Missions, &Mission{})
			if err := m.Missions[len(m.Missions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Treasure", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Treasure == nil {
				m.Treasure = &MissionTreasure{}
			}
			if err := m.Treasure.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CanRefresh", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
			m.CanRefresh = bool(v != 0)
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefreshRemainTime", wireType)
			}
			m.RefreshRemainTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RefreshRemainTime |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func (m *TargetMission) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: TargetMission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TargetMission: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func (m *MissionReward) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: MissionReward: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MissionReward: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Jade", wireType)
			}
			m.Jade = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Jade |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gold", wireType)
			}
			m.Gold = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bowlder", wireType)
			}
			m.Bowlder = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Bowlder |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextMission", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.NextMission == nil {
				m.NextMission = &Mission{}
			}
			if err := m.NextMission.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func (m *OpenMissionTreasureReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: OpenMissionTreasureReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OpenMissionTreasureReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreasureReward", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextTreasure", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.NextTreasure == nil {
				m.NextTreasure = &MissionTreasure{}
			}
			if err := m.NextTreasure.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func (m *UpdateMissionProcessArg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMission
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
			return fmt.Errorf("proto: UpdateMissionProcessArg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UpdateMissionProcessArg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Missions", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMission
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
				return ErrInvalidLengthMission
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Missions = append(m.Missions, &Mission{})
			if err := m.Missions[len(m.Missions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMission
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
func skipMission(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMission
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
					return 0, ErrIntOverflowMission
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
					return 0, ErrIntOverflowMission
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
				return 0, ErrInvalidLengthMission
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMission
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
				next, err := skipMission(dAtA[start:])
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
	ErrInvalidLengthMission = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMission   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("mission.proto", fileDescriptorMission) }

var fileDescriptorMission = []byte{
	// 411 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xc1, 0xae, 0xd2, 0x40,
	0x14, 0x75, 0xfa, 0xde, 0x83, 0x72, 0x2b, 0xa0, 0x63, 0xd4, 0x86, 0x98, 0x4a, 0xba, 0xb1, 0x0b,
	0xc5, 0x04, 0x17, 0xae, 0x5c, 0x08, 0x24, 0xa6, 0x26, 0xa8, 0x19, 0xf1, 0x03, 0x8a, 0xbd, 0x60,
	0x63, 0xe9, 0x34, 0x33, 0x25, 0xe8, 0xce, 0x6f, 0x70, 0xe5, 0x8f, 0xf8, 0x0f, 0x2e, 0xfd, 0x04,
	0x83, 0x3f, 0x62, 0x3a, 0x9d, 0x29, 0x50, 0xf3, 0x12, 0x76, 0xf7, 0x9e, 0x73, 0x7a, 0xef, 0x99,
	0x73, 0x53, 0xe8, 0x6e, 0x12, 0x29, 0x13, 0x9e, 0x8d, 0x72, 0xc1, 0x0b, 0x4e, 0xad, 0x7c, 0x39,
	0xe8, 0x15, 0x02, 0x23, 0xb9, 0x15, 0x58, 0x61, 0xfe, 0x67, 0x68, 0xcf, 0x2b, 0x11, 0x7d, 0x00,
	0x1d, 0x5d, 0x86, 0x33, 0x97, 0x0c, 0x49, 0x70, 0xc5, 0x0e, 0x00, 0xbd, 0x07, 0xad, 0xe9, 0x56,
	0x4c, 0xb3, 0xc2, 0xb5, 0x14, 0xa5, 0x3b, 0x3a, 0x00, 0x3b, 0x94, 0x0c, 0x77, 0x91, 0x88, 0xdd,
	0x8b, 0x21, 0x09, 0x6c, 0x56, 0xf7, 0xb4, 0x07, 0x56, 0x38, 0x73, 0x2f, 0x95, 0xde, 0x0a, 0x67,
	0xfe, 0x7b, 0xe8, 0xeb, 0x81, 0x0b, 0xed, 0x82, 0x06, 0xd0, 0x37, 0xf5, 0x9c, 0xc7, 0x98, 0xea,
	0xd5, 0x1d, 0xd6, 0x84, 0xaf, 0x33, 0xe0, 0xff, 0x24, 0xe0, 0x18, 0x9b, 0xd9, 0x8a, 0xd3, 0x47,
	0x60, 0xeb, 0x56, 0xba, 0x64, 0x78, 0x11, 0x38, 0x63, 0x67, 0x94, 0x2f, 0x47, 0x1a, 0x63, 0x35,
	0x49, 0x9f, 0x82, 0x6d, 0x76, 0xa8, 0x91, 0xce, 0xf8, 0xce, 0x91, 0xd0, 0x50, 0xac, 0x16, 0x51,
	0x0f, 0x60, 0x1a, 0x65, 0x0c, 0x57, 0x02, 0xe5, 0x27, 0xfd, 0xd8, 0x23, 0x84, 0x3e, 0x86, 0xdb,
	0xba, 0x64, 0xb8, 0x89, 0x92, 0x6c, 0x91, 0x6c, 0x50, 0xbf, 0xfe, 0x7f, 0xc2, 0x7f, 0x08, 0xdd,
	0x45, 0x24, 0xd6, 0x58, 0x98, 0xfc, 0xab, 0xb4, 0x48, 0x9d, 0xd6, 0x37, 0x02, 0x5d, 0xe3, 0xba,
	0xca, 0x93, 0xc2, 0xe5, 0xeb, 0x28, 0x46, 0xad, 0x51, 0x75, 0x89, 0xbd, 0xe2, 0x69, 0xac, 0x43,
	0x51, 0x35, 0x75, 0xa1, 0x3d, 0xe1, 0xbb, 0x34, 0x46, 0xa1, 0x5c, 0x5e, 0x31, 0xd3, 0xd2, 0x27,
	0xe0, 0xbc, 0xc1, 0x2f, 0x66, 0xa5, 0x32, 0xd7, 0xc8, 0xe7, 0x98, 0xf7, 0xbf, 0x13, 0x70, 0xdf,
	0xe6, 0x98, 0x35, 0x33, 0xc1, 0x3c, 0xfd, 0x4a, 0x5f, 0x40, 0xef, 0x00, 0xa8, 0xfb, 0x13, 0x35,
	0xee, 0x6e, 0x39, 0xae, 0xfc, 0xea, 0x44, 0xce, 0x1a, 0x62, 0xfa, 0x1c, 0x6e, 0x96, 0xab, 0xce,
	0x39, 0xc1, 0x89, 0xd0, 0x9f, 0xc0, 0xfd, 0x0f, 0x79, 0x1c, 0x15, 0xa8, 0x65, 0xef, 0x04, 0xff,
	0x88, 0x52, 0xbe, 0x14, 0xeb, 0xb3, 0x6f, 0x3f, 0xb9, 0xf5, 0x6b, 0xef, 0x91, 0xdf, 0x7b, 0x8f,
	0xfc, 0xd9, 0x7b, 0xe4, 0xc7, 0x5f, 0xef, 0xc6, 0xb2, 0xa5, 0xfe, 0x87, 0x67, 0xff, 0x02, 0x00,
	0x00, 0xff, 0xff, 0x1a, 0x63, 0xbe, 0x1a, 0x34, 0x03, 0x00, 0x00,
}
