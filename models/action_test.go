package models

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"code.gitea.io/gitea/modules/setting"

	"github.com/stretchr/testify/assert"
)

func TestAction_GetRepoPath(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	repo := AssertExistsAndLoadBean(t, &Repository{}).(*Repository)
	owner := AssertExistsAndLoadBean(t, &User{ID: repo.OwnerID}).(*User)
	action := &Action{RepoID: repo.ID}
	assert.Equal(t, path.Join(owner.Name, repo.Name), action.GetRepoPath())
}

func TestAction_GetRepoLink(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	repo := AssertExistsAndLoadBean(t, &Repository{}).(*Repository)
	owner := AssertExistsAndLoadBean(t, &User{ID: repo.OwnerID}).(*User)
	action := &Action{RepoID: repo.ID}
	setting.AppSubURL = "/suburl/"
	expected := path.Join(setting.AppSubURL, owner.Name, repo.Name)
	assert.Equal(t, expected, action.GetRepoLink())
}

func TestNewRepoAction(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	repo := AssertExistsAndLoadBean(t, &Repository{OwnerID: user.ID}).(*Repository)
	repo.Owner = user

	actionBean := &Action{
		OpType:    ActionCreateRepo,
		ActUserID: user.ID,
		RepoID:    repo.ID,
		ActUser:   user,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
	}

	AssertNotExistsBean(t, actionBean)
	assert.NoError(t, NewRepoAction(user, repo))
	AssertExistsAndLoadBean(t, actionBean)
	CheckConsistencyFor(t, &Action{})
}

func TestRenameRepoAction(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	repo := AssertExistsAndLoadBean(t, &Repository{OwnerID: user.ID}).(*Repository)
	repo.Owner = user

	oldRepoName := repo.Name
	const newRepoName = "newRepoName"
	repo.Name = newRepoName
	repo.LowerName = strings.ToLower(newRepoName)

	actionBean := &Action{
		OpType:    ActionRenameRepo,
		ActUserID: user.ID,
		ActUser:   user,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		Content:   oldRepoName,
	}
	AssertNotExistsBean(t, actionBean)
	assert.NoError(t, RenameRepoAction(user, oldRepoName, repo))
	AssertExistsAndLoadBean(t, actionBean)

	_, err := x.ID(repo.ID).Cols("name", "lower_name").Update(repo)
	assert.NoError(t, err)
	CheckConsistencyFor(t, &Action{})
}

func TestPushCommits_ToAPIPayloadCommits(t *testing.T) {
	pushCommits := NewPushCommits()
	pushCommits.Commits = []*PushCommit{
		{
			Sha1:           "69554a6",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User2",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User2",
			Message:        "not signed commit",
		},
		{
			Sha1:           "27566bd",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User2",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User2",
			Message:        "good signed commit (with not yet validated email)",
		},
		{
			Sha1:           "5099b81",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User2",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User2",
			Message:        "good signed commit",
		},
	}
	pushCommits.Len = len(pushCommits.Commits)

	repo := AssertExistsAndLoadBean(t, &Repository{ID: 16}).(*Repository)
	payloadCommits, err := pushCommits.ToAPIPayloadCommits(repo.RepoPath(), "/user2/repo16")
	assert.NoError(t, err)
	assert.EqualValues(t, 3, len(payloadCommits))

	assert.Equal(t, "69554a6", payloadCommits[0].ID)
	assert.Equal(t, "not signed commit", payloadCommits[0].Message)
	assert.Equal(t, "/user2/repo16/commit/69554a6", payloadCommits[0].URL)
	assert.Equal(t, "User2", payloadCommits[0].Committer.Name)
	assert.Equal(t, "user2", payloadCommits[0].Committer.UserName)
	assert.Equal(t, "User2", payloadCommits[0].Author.Name)
	assert.Equal(t, "user2", payloadCommits[0].Author.UserName)
	assert.EqualValues(t, []string{}, payloadCommits[0].Added)
	assert.EqualValues(t, []string{}, payloadCommits[0].Removed)
	assert.EqualValues(t, []string{"readme.md"}, payloadCommits[0].Modified)

	assert.Equal(t, "27566bd", payloadCommits[1].ID)
	assert.Equal(t, "good signed commit (with not yet validated email)", payloadCommits[1].Message)
	assert.Equal(t, "/user2/repo16/commit/27566bd", payloadCommits[1].URL)
	assert.Equal(t, "User2", payloadCommits[1].Committer.Name)
	assert.Equal(t, "user2", payloadCommits[1].Committer.UserName)
	assert.Equal(t, "User2", payloadCommits[1].Author.Name)
	assert.Equal(t, "user2", payloadCommits[1].Author.UserName)
	assert.EqualValues(t, []string{}, payloadCommits[1].Added)
	assert.EqualValues(t, []string{}, payloadCommits[1].Removed)
	assert.EqualValues(t, []string{"readme.md"}, payloadCommits[1].Modified)

	assert.Equal(t, "5099b81", payloadCommits[2].ID)
	assert.Equal(t, "good signed commit", payloadCommits[2].Message)
	assert.Equal(t, "/user2/repo16/commit/5099b81", payloadCommits[2].URL)
	assert.Equal(t, "User2", payloadCommits[2].Committer.Name)
	assert.Equal(t, "user2", payloadCommits[2].Committer.UserName)
	assert.Equal(t, "User2", payloadCommits[2].Author.Name)
	assert.Equal(t, "user2", payloadCommits[2].Author.UserName)
	assert.EqualValues(t, []string{"readme.md"}, payloadCommits[2].Added)
	assert.EqualValues(t, []string{}, payloadCommits[2].Removed)
	assert.EqualValues(t, []string{}, payloadCommits[2].Modified)
}

