#! /usr/bin/env python
# -*- coding: utf-8 -*-

import pymongo

MONGO_IP = "127.0.0.1"
MONG_PORT = 27017
MONGO_USER = ""
MONGO_PWD = "1qa@WS3ed_[kingwar]"
MONGO_DB = "kingwar"

REGION_MONGO_IP = "127.0.0.1"
REGION_MONG_PORT = 27017
REGION_MONGO_USER = ""
REGION_MONGO_PWD = "1qa@WS3ed_[kingwar]"
REGION_MONGO_DB = "kingwar"

def get_mongo_by_config(ip, port, database, user, pwd):
    conn = pymongo.MongoClient(ip, port)
    db = conn[database]
    if user:
        db.authenticate(user, pwd, mechanism='SCRAM-SHA-1')
    return db

def get_mongo(is_global):
    if is_global:
        return get_mongo_by_config(MONGO_IP, MONG_PORT, MONGO_DB, MONGO_USER, MONGO_PWD)
    else:
        return get_mongo_by_config(REGION_MONGO_IP, REGION_MONG_PORT, REGION_MONGO_DB, REGION_MONGO_USER, REGION_MONGO_PWD)

