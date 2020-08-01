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
	Name *string									`yaml:"name"`
	Branches []string								`yaml:"branches"`
	BranchPoliciesSettings SettingsPolicy	`yaml:"branchPoliciesSettings"`
}

type SettingsBuild struct {
	ManualQueueOnly			 bool
	QueueOnSourceUpdateOnly	 bool
}

type Scope struct {
	RepositoryId 	uuid.UUID `json:"repositoryId"`
	RefName 		string		`json:"refName"`
	MatchKind 		string		`json:"matchKind"`
}

type SettingsPolicy struct {
	AllowDownvotes			bool
	BlockLastPusherVote		bool
	CreatorVoteCounts 		bool
	MinimumApproverCount 	int
	ResetOnSourcePush		bool
	AllowNoFastForward		bool
	AllowRebase 			bool
	AllowRebaseMerge 		bool
	AllowSquash 			bool
	Message 				string
	RequiredReviewerIds 	[]string
	Scope 					[]Scope
}

type PolicyRepoIdAndBranch struct {
	RepoId		uuid.UUID
	Branches	[]string
}