func TestPushCommits_AvatarLink(t *testing.T) {
	pushCommits := NewPushCommits()
	pushCommits.Commits = []*PushCommit{
		{
			Sha1:           "abcdef1",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user4@example.com",
			AuthorName:     "User Four",
			Message:        "message1",
		},
		{
			Sha1:           "abcdef2",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User Two",
			Message:        "message2",
		},
	}
	pushCommits.Len = len(pushCommits.Commits)

	assert.Equal(t,
		"https://secure.gravatar.com/avatar/ab53a2911ddf9b4817ac01ddcd3d975f?d=identicon",
		pushCommits.AvatarLink("user2@example.com"))

	assert.Equal(t,
		"https://secure.gravatar.com/avatar/19ade630b94e1e0535b3df7387434154?d=identicon",
		pushCommits.AvatarLink("nonexistent@example.com"))
}

func TestRegExp_issueReferenceKeywordsPat(t *testing.T) {
	trueTestCases := []string{
		"#2",
		"[#2]",
		"please see go-gitea/gitea#5",
		"#2:",
	}
	falseTestCases := []string{
		"kb#2",
		"#2xy",
	}

	for _, testCase := range trueTestCases {
		assert.True(t, issueReferenceKeywordsPat.MatchString(testCase))
	}
	for _, testCase := range falseTestCases {
		assert.False(t, issueReferenceKeywordsPat.MatchString(testCase))
	}
}

func Test_getIssueFromRef(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	repo := AssertExistsAndLoadBean(t, &Repository{ID: 1}).(*Repository)
	for _, test := range []struct {
		Ref             string
		ExpectedIssueID int64
	}{
		{"#2", 2},
		{"reopen #2", 2},
		{"user2/repo2#1", 4},
		{"fixes user2/repo2#1", 4},
		{"fixes: user2/repo2#1", 4},
	} {
		issue, err := getIssueFromRef(repo, test.Ref)
		assert.NoError(t, err)
		if assert.NotNil(t, issue) {
			assert.EqualValues(t, test.ExpectedIssueID, issue.ID)
		}
	}

	for _, badRef := range []string{
		"doesnotexist/doesnotexist#1",
		fmt.Sprintf("#%d", NonexistentID),
	} {
		issue, err := getIssueFromRef(repo, badRef)
		assert.NoError(t, err)
		assert.Nil(t, issue)
	}
}

