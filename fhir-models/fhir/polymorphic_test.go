// Copyright 2019 - 2022 The Samply Community
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fhir_test

import (
	"encoding/json"
	"testing"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// assertAbsentKey fails the test if key is present in the JSON object.
func assertAbsentKey(t *testing.T, label string, raw []byte, key string) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s: unmarshal for key check: %v", label, err)
	}
	if _, ok := m[key]; ok {
		t.Errorf("%s: key %q must be absent in JSON to avoid non-compliant FHIR output, but it is present", label, key)
	}
}

// assertPresentKey fails the test if key is absent from the JSON object.
func assertPresentKey(t *testing.T, label string, raw []byte, key string) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s: unmarshal for key check: %v", label, err)
	}
	if _, ok := m[key]; !ok {
		t.Errorf("%s: key %q is absent from JSON but should be present", label, key)
	}
}

// baseImmunization returns the minimum required fields for a valid Immunization
// so individual tests only need to set the fields under test.
func baseImmunization() fhir.Immunization {
	trueVal := true
	return fhir.Immunization{
		Status: fhir.ImmunizationStatusCodesCompleted,
		VaccineCode: fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strPtr("http://hl7.org/fhir/sid/cvx"), Code: strPtr("140")}},
		},
		Patient:       fhir.Reference{Reference: strPtr("Patient/pat-1")},
		PrimarySource: &trueVal,
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// ---------------------------------------------------------------------------
// occurrence[x]: dateTime vs string
// ---------------------------------------------------------------------------

// TestImmunizationOccurrenceDateTimeVariant verifies that when occurrenceDateTime
// is set, the serialized JSON contains only "occurrenceDateTime" and not
// "occurrenceString", and that the value survives a round-trip.
func TestImmunizationOccurrenceDateTimeVariant(t *testing.T) {
	imm := baseImmunization()
	imm.OccurrenceDateTime = strPtr("2023-09-15T10:30:00Z")

	b, err := json.Marshal(imm)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	assertPresentKey(t, "occurrenceDateTime", b, "occurrenceDateTime")
	assertAbsentKey(t, "occurrenceString", b, "occurrenceString")

	// round-trip
	imm2, err := fhir.UnmarshalImmunization(b)
	if err != nil {
		t.Fatalf("UnmarshalImmunization: %v", err)
	}
	if imm2.OccurrenceDateTime == nil || *imm2.OccurrenceDateTime != "2023-09-15T10:30:00Z" {
		t.Errorf("RT OccurrenceDateTime: got %v, want \"2023-09-15T10:30:00Z\"", imm2.OccurrenceDateTime)
	}
	if imm2.OccurrenceString != nil {
		t.Errorf("RT OccurrenceString: expected nil, got %v", *imm2.OccurrenceString)
	}
}

// TestImmunizationOccurrenceStringVariant verifies the complementary case:
// occurrenceString set → only "occurrenceString" in JSON.
func TestImmunizationOccurrenceStringVariant(t *testing.T) {
	imm := baseImmunization()
	imm.OccurrenceString = strPtr("approximately 2 weeks ago")

	b, err := json.Marshal(imm)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	assertPresentKey(t, "occurrenceString", b, "occurrenceString")
	assertAbsentKey(t, "occurrenceDateTime", b, "occurrenceDateTime")

	// round-trip
	imm2, err := fhir.UnmarshalImmunization(b)
	if err != nil {
		t.Fatalf("UnmarshalImmunization: %v", err)
	}
	if imm2.OccurrenceString == nil || *imm2.OccurrenceString != "approximately 2 weeks ago" {
		t.Errorf("RT OccurrenceString: got %v", imm2.OccurrenceString)
	}
	if imm2.OccurrenceDateTime != nil {
		t.Errorf("RT OccurrenceDateTime: expected nil, got %v", *imm2.OccurrenceDateTime)
	}
}

