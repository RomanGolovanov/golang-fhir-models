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

// assertPrimitiveElement is a helper that fatally fails if the companion element
// is nil, then checks that the extension URL at index 0 matches want.
func assertPrimitiveElement(t *testing.T, label string, elem *fhir.PrimitiveElement, wantURL string) {
	t.Helper()
	if elem == nil {
		t.Fatalf("%s: PrimitiveElement is nil — extension was lost", label)
	}
	if len(elem.Extension) == 0 {
		t.Fatalf("%s: Extension slice is empty", label)
	}
	if elem.Extension[0].Url != wantURL {
		t.Errorf("%s: Extension[0].Url = %q, want %q", label, elem.Extension[0].Url, wantURL)
	}
}

// roundTrip marshals v to JSON and unmarshals into dst, returning the raw bytes.
func roundTrip(t *testing.T, v interface{}, dst interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("Unmarshal after round-trip: %v", err)
	}
	return b
}

// ---------------------------------------------------------------------------
// Patient
// ---------------------------------------------------------------------------

// TestPatientPrimitiveExtensions exercises a realistic Patient with primitive
// extensions on multiple fields of different types:
//   - string field:  birthDate / _birthDate
//   - bool field:    active / _active
//   - code/enum:     gender / _gender
//   - nested bool:   communication[0].preferred / _preferred
//   - array strings: name[0].given / _given (parallel array, one null entry)
func TestPatientPrimitiveExtensions(t *testing.T) {
	const input = `{
		"resourceType": "Patient",
		"id": "pat-1",
		"active": true,
		"_active": {
			"extension": [{"url": "http://example.com/ext/verified", "valueBoolean": true}]
		},
		"gender": "male",
		"_gender": {
			"id": "gen-1",
			"extension": [{"url": "http://example.com/ext/gender-identity", "valueString": "male"}]
		},
		"birthDate": "1985-03-22",
		"_birthDate": {
			"extension": [{"url": "http://example.com/ext/birth-time", "valueDateTime": "1985-03-22T04:15:00Z"}]
		},
		"name": [{
			"family": "Smith",
			"given": ["Alice", "Marie"],
			"_given": [
				{"extension": [{"url": "http://example.com/ext/preferred", "valueBoolean": true}]},
				null
			]
		}],
		"communication": [{
			"language": {"coding": [{"system": "urn:ietf:bcp:47", "code": "en"}]},
			"preferred": true,
			"_preferred": {
				"extension": [{"url": "http://example.com/ext/lang-source", "valueString": "self-reported"}]
			}
		}]
	}`

	patient, err := fhir.UnmarshalPatient([]byte(input))
	if err != nil {
		t.Fatalf("UnmarshalPatient: %v", err)
	}

	// --- active (bool) ---
	if patient.Active == nil || !*patient.Active {
		t.Error("Active: expected true")
	}
	assertPrimitiveElement(t, "ActiveElement", patient.ActiveElement, "http://example.com/ext/verified")

	// --- gender (code/enum) ---
	assertPrimitiveElement(t, "GenderElement", patient.GenderElement, "http://example.com/ext/gender-identity")
	if patient.GenderElement.Id == nil || *patient.GenderElement.Id != "gen-1" {
		t.Errorf("GenderElement.Id: got %v, want \"gen-1\"", patient.GenderElement.Id)
	}

	// --- birthDate (string) ---
	if patient.BirthDate == nil || *patient.BirthDate != "1985-03-22" {
		t.Errorf("BirthDate: got %v, want \"1985-03-22\"", patient.BirthDate)
	}
	assertPrimitiveElement(t, "BirthDateElement", patient.BirthDateElement, "http://example.com/ext/birth-time")

	// --- name[0].given (string array, parallel _given) ---
	if len(patient.Name) == 0 {
		t.Fatal("Name slice is empty")
	}
	name := patient.Name[0]
	if len(name.Given) != 2 {
		t.Fatalf("Given: got %d, want 2", len(name.Given))
	}
	if len(name.GivenElement) != 2 {
		t.Fatalf("GivenElement: got %d, want 2", len(name.GivenElement))
	}
	if name.GivenElement[0] == nil {
		t.Fatal("GivenElement[0]: expected non-nil")
	}
	assertPrimitiveElement(t, "GivenElement[0]", name.GivenElement[0], "http://example.com/ext/preferred")
	if name.GivenElement[1] != nil {
		t.Errorf("GivenElement[1]: expected nil (no extension), got %+v", name.GivenElement[1])
	}

	// --- communication[0].preferred (nested backbone bool) ---
	if len(patient.Communication) == 0 {
		t.Fatal("Communication slice is empty")
	}
	comm := patient.Communication[0]
	if comm.Preferred == nil || !*comm.Preferred {
		t.Error("Communication[0].Preferred: expected true")
	}
	assertPrimitiveElement(t, "Communication[0].PreferredElement", comm.PreferredElement, "http://example.com/ext/lang-source")

	// --- round-trip ---
	var patient2 fhir.Patient
	roundTrip(t, patient, &patient2)

	assertPrimitiveElement(t, "RT ActiveElement", patient2.ActiveElement, "http://example.com/ext/verified")
	assertPrimitiveElement(t, "RT GenderElement", patient2.GenderElement, "http://example.com/ext/gender-identity")
	assertPrimitiveElement(t, "RT BirthDateElement", patient2.BirthDateElement, "http://example.com/ext/birth-time")

	if len(patient2.Name) == 0 || len(patient2.Name[0].GivenElement) != 2 {
		t.Fatal("RT GivenElement: array lost after round-trip")
	}
	assertPrimitiveElement(t, "RT GivenElement[0]", patient2.Name[0].GivenElement[0], "http://example.com/ext/preferred")
	if patient2.Name[0].GivenElement[1] != nil {
		t.Error("RT GivenElement[1]: expected nil after round-trip")
	}

	if len(patient2.Communication) == 0 {
		t.Fatal("RT Communication: slice lost after round-trip")
	}
	assertPrimitiveElement(t, "RT PreferredElement", patient2.Communication[0].PreferredElement, "http://example.com/ext/lang-source")
}

