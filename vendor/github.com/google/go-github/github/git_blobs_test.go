// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGitService_GetBlob(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/git/blobs/s", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeGraphQLNodeIDPreview)

		fmt.Fprint(w, `{
			  "sha": "s",
			  "content": "blob content"
			}`)
	})

	blob, _, err := client.Git.GetBlob(context.Background(), "o", "r", "s")
	if err != nil {
		t.Errorf("Git.GetBlob returned error: %v", err)
	}

	want := Blob{
		SHA:     String("s"),
		Content: String("blob content"),
	}

	if !reflect.DeepEqual(*blob, want) {
		t.Errorf("Blob.Get returned %+v, want %+v", *blob, want)
	}
}

func TestGitService_GetBlob_invalidOwner(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Git.GetBlob(context.Background(), "%", "%", "%")
	testURLParseError(t, err)
}

func TestGitService_CreateBlob(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	input := &Blob{
		SHA:      String("s"),
		Content:  String("blob content"),
		Encoding: String("utf-8"),
		Size:     Int(12),
	}

	mux.HandleFunc("/repos/o/r/git/blobs", func(w http.ResponseWriter, r *http.Request) {
		v := new(Blob)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeGraphQLNodeIDPreview)

		want := input
		if !reflect.DeepEqual(v, want) {
			t.Errorf("Git.CreateBlob request body: %+v, want %+v", v, want)
		}

		fmt.Fprint(w, `{
		 "sha": "s",
		 "content": "blob content",
		 "encoding": "utf-8",
		 "size": 12
		}`)
	})

	blob, _, err := client.Git.CreateBlob(context.Background(), "o", "r", input)
	if err != nil {
		t.Errorf("Git.CreateBlob returned error: %v", err)
	}

	want := input

	if !reflect.DeepEqual(*blob, *want) {
		t.Errorf("Git.CreateBlob returned %+v, want %+v", *blob, *want)
	}
}

func TestGitService_CreateBlob_invalidOwner(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	_, _, err := client.Git.CreateBlob(context.Background(), "%", "%", &Blob{})
	testURLParseError(t, err)
}
