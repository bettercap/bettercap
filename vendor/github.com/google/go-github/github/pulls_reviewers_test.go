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

func TestRequestReviewers(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"reviewers":["octocat","googlebot"],"team_reviewers":["justice-league","injustice-league"]}`+"\n")
		testHeader(t, r, "Accept", mediaTypeTeamReviewPreview)
		fmt.Fprint(w, `{"number":1}`)
	})

	// This returns a PR, unmarshalling of which is tested elsewhere
	pull, _, err := client.PullRequests.RequestReviewers(context.Background(), "o", "r", 1, ReviewersRequest{Reviewers: []string{"octocat", "googlebot"}, TeamReviewers: []string{"justice-league", "injustice-league"}})
	if err != nil {
		t.Errorf("PullRequests.RequestReviewers returned error: %v", err)
	}
	want := &PullRequest{Number: Int(1)}
	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.RequestReviewers returned %+v, want %+v", pull, want)
	}
}

func TestRemoveReviewers(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeTeamReviewPreview)
		testBody(t, r, `{"reviewers":["octocat","googlebot"],"team_reviewers":["justice-league"]}`+"\n")
	})

	_, err := client.PullRequests.RemoveReviewers(context.Background(), "o", "r", 1, ReviewersRequest{Reviewers: []string{"octocat", "googlebot"}, TeamReviewers: []string{"justice-league"}})
	if err != nil {
		t.Errorf("PullRequests.RemoveReviewers returned error: %v", err)
	}
}

func TestListReviewers(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeTeamReviewPreview)
		fmt.Fprint(w, `{"users":[{"login":"octocat","id":1}],"teams":[{"id":1,"name":"Justice League"}]}`)
	})

	reviewers, _, err := client.PullRequests.ListReviewers(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("PullRequests.ListReviewers returned error: %v", err)
	}

	want := &Reviewers{
		Users: []*User{
			{
				Login: String("octocat"),
				ID:    Int64(1),
			},
		},
		Teams: []*Team{
			{
				ID:   Int64(1),
				Name: String("Justice League"),
			},
		},
	}
	if !reflect.DeepEqual(reviewers, want) {
		t.Errorf("PullRequests.ListReviewers returned %+v, want %+v", reviewers, want)
	}
}

func TestListReviewers_withOptions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page": "2",
		})
		fmt.Fprint(w, `{}`)
	})

	_, _, err := client.PullRequests.ListReviewers(context.Background(), "o", "r", 1, &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("PullRequests.ListReviewers returned error: %v", err)
	}
}
