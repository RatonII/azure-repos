package main

import (
	"github.com/google/uuid"
)

type ReposConfig struct {
	OrganizationUrl *string 				`yaml:"organizationUrl"`
	Project *string							`yaml:"project"`
	Repositories []Repo						`yaml:"repositories"`
	BranchPoliciesSettings SettingsPolicy	`yaml:"branchPoliciesSettings"`
}

type Repo struct {
	Name *string							`yaml:"name"`
	Branches *[]string						`yaml:"branches"`
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
	MinimumApproverCount 	int			`yaml:"minimumApproverCount"`
	AllowDownvotes			bool		`yaml:"allowDownvotes"`
	BlockLastPusherVote		bool		`yaml:"blockLastPusherVote"`
	CreatorVoteCounts 		bool		`yaml:"creatorVoteCounts"`
	ResetOnSourcePush		bool		`yaml:"resetOnSourcePush"`
	AllowNoFastForward		bool		`yaml:"allowNoFastForward"`
	AllowRebase 			bool		`yaml:"allowRebase"`
	AllowRebaseMerge 		bool		`yaml:"allowRebaseMerge"`
	AllowSquash 			bool		`yaml:"allowSquash"`
	RequiredReviewerIds 	[]string	`yaml:"requiredReviewerIds"`
	Message 				string
	Scope					[]Scope
}

type PolicyRepoIdAndBranch struct {
	RepoName	string
	RepoId		uuid.UUID
	Branches	[]string
}

type CreatedPolicies []struct {
	Policyid 		int			`yaml:"policyid"`
	Repoid  		uuid.UUID	`yaml:"repoid"`
	Branch 			string		`yaml:"branch"`
	Typeid 			string		`yaml:"typeid"`
	TypeDisplayName	string		`yaml:"typedisplayname"`
}
