package main

import (
	"github.com/google/uuid"
)

type ReposConfig struct {
	OrganizationUrl *string `yaml:"organizationUrl"`
	Project *string		`yaml:"project"`
	Repositories []Repo
}

type Repo struct {
	Name *string		`yaml:"name"`
	Branches []string	`yaml:"branches"`
}

type SettingsMinNrReviewers struct {
	AllowDownvotes			bool
	BlockLastPusherVote		bool
	CreatorVoteCounts 		bool
	MinimumApproverCount 	int
	ResetOnSourcePush		bool
	Scope 					[]Scope
}

type SettingsWorkItemLinking struct {
	Scope 					[]Scope
}

type SettingsCommentRequirments struct {
	Scope 					[]Scope
}

type SettingsRequireMergeStrategy struct {
	AllowNoFastForward		bool
	AllowRebase 			bool
	AllowRebaseMerge 		bool
	AllowSquash 			bool
	Scope 					[]Scope
}

type SettingsRequiredReviewers struct {
	CreatorVoteCounts 		bool
	MinimumApproverCount	int
	Message 				string
	RequiredReviewerIds 	[]string
	Scope 					[]Scope
}

type Scope struct {
	RepositoryId 	uuid.UUID `json:"repositoryId"`
	RefName 		string		`json:"refName"`
	MatchKind 		string		`json:"matchKind"`
}

type PolicyRepoIdAndBranch struct {
	RepoId		uuid.UUID
	Branches	[]string
}