// TestImmunizationOccurrenceFromJSON verifies deserialization from FHIR JSON
// for both variants — the unused variant must remain nil.
func TestImmunizationOccurrenceFromJSON(t *testing.T) {
	t.Run("dateTime", func(t *testing.T) {
		const input = `{
			"resourceType": "Immunization",
			"status": "completed",
			"vaccineCode": {"coding": [{"system": "http://hl7.org/fhir/sid/cvx", "code": "140"}]},
			"patient": {"reference": "Patient/pat-1"},
			"occurrenceDateTime": "2023-09-15T10:30:00Z"
		}`
		imm, err := fhir.UnmarshalImmunization([]byte(input))
		if err != nil {
			t.Fatalf("UnmarshalImmunization: %v", err)
		}
		if imm.OccurrenceDateTime == nil || *imm.OccurrenceDateTime != "2023-09-15T10:30:00Z" {
			t.Errorf("OccurrenceDateTime: got %v", imm.OccurrenceDateTime)
		}
		if imm.OccurrenceString != nil {
			t.Errorf("OccurrenceString: expected nil, got %q", *imm.OccurrenceString)
		}
	})

	t.Run("string", func(t *testing.T) {
		const input = `{
			"resourceType": "Immunization",
			"status": "completed",
			"vaccineCode": {"coding": [{"system": "http://hl7.org/fhir/sid/cvx", "code": "140"}]},
			"patient": {"reference": "Patient/pat-1"},
			"occurrenceString": "approximately 2 weeks ago"
		}`
		imm, err := fhir.UnmarshalImmunization([]byte(input))
		if err != nil {
			t.Fatalf("UnmarshalImmunization: %v", err)
		}
		if imm.OccurrenceString == nil || *imm.OccurrenceString != "approximately 2 weeks ago" {
			t.Errorf("OccurrenceString: got %v", imm.OccurrenceString)
		}
		if imm.OccurrenceDateTime != nil {
			t.Errorf("OccurrenceDateTime: expected nil, got %q", *imm.OccurrenceDateTime)
		}
	})
}

// ---------------------------------------------------------------------------
// protocolApplied[].doseNumber[x] and seriesDoses[x]
// ---------------------------------------------------------------------------

// TestImmunizationProtocolAppliedDoseNumberVariants verifies that doseNumber[x]
// and seriesDoses[x] inside a backbone element serialize to exactly one variant.
func TestImmunizationProtocolAppliedDoseNumberVariants(t *testing.T) {
	t.Run("doseNumberPositiveInt/seriesDosesPositiveInt", func(t *testing.T) {
		imm := baseImmunization()
		imm.OccurrenceDateTime = strPtr("2023-09-15")
		imm.ProtocolApplied = []fhir.ImmunizationProtocolApplied{{
			DoseNumberPositiveInt:  intPtr(1),
			SeriesDosesPositiveInt: intPtr(2),
		}}

		b, err := json.Marshal(imm)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}

		// unwrap protocolApplied[0]
		var top map[string]json.RawMessage
		_ = json.Unmarshal(b, &top)
		var protocols []json.RawMessage
		_ = json.Unmarshal(top["protocolApplied"], &protocols)
		if len(protocols) == 0 {
			t.Fatal("protocolApplied is empty in JSON")
		}
		proto := protocols[0]

		assertPresentKey(t, "doseNumberPositiveInt", proto, "doseNumberPositiveInt")
		assertAbsentKey(t, "doseNumberString", proto, "doseNumberString")
		assertPresentKey(t, "seriesDosesPositiveInt", proto, "seriesDosesPositiveInt")
		assertAbsentKey(t, "seriesDosesString", proto, "seriesDosesString")

		// round-trip
		imm2, err := fhir.UnmarshalImmunization(b)
		if err != nil {
			t.Fatalf("UnmarshalImmunization: %v", err)
		}
		p := imm2.ProtocolApplied[0]
		if p.DoseNumberPositiveInt == nil || *p.DoseNumberPositiveInt != 1 {
			t.Errorf("RT DoseNumberPositiveInt: got %v, want 1", p.DoseNumberPositiveInt)
		}
		if p.DoseNumberString != nil {
			t.Errorf("RT DoseNumberString: expected nil, got %q", *p.DoseNumberString)
		}
		if p.SeriesDosesPositiveInt == nil || *p.SeriesDosesPositiveInt != 2 {
			t.Errorf("RT SeriesDosesPositiveInt: got %v, want 2", p.SeriesDosesPositiveInt)
		}
		if p.SeriesDosesString != nil {
			t.Errorf("RT SeriesDosesString: expected nil, got %q", *p.SeriesDosesString)
		}
	})

	t.Run("doseNumberString/seriesDosesString", func(t *testing.T) {
		imm := baseImmunization()
		imm.OccurrenceDateTime = strPtr("2023-09-15")
		imm.ProtocolApplied = []fhir.ImmunizationProtocolApplied{{
			DoseNumberString:  strPtr("first"),
			SeriesDosesString: strPtr("three"),
		}}

		b, err := json.Marshal(imm)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}

		var top map[string]json.RawMessage
		_ = json.Unmarshal(b, &top)
		var protocols []json.RawMessage
		_ = json.Unmarshal(top["protocolApplied"], &protocols)
		proto := protocols[0]

		assertPresentKey(t, "doseNumberString", proto, "doseNumberString")
		assertAbsentKey(t, "doseNumberPositiveInt", proto, "doseNumberPositiveInt")
		assertPresentKey(t, "seriesDosesString", proto, "seriesDosesString")
		assertAbsentKey(t, "seriesDosesPositiveInt", proto, "seriesDosesPositiveInt")

		// round-trip
		imm2, err := fhir.UnmarshalImmunization(b)
		if err != nil {
			t.Fatalf("UnmarshalImmunization: %v", err)
		}
		p := imm2.ProtocolApplied[0]
		if p.DoseNumberString == nil || *p.DoseNumberString != "first" {
			t.Errorf("RT DoseNumberString: got %v, want \"first\"", p.DoseNumberString)
		}
		if p.DoseNumberPositiveInt != nil {
			t.Errorf("RT DoseNumberPositiveInt: expected nil, got %v", *p.DoseNumberPositiveInt)
		}
	})
}