func TestUpdateIssuesCommit(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	pushCommits := []*PushCommit{
		{
			Sha1:           "abcdef1",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user4@example.com",
			AuthorName:     "User Four",
			Message:        "start working on #FST-1, #1",
		},
		{
			Sha1:           "abcdef2",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User Two",
			Message:        "a plain message",
		},
		{
			Sha1:           "abcdef2",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User Two",
			Message:        "close #2",
		},
	}

	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	repo := AssertExistsAndLoadBean(t, &Repository{ID: 1}).(*Repository)
	repo.Owner = user

	commentBean := &Comment{
		Type:      CommentTypeCommitRef,
		CommitSHA: "abcdef1",
		PosterID:  user.ID,
		IssueID:   1,
	}
	issueBean := &Issue{RepoID: repo.ID, Index: 2}

	AssertNotExistsBean(t, commentBean)
	AssertNotExistsBean(t, &Issue{RepoID: repo.ID, Index: 2}, "is_closed=1")
	assert.NoError(t, UpdateIssuesCommit(user, repo, pushCommits, repo.DefaultBranch))
	AssertExistsAndLoadBean(t, commentBean)
	AssertExistsAndLoadBean(t, issueBean, "is_closed=1")
	CheckConsistencyFor(t, &Action{})

	// Test that push to a non-default branch closes no issue.
	pushCommits = []*PushCommit{
		{
			Sha1:           "abcdef1",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user4@example.com",
			AuthorName:     "User Four",
			Message:        "close #1",
		},
	}
	repo = AssertExistsAndLoadBean(t, &Repository{ID: 3}).(*Repository)
	commentBean = &Comment{
		Type:      CommentTypeCommitRef,
		CommitSHA: "abcdef1",
		PosterID:  user.ID,
		IssueID:   6,
	}
	issueBean = &Issue{RepoID: repo.ID, Index: 1}

	AssertNotExistsBean(t, commentBean)
	AssertNotExistsBean(t, &Issue{RepoID: repo.ID, Index: 1}, "is_closed=1")
	assert.NoError(t, UpdateIssuesCommit(user, repo, pushCommits, "non-existing-branch"))
	AssertExistsAndLoadBean(t, commentBean)
	AssertNotExistsBean(t, issueBean, "is_closed=1")
	CheckConsistencyFor(t, &Action{})
}

func TestUpdateIssuesCommit_Colon(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	pushCommits := []*PushCommit{
		{
			Sha1:           "abcdef2",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User Two",
			Message:        "close: #2",
		},
	}

	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	repo := AssertExistsAndLoadBean(t, &Repository{ID: 1}).(*Repository)
	repo.Owner = user

	issueBean := &Issue{RepoID: repo.ID, Index: 2}

	AssertNotExistsBean(t, &Issue{RepoID: repo.ID, Index: 2}, "is_closed=1")
	assert.NoError(t, UpdateIssuesCommit(user, repo, pushCommits, repo.DefaultBranch))
	AssertExistsAndLoadBean(t, issueBean, "is_closed=1")
	CheckConsistencyFor(t, &Action{})
}

func TestUpdateIssuesCommit_Issue5957(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)

	// Test that push to a non-default branch closes an issue.
	pushCommits := []*PushCommit{
		{
			Sha1:           "abcdef1",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user4@example.com",
			AuthorName:     "User Four",
			Message:        "close #2",
		},
	}

	repo := AssertExistsAndLoadBean(t, &Repository{ID: 2}).(*Repository)
	commentBean := &Comment{
		Type:      CommentTypeCommitRef,
		CommitSHA: "abcdef1",
		PosterID:  user.ID,
		IssueID:   7,
	}

	issueBean := &Issue{RepoID: repo.ID, Index: 2, ID: 7}

	AssertNotExistsBean(t, commentBean)
	AssertNotExistsBean(t, issueBean, "is_closed=1")
	assert.NoError(t, UpdateIssuesCommit(user, repo, pushCommits, "non-existing-branch"))
	AssertExistsAndLoadBean(t, commentBean)
	AssertExistsAndLoadBean(t, issueBean, "is_closed=1")
	CheckConsistencyFor(t, &Action{})
}

func TestUpdateIssuesCommit_AnotherRepo(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)

	// Test that a push to default branch closes issue in another repo
	// If the user also has push permissions to that repo
	pushCommits := []*PushCommit{
		{
			Sha1:           "abcdef1",
			CommitterEmail: "user2@example.com",
			CommitterName:  "User Two",
			AuthorEmail:    "user2@example.com",
			AuthorName:     "User Two",
			Message:        "close user2/repo1#1",
		},
	}

	repo := AssertExistsAndLoadBean(t, &Repository{ID: 2}).(*Repository)
	commentBean := &Comment{
		Type:      CommentTypeCommitRef,
		CommitSHA: "abcdef1",
		PosterID:  user.ID,
		IssueID:   1,
	}

	issueBean := &Issue{RepoID: 1, Index: 1, ID: 1}

	AssertNotExistsBean(t, commentBean)
	AssertNotExistsBean(t, issueBean, "is_closed=1")
	assert.NoError(t, UpdateIssuesCommit(user, repo, pushCommits, repo.DefaultBranch))
	AssertExistsAndLoadBean(t, commentBean)
	AssertExistsAndLoadBean(t, issueBean, "is_closed=1")
	CheckConsistencyFor(t, &Action{})
}

