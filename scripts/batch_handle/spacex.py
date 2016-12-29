#!/usr/bin/python
#encoding:utf-8
'''
Author: chenwenjiang
Email: chenwenjiang@oneniceapp.com
''' 
import sys
reload(sys)
sys.setdefaultencoding("utf-8")
import os

sys.path.append(os.path.join(os.path.dirname(os.path.abspath(__file__)), "../../tinder/python"))

import json
from util.helpers import mysql_helper

confile = sys.argv[1]
version = sys.argv[2]

db_helper = mysql_helper.get_slave_db(db = "ad_base_db", confile = confile)

def packaging(uid, immutable_box, mutable_box):
    tbname = "ad_user_info_%s" % (int(uid) % 100)
    sql = """INSERT INTO {tbname} (`uid`, `immutable_payload`, `mutable_payload`, `version`) VALUES (%s, %s, %s, %s)""".format(tbname=tbname)
    #db_helper.execute_many(sql, [(uid, json.dumps(immutable_box), json.dumps(mutable_box), version)])
    return sql, (uid, json.dumps(immutable_box), json.dumps(mutable_box), version)

def foreach():

    buffers = {}    
    
    n = 1   
    for line in sys.stdin:
        line = line.rstrip()
        line = line.decode("utf-8", "ignore")
        parts = line.split("\001")
        uid, name, gender, age, channel, ctime, pf, tags = parts
        tags = tags.lower()
        if tags != "\\n":
            tags = tags.split("\002")
        else:
            tags = [""]
        immutable_box = {
                "name":name,
                "gender":gender,
                "age":int(age),
                "download_channel":channel,
                "ctime":int(ctime),
                "platform":pf,
            }
        
        mutable_box = {
                "tags": tags
            }
        
        sql, data = packaging(uid, immutable_box, mutable_box)    
        buffers.setdefault(sql, []).append(data)
        n += 1
    
        if n % 500 == 0:
            if buffers:
                dumps(buffers)
            buffers = {}
            n = 1
    else:
        if buffers:
            dumps(buffers)
        buffers = {}

def dumps(buffers):
    
    for sql, data in buffers.iteritems():
        print sql, "dump size:%s" % len(data)
        db_helper.execute_many(sql, data)

def updateVersionPointer():
    sql = """UPDATE ad_cur_version SET `version` = %s """ % version
    print sql
    db_helper.execute(sql)

def main():
    foreach()
    updateVersionPointer()

if __name__ == "__main__":
    main()