// TestImmunizationFullRoundTrip exercises a complete, realistic Immunization
// with occurrenceDateTime, protocolApplied with doseNumber, and verifies the
// entire structure survives a marshal → unmarshal cycle unchanged.
func TestImmunizationFullRoundTrip(t *testing.T) {
	const input = `{
		"resourceType": "Immunization",
		"id": "imm-1",
		"status": "completed",
		"vaccineCode": {
			"coding": [{"system": "http://hl7.org/fhir/sid/cvx", "code": "140", "display": "Influenza, seasonal, injectable, preservative free"}]
		},
		"patient": {"reference": "Patient/pat-1"},
		"occurrenceDateTime": "2023-09-15T10:30:00Z",
		"primarySource": true,
		"lotNumber": "AAJN11K",
		"expirationDate": "2025-02-15",
		"site": {"coding": [{"system": "http://terminology.hl7.org/CodeSystem/v3-ActSite", "code": "LA"}]},
		"route": {"coding": [{"system": "http://terminology.hl7.org/CodeSystem/v3-RouteOfAdministration", "code": "IM"}]},
		"doseQuantity": {"value": 0.5, "unit": "mL"},
		"performer": [{"actor": {"reference": "Practitioner/pract-1"}}],
		"protocolApplied": [{
			"series": "Seasonal Influenza",
			"targetDisease": [{"coding": [{"system": "http://snomed.info/sct", "code": "6142004"}]}],
			"doseNumberPositiveInt": 1,
			"seriesDosesPositiveInt": 1
		}]
	}`

	imm, err := fhir.UnmarshalImmunization([]byte(input))
	if err != nil {
		t.Fatalf("UnmarshalImmunization: %v", err)
	}

	// Spot-check key fields
	if imm.OccurrenceDateTime == nil || *imm.OccurrenceDateTime != "2023-09-15T10:30:00Z" {
		t.Errorf("OccurrenceDateTime: got %v", imm.OccurrenceDateTime)
	}
	if imm.OccurrenceString != nil {
		t.Errorf("OccurrenceString: must be nil when occurrenceDateTime is set, got %q", *imm.OccurrenceString)
	}
	if len(imm.ProtocolApplied) == 0 {
		t.Fatal("ProtocolApplied: empty")
	}
	p := imm.ProtocolApplied[0]
	if p.DoseNumberPositiveInt == nil || *p.DoseNumberPositiveInt != 1 {
		t.Errorf("DoseNumberPositiveInt: got %v, want 1", p.DoseNumberPositiveInt)
	}
	if p.DoseNumberString != nil {
		t.Errorf("DoseNumberString: must be nil, got %q", *p.DoseNumberString)
	}

	// Serialize and confirm no spurious polymorphic keys leak into output
	b, err := json.Marshal(imm)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	assertPresentKey(t, "occurrenceDateTime", b, "occurrenceDateTime")
	assertAbsentKey(t, "occurrenceString", b, "occurrenceString")

	var top map[string]json.RawMessage
	_ = json.Unmarshal(b, &top)
	var protocols []json.RawMessage
	_ = json.Unmarshal(top["protocolApplied"], &protocols)
	proto := protocols[0]
	assertPresentKey(t, "doseNumberPositiveInt", proto, "doseNumberPositiveInt")
	assertAbsentKey(t, "doseNumberString", proto, "doseNumberString")
	assertPresentKey(t, "seriesDosesPositiveInt", proto, "seriesDosesPositiveInt")
	assertAbsentKey(t, "seriesDosesString", proto, "seriesDosesString")

	// Second round-trip: unmarshal the output and compare
	imm2, err := fhir.UnmarshalImmunization(b)
	if err != nil {
		t.Fatalf("UnmarshalImmunization (2nd): %v", err)
	}
	if imm2.LotNumber == nil || *imm2.LotNumber != "AAJN11K" {
		t.Errorf("RT LotNumber: got %v, want \"AAJN11K\"", imm2.LotNumber)
	}
	if len(imm2.ProtocolApplied) == 0 || imm2.ProtocolApplied[0].Series == nil {
		t.Error("RT ProtocolApplied.Series: lost after round-trip")
	}
}
