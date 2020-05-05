#!/usr/bin/python
import os
import sys
import yaml

try:
    service = sys.argv[1]
except:
    service = 'journal'

base = '/var/tmp/jacobi/gocode/src/github.com/mikerjacobi/lambda-journal'
services = yaml.safe_load(open(base+'/server/services.yaml'))

params = {"service":service}
for handler in services[service]:
    params["handler"] = handler
    os.system('cd server && env GOOS=linux go build -ldflags="-s -w" -o bin/{handler} ./{service}/{handler}/...'.format(**params))
    os.system('docker service scale -d journal_{handler}=0 && docker service scale -d journal_{handler}=1'.format(**params))