// TestPatientNoExtensionsUnaffected verifies that a plain Patient without any
// primitive extensions deserializes and serializes without regressions.
func TestPatientNoExtensionsUnaffected(t *testing.T) {
	const input = `{
		"resourceType": "Patient",
		"id": "pat-plain",
		"active": false,
		"gender": "female",
		"birthDate": "2000-01-01",
		"name": [{"family": "Doe", "given": ["Jane"]}]
	}`

	patient, err := fhir.UnmarshalPatient([]byte(input))
	if err != nil {
		t.Fatalf("UnmarshalPatient: %v", err)
	}

	if patient.ActiveElement != nil {
		t.Error("ActiveElement: expected nil for field without extension")
	}
	if patient.BirthDateElement != nil {
		t.Error("BirthDateElement: expected nil for field without extension")
	}
	if patient.GenderElement != nil {
		t.Error("GenderElement: expected nil for field without extension")
	}
	if len(patient.Name) > 0 && patient.Name[0].GivenElement != nil {
		t.Error("GivenElement: expected nil for array without extension")
	}

	// Confirm round-trip still produces the right values
	var patient2 fhir.Patient
	roundTrip(t, patient, &patient2)

	if patient2.BirthDate == nil || *patient2.BirthDate != "2000-01-01" {
		t.Errorf("RT BirthDate: got %v, want \"2000-01-01\"", patient2.BirthDate)
	}
	if patient2.Active == nil || *patient2.Active != false {
		t.Error("RT Active: got unexpected value")
	}
}

// ---------------------------------------------------------------------------
// DocumentReference
// ---------------------------------------------------------------------------

