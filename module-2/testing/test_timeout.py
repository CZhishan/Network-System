# -*- coding: utf-8 -*-
"""
Created on Tue Nov  3 21:46:31 2020

@author: CHEN Zhishan
"""
from socket import socket
import time

testname = 'timeout_test1'
with open(testname+'.input', 'rt', newline='') as f:
    reqs = f.read()
f.close()

# Create connection to the server
s = socket()
s.connect(("localhost" , 8080))
# Compose the message/HTTP request we want to send to the server
msgPart1 = reqs.encode(encoding="utf-8")
# Send out the request
s.sendall(msgPart1)
# Listen for response and print it out
with open(testname+'.output', 'wb') as f:
    startTime = time.time()
    while True:
        data = s.recv(4096)
        if not data:
            break
        f.write(data)
    f.write(str.encode("This connection is closed after waiting for "+ \
    str(int(time.time()-startTime))+" seconds"))
f.close()
