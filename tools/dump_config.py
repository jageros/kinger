#! /usr/bin/env python
# -*- coding: utf-8 -*-

import sys
import config

def dump(is_global):
    db = config.get_mongo(False)
    f = open("region.toml", "r")
    region_cfg = f.read()
    f.close()
    db.config.update({"_id":"region"}, {"data":{"data":region_cfg}}, upsert=True)

    if not is_global:
        return

    db = config.get_mongo(True)
    f = open("gopuppy.toml", "r")
    gopuppy_cfg = f.read()
    f.close()
    db.config.update({"_id":"gopuppy"}, {"data":{"data":gopuppy_cfg}}, upsert=True)

    f = open("kingwar.toml", "r")
    kingwar_cfg = f.read()
    f.close()
    db.config.update({"_id":"kingwar"}, {"data":{"data":kingwar_cfg}}, upsert=True)

if __name__ == "__main__":
    dump(len(sys.argv) > 1)

