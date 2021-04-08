#!/usr/bin/python
import os
import six
import re

base = os.path.dirname(os.path.realpath(__file__))
# "https://code.wireshark.org/review/gitweb?p=wireshark.git;a=blob_plain;f=manuf;hb=HEAD"

with open(os.path.join(base, 'manuf.go.template')) as fp:
    template = fp.read()

with open(os.path.join(base, 'manuf')) as fp:
    lines = [l.strip() for l in fp.readlines()]
    lines = [l for l in lines if l != "" and l[0] != '#']    

def get_mac_and_mask(mac):
    # simple case 
    if not "/" in mac:
        mac_hex = mac.replace(":", '')
        mask    = 48 - 4 * len(mac_hex)
        mac_int = int(mac_hex, 16) << mask

    # 00:1B:C5:00:00:00/36
    else:
        parts = mac.split("/")
        mac_hex = parts[0].replace(":", '')
        mask    = 48 - int(parts[1])
        mac_int = int(mac_hex, 16) << mask 

    return (mac_int, mask)

index = {}

for line in lines:
    m = re.match( r'^([^\s]+)\s+([^\s]+)(.*)$', line, re.M)
    parts = m.groups()
    mac = parts[0]
    short = parts[1]
    manuf = parts[2].strip()
    if manuf == "":
        manuf  = short

    m = re.match( r'^([^#]+)#.+$', manuf)
    if m is not None:
        manuf = m.groups()[0].strip()

    mac_int, mask = get_mac_and_mask(mac)

    key = "%d.%d" % ( mask, mac_int >> mask )
    index[key] = manuf

code = "map[string]string {\n"

for key, vendor in six.iteritems(index):
    code += "    \"%s\": \"%s\",\n" % ( key, vendor.replace( '"', '\\"' ))

code += "}\n"

code = template.replace('#MAP#', code)

with open(os.path.join(base, 'manuf.go'), 'w+t') as fp:
    fp.write(code)
