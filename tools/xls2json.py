#coding:utf-8

import sys
try:
    reload(sys)
    sys.setdefaultencoding('utf-8')
except:
    pass

import os
import re
import codecs
import xlrd
import pymongo
import config

XLS_PATH = os.getenv("HOME") +  "/puppy/king_design/exl/"
XLS_PATH_PC = os.getenv("HOME") +  "/puppy/king_design/exlpc/"
DST_PATH = "jsondata/"

reward_tbls = []

def get_xls_files(device):
    xlspath = XLS_PATH
    if device == "pc":
        xlspath = XLS_PATH_PC
    xlsx_files = []
    for f in os.listdir(xlspath):
        if f[-4:] == "xlsx":
            xlsx_files.append(xlspath + f)
    return xlsx_files

class StringType(object):
    def gen_json_str(self, value):
        return "\"%s\"" % str(value)

class IntType(object):
    def gen_json_str(self, value):
        if value:
            value = int(value)
        else:
            value = 0
        return str(int(value))

class FloatType(object):
    def gen_json_str(self, value):
        if not value:
            value = 0.0
        return str(value)

class ListType(object):
    def __init__(self, elemType):
        self.elemType = elemType

    def gen_json_str(self, value):
        if not value:
            return "[]"
        strVlist = str(value).split(";")
        vlist = []
        for v in strVlist:
            elem = self.elemType.gen_json_str(v)
            if elem and elem != "[]":
                vlist.append(elem)
        return "[%s]" % ", ".join(vlist)

class ArgListType(object):
    def __init__(self, elemType):
        self.elemType = elemType

    def gen_json_str(self, value):
        if not value:
            return "[]"
        strVlist = str(value).split(":")
        vlist = []
        for v in strVlist:
            elem = self.elemType.gen_json_str(v)
            if elem and elem != "[]":
                vlist.append(elem)
        return "[%s]" % ", ".join(vlist)

def gen_key_type(_type):
    if _type == "str":
        return StringType()
    if _type == "int":
        return IntType()
    if _type == "text":
        return IntType()
    if _type == "float":
        return FloatType()
    if _type[:4] == "list":
        elemType = gen_key_type(_type[5:-1])
        return ListType(elemType)
    if _type[:7] == "arglist":
        elemType = gen_key_type(_type[8:-1])
        return ArgListType(elemType)
    return None

# get the key name and clumn type
def process_key(key):
    # Rules:
    # 1. no key the clumn not output
    # 2. if the key start with '#' then the clumn not output( think of python comment )
    # 3. if the key start with '--' then the clumn not output( think of lua comment )
    # key has the following type: int, float, str, py, lua, default
    if key == "": return None, None
    if re.match(r"^#.*", key): return None, None
    if re.match(r"\-\-.*", key): return None, None
    ls = key.replace(")", "").split("(")
    ls[1] = gen_key_type( ls[1].replace(" ", "") )
    return ls[0], ls[1]

def process_header(data):
    key_row = None              # the key of the dict is in which row
    for row in data:
        if len(row) < 2:
            continue
        elif row[0] == "key_row":
            key_row = row[1]
            break
    return key_row

def need_output(sheet):
    try:
        return (sheet.cell_value(0,0) == "server_output" and sheet.cell_value(0,1)) or sheet.name == "text"
    except IndexError:
        return False
    return False


def dump_to_mongo(name, data):
    print("begin dump_to_mongo")
    conn = pymongo.MongoClient(config.MONGO_IP, config.MONG_PORT)
    db = conn[config.MONGO_DB]
    if config.MONGO_USER:
        db.authenticate(config.MONGO_USER, config.MONGO_PWD, mechanism='SCRAM-SHA-1')
    db.gamedata.update({"_id": name}, {"data":{"data":data}}, upsert=True)
    print("dump_to_mongo ok")

def sheet2json(sheet_data, dataname):
    jsonfilename = DST_PATH + dataname + ".json"
    key_row = process_header(sheet_data)
    keys = [ process_key(key) for key in sheet_data[key_row-1] ]
    real_data = ( row for row in iter(sheet_data[key_row:]) if any(row) )
    f = codecs.open(jsonfilename, "w", "utf-8")
    f.write("[\n")

    flagi = 0
    for i, row in enumerate(real_data):
        _v = row[0]
        if _v == "":
            continue

        if flagi == 0:
            f.write("\t{ ")
        else:
            f.write(",\n\t{ ")

        flagi += 1
        flagj = 0
        for j, (k, t) in enumerate(keys):
            if not k:
                continue
            if flagj == 0:
                f.write("\"%s\": " % k)
            else:
                f.write(", \"%s\": " % k)

            v = row[j]
            if t:
                f.write(t.gen_json_str(v))
            else:
                print("what the fuck type")
                print(jsonfilename)
                print("i=%d, j=%d" % (i, j))
                print("k=%s, v=%s, t=%s" % (k, v, t))
                return
            flagj += 1

        f.write(" }")

    f.write("\n]\n")
    f.close()
    f = codecs.open(jsonfilename, "r", "utf-8")
    data = f.read()
    f.close()
    dump_to_mongo(dataname, data)

    if dataname.find("info_reward_") == 0:
        reward_tbls.append(dataname)

def format_value(value, vtype):
    if vtype == 2 and value == int(value):
        return int(value)
    else:
        return value

def sheet2dict(sheet):
    data = []
    for ridx in range(sheet.nrows):
        row = []
        for cidx in range(sheet.ncols):
            value = sheet.cell_value(ridx, cidx)
            vtype = sheet.cell_type(ridx, cidx)
            v = format_value(value, vtype)
            row.append(v)
        data.append(row)
    return data

def parse(device):
    #xls_file = XLS_PATH + get_xls_file(device) + ".xlsx"
    for xls_file in get_xls_files(device):
        if "$" in xls_file:
            continue
        book = xlrd.open_workbook(xls_file)
        sheets = filter( need_output, book.sheets() )
        for sheet in sheets:
            print("i am fucking %s ..." % sheet.name)
            sdata = sheet2dict(sheet)
            sheet2json(sdata, sheet.name)
            print("%s fuck done" % sheet.name)

    dump_to_mongo("__reward_tbl__", reward_tbls)

def usage():
    print("""\
-----------------------------------------
usage:
    python xls2json.py <device>

    python xls2json.py mobile
    python xls2json.py pc
""")

if __name__ == "__main__":
    parse("mobile")

