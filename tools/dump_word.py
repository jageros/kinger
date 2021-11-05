#! /usr/bin/env python
# -*- coding: utf-8 -*-

import json
import pymongo
import config

def dump():
    conn = pymongo.MongoClient(config.MONGO_IP, config.MONG_PORT)
    db = conn[config.MONGO_DB]
    if config.MONGO_USER:
        db.authenticate(config.MONGO_USER, config.MONGO_PWD, mechanism='SCRAM-SHA-1')

    f = open("jsondata/dirty_words.json", "r")
    words = json.loads(f.read())
    f.close()

    db.gamedata.update({"_id": "dirty_word"}, {"data":{"words":words}}, upsert=True)

if __name__ == "__main__":
    dump()

