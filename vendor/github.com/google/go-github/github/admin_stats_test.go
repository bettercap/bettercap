package github

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestAdminService_GetAdminStats(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/enterprise/stats/all", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		fmt.Fprint(w, `
{
  "repos": {
    "total_repos": 212,
    "root_repos": 194,
    "fork_repos": 18,
    "org_repos": 51,
    "total_pushes": 3082,
    "total_wikis": 15
  },
  "hooks": {
    "total_hooks": 27,
    "active_hooks": 23,
    "inactive_hooks": 4
  },
  "pages": {
    "total_pages": 36
  },
  "orgs": {
    "total_orgs": 33,
    "disabled_orgs": 0,
    "total_teams": 60,
    "total_team_members": 314
  },
  "users": {
    "total_users": 254,
    "admin_users": 45,
    "suspended_users": 21
  },
  "pulls": {
    "total_pulls": 86,
    "merged_pulls": 60,
    "mergeable_pulls": 21,
    "unmergeable_pulls": 3
  },
  "issues": {
    "total_issues": 179,
    "open_issues": 83,
    "closed_issues": 96
  },
  "milestones": {
    "total_milestones": 7,
    "open_milestones": 6,
    "closed_milestones": 1
  },
  "gists": {
    "total_gists": 178,
    "private_gists": 151,
    "public_gists": 25
  },
  "comments": {
    "total_commit_comments": 6,
    "total_gist_comments": 28,
    "total_issue_comments": 366,
    "total_pull_request_comments": 30
  }
}
`)
	})

	stats, _, err := client.Admin.GetAdminStats(context.Background())
	if err != nil {
		t.Errorf("AdminService.GetAdminStats returned error: %v", err)
	}

	want := &AdminStats{
		Repos: &RepoStats{
			TotalRepos:  Int(212),
			RootRepos:   Int(194),
			ForkRepos:   Int(18),
			OrgRepos:    Int(51),
			TotalPushes: Int(3082),
			TotalWikis:  Int(15),
		},
		Hooks: &HookStats{
			TotalHooks:    Int(27),
			ActiveHooks:   Int(23),
			InactiveHooks: Int(4),
		},
		Pages: &PageStats{
			TotalPages: Int(36),
		},
		Orgs: &OrgStats{
			TotalOrgs:        Int(33),
			DisabledOrgs:     Int(0),
			TotalTeams:       Int(60),
			TotalTeamMembers: Int(314),
		},
		Users: &UserStats{
			TotalUsers:     Int(254),
			AdminUsers:     Int(45),
			SuspendedUsers: Int(21),
		},
		Pulls: &PullStats{
			TotalPulls:      Int(86),
			MergedPulls:     Int(60),
			MergablePulls:   Int(21),
			UnmergablePulls: Int(3),
		},
		Issues: &IssueStats{
			TotalIssues:  Int(179),
			OpenIssues:   Int(83),
			ClosedIssues: Int(96),
		},
		Milestones: &MilestoneStats{
			TotalMilestones:  Int(7),
			OpenMilestones:   Int(6),
			ClosedMilestones: Int(1),
		},
		Gists: &GistStats{
			TotalGists:   Int(178),
			PrivateGists: Int(151),
			PublicGists:  Int(25),
		},
		Comments: &CommentStats{
			TotalCommitComments:      Int(6),
			TotalGistComments:        Int(28),
			TotalIssueComments:       Int(366),
			TotalPullRequestComments: Int(30),
		},
	}
	if !reflect.DeepEqual(stats, want) {
		t.Errorf("AdminService.GetAdminStats returned %+v, want %+v", stats, want)
	}
}
