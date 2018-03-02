#!/usr/bin/python
import os
import six

base = os.path.dirname(os.path.realpath(__file__))

with open(os.path.join(base, 'oui.go.template')) as fp:
    template = fp.read()

with open(os.path.join(base, 'oui.dat')) as fp:
    lines = [l.strip() for l in fp.readlines()]

m = {}
for line in lines:
    if line == "" or line[0] == '#':
        continue

    parts = line.split(' ', 1)
    if len(parts) != 2:
        continue

    prefix = parts[0].strip().lower()
    vendor = parts[1].strip()

    m[prefix] = vendor

code = "map[string]string {\n"

for prefix, vendor in six.iteritems(m):
    code += "    \"%s\": \"%s\",\n" % ( prefix, vendor )

code += "}\n"

code = template.replace('#MAP#', code)

with open(os.path.join(base, 'oui.go'), 'w+t') as fp:
    fp.write(code)