// TestDocumentReferencePrimitiveExtensions exercises a DocumentReference with
// primitive extensions on:
//   - required enum:   status / _status  (DocumentReferenceStatus, min=1)
//   - optional string: description / _description
//   - optional string: date / _date
//   - nested enum:     relatesTo[0].code / _code (DocumentRelationshipType, min=1)
func TestDocumentReferencePrimitiveExtensions(t *testing.T) {
	const input = `{
		"resourceType": "DocumentReference",
		"id": "docref-1",
		"status": "current",
		"_status": {
			"extension": [{"url": "http://example.com/ext/status-reason", "valueString": "approved"}]
		},
		"date": "2024-11-01T10:00:00Z",
		"_date": {
			"id": "date-1",
			"extension": [{"url": "http://example.com/ext/recorded-by", "valueString": "system"}]
		},
		"description": "Discharge summary",
		"_description": {
			"extension": [{"url": "http://example.com/ext/lang", "valueCode": "en"}]
		},
		"relatesTo": [{
			"code": "replaces",
			"_code": {
				"extension": [{"url": "http://example.com/ext/relation-note", "valueString": "supersedes v1"}]
			},
			"target": {"reference": "DocumentReference/docref-0"}
		}],
		"content": [{
			"attachment": {"contentType": "application/pdf", "url": "http://example.com/doc.pdf"}
		}]
	}`

	doc, err := fhir.UnmarshalDocumentReference([]byte(input))
	if err != nil {
		t.Fatalf("UnmarshalDocumentReference: %v", err)
	}

	// --- status (required enum, min=1) ---
	assertPrimitiveElement(t, "StatusElement", doc.StatusElement, "http://example.com/ext/status-reason")

	// --- date (optional string) ---
	if doc.Date == nil || *doc.Date != "2024-11-01T10:00:00Z" {
		t.Errorf("Date: got %v", doc.Date)
	}
	assertPrimitiveElement(t, "DateElement", doc.DateElement, "http://example.com/ext/recorded-by")
	if doc.DateElement.Id == nil || *doc.DateElement.Id != "date-1" {
		t.Errorf("DateElement.Id: got %v, want \"date-1\"", doc.DateElement.Id)
	}

	// --- description (optional string) ---
	if doc.Description == nil || *doc.Description != "Discharge summary" {
		t.Errorf("Description: got %v", doc.Description)
	}
	assertPrimitiveElement(t, "DescriptionElement", doc.DescriptionElement, "http://example.com/ext/lang")

	// --- relatesTo[0].code (required enum in backbone element) ---
	if len(doc.RelatesTo) == 0 {
		t.Fatal("RelatesTo slice is empty")
	}
	assertPrimitiveElement(t, "RelatesTo[0].CodeElement", doc.RelatesTo[0].CodeElement, "http://example.com/ext/relation-note")

	// --- round-trip ---
	var doc2 fhir.DocumentReference
	roundTrip(t, doc, &doc2)

	assertPrimitiveElement(t, "RT StatusElement", doc2.StatusElement, "http://example.com/ext/status-reason")
	assertPrimitiveElement(t, "RT DateElement", doc2.DateElement, "http://example.com/ext/recorded-by")
	assertPrimitiveElement(t, "RT DescriptionElement", doc2.DescriptionElement, "http://example.com/ext/lang")

	if len(doc2.RelatesTo) == 0 {
		t.Fatal("RT RelatesTo: slice lost after round-trip")
	}
	assertPrimitiveElement(t, "RT RelatesTo[0].CodeElement", doc2.RelatesTo[0].CodeElement, "http://example.com/ext/relation-note")
}

// TestDocumentReferencePartialExtensions verifies that a DocumentReference
// where only some primitive fields carry extensions is handled correctly —
// fields without extensions have nil companion elements.
func TestDocumentReferencePartialExtensions(t *testing.T) {
	const input = `{
		"resourceType": "DocumentReference",
		"status": "superseded",
		"description": "Old report",
		"_description": {
			"extension": [{"url": "http://example.com/ext/note", "valueString": "archived"}]
		},
		"content": [{
			"attachment": {"contentType": "text/plain"}
		}]
	}`

	doc, err := fhir.UnmarshalDocumentReference([]byte(input))
	if err != nil {
		t.Fatalf("UnmarshalDocumentReference: %v", err)
	}

	// status has no companion
	if doc.StatusElement != nil {
		t.Error("StatusElement: expected nil (no extension provided)")
	}
	// date has no companion
	if doc.DateElement != nil {
		t.Error("DateElement: expected nil (field absent)")
	}
	// description has companion
	assertPrimitiveElement(t, "DescriptionElement", doc.DescriptionElement, "http://example.com/ext/note")
}