func TestUpdateIssuesCommit_AnotherRepoNoPermission(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	user := AssertExistsAndLoadBean(t, &User{ID: 10}).(*User)

	// Test that a push with close reference *can not* close issue
	// If the commiter doesn't have push rights in that repo
	pushCommits := []*PushCommit{
		{
			Sha1:           "abcdef3",
			CommitterEmail: "user10@example.com",
			CommitterName:  "User Ten",
			AuthorEmail:    "user10@example.com",
			AuthorName:     "User Ten",
			Message:        "close user3/repo3#1",
		},
	}

	repo := AssertExistsAndLoadBean(t, &Repository{ID: 6}).(*Repository)
	commentBean := &Comment{
		Type:      CommentTypeCommitRef,
		CommitSHA: "abcdef3",
		PosterID:  user.ID,
		IssueID:   6,
	}

	issueBean := &Issue{RepoID: 3, Index: 1, ID: 6}

	AssertNotExistsBean(t, commentBean)
	AssertNotExistsBean(t, issueBean, "is_closed=1")
	assert.NoError(t, UpdateIssuesCommit(user, repo, pushCommits, repo.DefaultBranch))
	AssertExistsAndLoadBean(t, commentBean)
	AssertNotExistsBean(t, issueBean, "is_closed=1")
	CheckConsistencyFor(t, &Action{})
}

func TestTransferRepoAction(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	user2 := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	user4 := AssertExistsAndLoadBean(t, &User{ID: 4}).(*User)
	repo := AssertExistsAndLoadBean(t, &Repository{ID: 1, OwnerID: user2.ID}).(*Repository)

	repo.OwnerID = user4.ID
	repo.Owner = user4

	actionBean := &Action{
		OpType:    ActionTransferRepo,
		ActUserID: user2.ID,
		ActUser:   user2,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
	}
	AssertNotExistsBean(t, actionBean)
	assert.NoError(t, TransferRepoAction(user2, user2, repo))
	AssertExistsAndLoadBean(t, actionBean)

	_, err := x.ID(repo.ID).Cols("owner_id").Update(repo)
	assert.NoError(t, err)
	CheckConsistencyFor(t, &Action{})
}

func TestMergePullRequestAction(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	repo := AssertExistsAndLoadBean(t, &Repository{ID: 1, OwnerID: user.ID}).(*Repository)
	repo.Owner = user
	issue := AssertExistsAndLoadBean(t, &Issue{ID: 3, RepoID: repo.ID}).(*Issue)

	actionBean := &Action{
		OpType:    ActionMergePullRequest,
		ActUserID: user.ID,
		ActUser:   user,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
	}
	AssertNotExistsBean(t, actionBean)
	assert.NoError(t, MergePullRequestAction(user, repo, issue))
	AssertExistsAndLoadBean(t, actionBean)
	CheckConsistencyFor(t, &Action{})
}

func TestGetFeeds(t *testing.T) {
	// test with an individual user
	assert.NoError(t, PrepareTestDatabase())
	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)

	actions, err := GetFeeds(GetFeedsOptions{
		RequestedUser:    user,
		RequestingUserID: user.ID,
		IncludePrivate:   true,
		OnlyPerformedBy:  false,
		IncludeDeleted:   true,
	})
	assert.NoError(t, err)
	if assert.Len(t, actions, 1) {
		assert.EqualValues(t, 1, actions[0].ID)
		assert.EqualValues(t, user.ID, actions[0].UserID)
	}

	actions, err = GetFeeds(GetFeedsOptions{
		RequestedUser:    user,
		RequestingUserID: user.ID,
		IncludePrivate:   false,
		OnlyPerformedBy:  false,
	})
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
}

func TestGetFeeds2(t *testing.T) {
	// test with an organization user
	assert.NoError(t, PrepareTestDatabase())
	org := AssertExistsAndLoadBean(t, &User{ID: 3}).(*User)
	const userID = 2 // user2 is an owner of the organization

	actions, err := GetFeeds(GetFeedsOptions{
		RequestedUser:    org,
		RequestingUserID: userID,
		IncludePrivate:   true,
		OnlyPerformedBy:  false,
		IncludeDeleted:   true,
	})
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	if assert.Len(t, actions, 1) {
		assert.EqualValues(t, 2, actions[0].ID)
		assert.EqualValues(t, org.ID, actions[0].UserID)
	}

	actions, err = GetFeeds(GetFeedsOptions{
		RequestedUser:    org,
		RequestingUserID: userID,
		IncludePrivate:   false,
		OnlyPerformedBy:  false,
		IncludeDeleted:   true,
	})
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
}
