#coding:utf-8

import os

def write_header():
	data = """// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"king_server/proto/pb"
	"gopuppy/network/protoc"
)

"""
	return data

def write_msg_struct(msgid, notes):
	data = """%s
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type %s_Meta struct {
}

func (m *%s_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_%s
}

"""
	data = data % (notes, msgid, msgid, msgid)
	return data

def write_encode_arg(msgid, req):
	if not req:
		data = """func (m *%s_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

"""
		data = data % msgid
	else:
		data = """func (m *%s_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.%s)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("%s_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

"""
		data = data % (msgid, req, msgid)

	return data

def write_decode_arg(msgid, req):
	if not req:
		data = """func (m *%s_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

"""
		data = data % msgid
	else:
		data = """func (m *%s_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.%s{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

"""
		data = data % (msgid, req)

	return data

def write_encode_reply(msgid, resp):
	if not resp or resp == "ok":
		data = """func (m *%s_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

"""
		data = data % msgid
	else:
		data = """func (m *%s_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.%s)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("%s_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

"""
		data = data % (msgid, resp, msgid)

	return data

def write_decode_reply(msgid, resp):
	if not resp or resp == "ok":
		data = """func (m *%s_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

"""
		data = data % msgid
	else:
		data = """func (m *%s_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.%s{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

"""
		data = data % (msgid, resp)

	return data

def write_end(msgid):
	data = """//------------------------------------ {0} END ----------------------------------------

"""
	data = data.format(msgid)
	return data

def write_msg_meta(msgid, req, resp, notes):
	data = write_msg_struct(msgid, notes)
	data += write_encode_arg(msgid, req)
	data += write_decode_arg(msgid, req)
	data += write_encode_reply(msgid, resp)
	data += write_decode_reply(msgid, resp)
	data += write_end(msgid)
	return data

def write_meta_file(filename, rpc_info):
	if not rpc_info:
		return

	data = write_header()

	for info in rpc_info:
		msgid = info[0]
		req = info[1]
		resp = info[2]
		notes = info[3]
		data += write_msg_meta(msgid, req, resp, notes)

	f = open("meta/" + filename + ".go", "w")
	f.write(data)
	f.close()

def write_meta_register(all_msgid):
	data = """// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"gopuppy/network/protoc"
)

func init()  {

"""
	register_data = ""
	for msgid in all_msgid:
		register_data += "    protoc.RegisterMeta(&%s_Meta{})\n" % msgid

	data += register_data
	data += """
}
"""
	f = open("meta/meta_register.go", "w")
	f.write(data)
	f.close()

def fuck():
	all_msgid = []
	for f in os.listdir("proto/pbdef"):
		if f[-6:] == ".proto":
			pbdata = ""
			pf = open("proto/pbdef/" + f, "r")
			rpc_info = []
			for l in pf:
				l = l.strip()
				if l[:3] == "//@":
					msg_info = l.split()
					msgid = msg_info[1]

					if "req:" in l:
						req = msg_info[3]
					else:
						req = ""

					if "resp:" in l:
						if req:
							i = 5
						else:
							i = 3
						resp = msg_info[i]
					else:
						resp = ""

					rpc_info.append([msgid, req, resp, l])
					all_msgid.append(msgid)

			pf.close()
			write_meta_file(f[:-6], rpc_info)

	write_meta_register(all_msgid)

if __name__ == "__main__":
	fuck()
