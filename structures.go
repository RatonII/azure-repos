package main

import "github.com/google/uuid"

type ReposConfig struct {
	OrganizationUrl *string `yaml:"organizationUrl"`
	Project *string		`yaml:"project"`
	Repositories Repositories
}
type Repositories []Repo

type Repo struct {
	Name *string		`yaml:"name"`
	Branches []string	`yaml:"branches"`
}

type Settings struct {
	CreatorVoteCounts 		bool
	AllowDownvotes			bool
	RequiredReviewerIds 	[]string
	AllowNoFastForward		bool
	BlockLastPusherVote		bool
	AllowRebase 			bool
	AllowRebaseMerge 		bool
	AllowSquash 			bool
	MinimumApproverCount 	int
	ResetOnSourcePush		bool
	Scope 					[]Scope
}

type Scope struct {
	RepositoryId 	uuid.UUID
	RefName 		string
	MatchKind 		string
}
