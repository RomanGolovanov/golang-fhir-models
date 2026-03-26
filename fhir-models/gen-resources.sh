#!/usr/bin/env bash

wget -O definitions.zip https://www.hl7.org/fhir/R4/definitions.json.zip
unzip definitions.zip profiles-resources.json profiles-types.json valuesets.json -d fhir
rm definitions.zip

go generate ./fhir
