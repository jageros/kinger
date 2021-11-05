#! /usr/bin/env python
# -*- coding: utf-8 -*-

import copy
import random
import pymongo

def gen_code(code_type, amount):
    """ 13位
        第3、6、7位组成批号
        第8、10、12位组成类型
        其他7位随机
        例如：1203401010718
        批号001, 类型001
    """

    if amount > 40000:
        raise Exception("too big amount, max = 40000")

    conn = pymongo.MongoClient("127.0.0.1", 27017)
    db = conn.kingwar
    #db.authenticate(MONGO_USER, MONGO_PWD)
    db.uuid.update({"_id":"codeBatch:" + str(code_type)}, {"$inc":{"n": 1}}, True)
    batch_info = db.uuid.find_one({"_id":"codeBatch:" + str(code_type)})
    if not batch_info:
        raise Exception("fuck")
    batch = batch_info["n"]
    if batch > 999:
        raise Exception("too big batch, max=999, you need new type")

    code_format = "%s%s%s%s%s%s%s%s%s%s%s%s"
    bit = ["", "", "", "", "", "", "", "", "", "", "", ""]
    random_list = range(10000000)
    str_batch = "%03d" % batch
    bit[2] = str_batch[0]
    bit[5] = str_batch[1]
    bit[6] = str_batch[2]

    str_type = "%03d" % code_type
    bit[7] = str_type[0]
    bit[9] = str_type[1]
    bit[11] = str_type[2]

    random_code = random.sample(random_list, amount)
    fp = open(str(code_type) + "_" + str(batch), "a", 0)
    for int_code in random_code:
        code = "%07d" % int_code
        bit[0] = code[0]
        bit[1] = code[1]
        bit[3] = code[2]
        bit[4] = code[3]
        bit[8] = code[4]
        bit[10] = code[5]
        bit[12] = code[6]
        str_code = code_format % tuple(bit)
        db.giftcode.insert_one({"_id":str_code, "data":{"type":code_type}})
        fp.write("%s\n"%str_code)
    fp.close()

if __name__ == "__main__":
    gen_code(1, 1000)

