// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestOrganizationsService_ListOutsideCollaborators(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/outside_collaborators", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"filter": "2fa_disabled",
			"page":   "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOutsideCollaboratorsOptions{
		Filter:      "2fa_disabled",
		ListOptions: ListOptions{Page: 2},
	}
	members, _, err := client.Organizations.ListOutsideCollaborators(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Organizations.ListOutsideCollaborators returned error: %v", err)
	}

	want := []*User{{ID: Int64(1)}}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("Organizations.ListOutsideCollaborators returned %+v, want %+v", members, want)
	}
}

func TestOrganizationsService_ListOutsideCollaborators_invalidOrg(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Organizations.ListOutsideCollaborators(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestOrganizationsService_RemoveOutsideCollaborator(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	}
	mux.HandleFunc("/orgs/o/outside_collaborators/u", handler)

	_, err := client.Organizations.RemoveOutsideCollaborator(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.RemoveOutsideCollaborator returned error: %v", err)
	}
}

func TestOrganizationsService_RemoveOutsideCollaborator_NonMember(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNotFound)
	}
	mux.HandleFunc("/orgs/o/outside_collaborators/u", handler)

	_, err := client.Organizations.RemoveOutsideCollaborator(context.Background(), "o", "u")
	if err, ok := err.(*ErrorResponse); !ok {
		t.Errorf("Organizations.RemoveOutsideCollaborator did not return an error")
	} else if err.Response.StatusCode != http.StatusNotFound {
		t.Errorf("Organizations.RemoveOutsideCollaborator did not return 404 status code")
	}
}

func TestOrganizationsService_RemoveOutsideCollaborator_Member(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	mux.HandleFunc("/orgs/o/outside_collaborators/u", handler)

	_, err := client.Organizations.RemoveOutsideCollaborator(context.Background(), "o", "u")
	if err, ok := err.(*ErrorResponse); !ok {
		t.Errorf("Organizations.RemoveOutsideCollaborator did not return an error")
	} else if err.Response.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Organizations.RemoveOutsideCollaborator did not return 422 status code")
	}
}

func TestOrganizationsService_ConvertMemberToOutsideCollaborator(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
	}
	mux.HandleFunc("/orgs/o/outside_collaborators/u", handler)

	_, err := client.Organizations.ConvertMemberToOutsideCollaborator(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.ConvertMemberToOutsideCollaborator returned error: %v", err)
	}
}

func TestOrganizationsService_ConvertMemberToOutsideCollaborator_NonMemberOrLastOwner(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusForbidden)
	}
	mux.HandleFunc("/orgs/o/outside_collaborators/u", handler)

	_, err := client.Organizations.ConvertMemberToOutsideCollaborator(context.Background(), "o", "u")
	if err, ok := err.(*ErrorResponse); !ok {
		t.Errorf("Organizations.ConvertMemberToOutsideCollaborator did not return an error")
	} else if err.Response.StatusCode != http.StatusForbidden {
		t.Errorf("Organizations.ConvertMemberToOutsideCollaborator did not return 403 status code")
	}
}
