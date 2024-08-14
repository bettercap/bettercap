#!/usr/bin/env /usr/bin/python3
import os
import json

# https://github.com/NordicSemiconductor/bluetooth-numbers-database

companies_source = "bluetooth-numbers-database/v1/company_ids.json"
companies_file = "network/ble_companies.go"
companies_template = """package network

var BLE_Companies = map[uint16]string{
	MAPPING
}"""

with open(companies_source, 'rt') as fp:
    companies = json.load(fp)

mapping = ""

for comp in companies:
    mapping += "  %d: %s,\n" % (comp['code'], json.dumps(comp['name']))

with open(companies_file, "w+t") as fp:
    fp.write(companies_template.replace("MAPPING", mapping.strip()))


def make_uuid(src):
    if '-' in src:
        return src
    else:
        # https://stackoverflow.com/questions/36212020/how-can-i-convert-a-bluetooth-16-bit-service-uuid-into-a-128-bit-uuid
        return '0000%s-0000-1000-8000-00805f9b34fb' % src


services_source = "bluetooth-numbers-database/v1/service_uuids.json"
services_file = "network/ble_services.go"
services_template = """package network

var BLE_Services = map[string]string{
	MAPPING
}"""

with open(services_source, 'rt') as fp:
    services = json.load(fp)

mapping = ""

for service in services:
    uuid = make_uuid(service['uuid'].lower())
    mapping += "  \"%s\": %s,\n" % (uuid, json.dumps(service['name']))

with open(services_file, "w+t") as fp:
    fp.write(services_template.replace("MAPPING", mapping.strip()))

chars_source = "bluetooth-numbers-database/v1/characteristic_uuids.json"
chars_file = "network/ble_characteristics.go"
chars_template = """package network

var BLE_Characteristics = map[string]string{
	MAPPING
}"""

with open(chars_source, 'rt') as fp:
    chars = json.load(fp)

mapping = ""

for char in chars:
    uuid = make_uuid(char['uuid'].lower())
    mapping += "  \"%s\": %s,\n" % (uuid, json.dumps(char['name']))

with open(chars_file, "w+t") as fp:
    fp.write(chars_template.replace("MAPPING", mapping.strip()))
