package main


//For azure-devops-go-api library please clone the repo into src/github.com/microsoft/ and change branch
//from default to azuredevops/v5.1.0-b1. The main branch at the time of writing this client didn't contain
//the changes to fix time issue when creating the branch policies for azure repos
import (
	"context"
	"flag"
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
	"os"
	"path/filepath"

	//"golang.org/x/crypto/ssh/terminal"
	//WARNING: Dont use go get for this library use git
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"log"
	"runtime"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var wf sync.WaitGroup
	runtime.GOMAXPROCS(4)
	reposFile := flag.String("file", "", "Add a config file yaml with all the pipelines contains")
	user := flag.String("user", "", "Add username(email) that you want to connect with")
	password := flag.String("pass", "", "Add the password for username(email) for provided username")
	flag.Parse()
	if *reposFile != "" {
		var r ReposConfig
		var p CreatedPolicies
		repositories := r.getConf(reposFile).Repositories
		branchpolicies := r.getConf(reposFile).BranchPoliciesSettings
		organizationUrl := r.getConf(reposFile).OrganizationUrl
		projname := r.getConf(reposFile).Project
		username := ""
		if *user != "" {
			username = *user //"mariuss2007@gmail.com"
		} else {
			log.Fatalln("Please provide a username to connect to azure repos using the argument: --user")
		}
		personalAccessToken := ""
		if *password != "" {
			personalAccessToken = *password
		} else {
			log.Fatalln("Please provide the PAT for the azure devops account you want to use")
		}
		// Create a connection to your organization
		connection := azuredevops.NewPatConnection(*organizationUrl, personalAccessToken)

		ctx := context.Background()

		// Create a client to interact with the Git area
		gitClient, err := git.NewClient(ctx, connection)
		if err != nil {
			log.Fatal(err)
		}
		// Get All existing repos for comparation
		existingrepos := GetAllRepos(gitClient,ctx,projname)
		oldobjectid := "0000000000000000000000000000000000000000"
		islocked := false
		repos := []Repo{}
		createdrepobranches := map[string][]string{}
		for _, repo := range repositories {
			if Find(existingrepos,*repo.Name) == false {
					repos = append(repos,repo)
			} else {
				wg.Add(1)
				go func() {
					for k,v := range GetCreatedReposBranches(gitClient,ctx,projname,repo.Name,createdrepobranches) {
						for _, branch := range *repo.Branches {
							if Find(v,branch) == false {
								//CreateBranches(gitClient,ctx,projname,repo.Name,GetCommitIdBranch(gitClient,ctx,projname,repo.Name),*branches)
								branchresponse, err := gitClient.UpdateRefs(ctx, git.UpdateRefsArgs{
									RefUpdates: &[]git.GitRefUpdate{{
										IsLocked:    &islocked,
										Name:        &branch,
										OldObjectId: &oldobjectid,
										NewObjectId: GetCommitIdBranch(gitClient,ctx,projname,&k),
									},
									},
									RepositoryId: &k,
									Project:      projname,
								})
								if err != nil {
									log.Fatalf("There was some error creating the tag %v", err)
								}
								for _, branch := range (*branchresponse) {
									fmt.Printf("The branch %s from  repo %v was created and it's success status is  %v\n", *branch.Name,*branch.RepositoryId, *branch.UpdateStatus)
								}
							}
						}
					}
					defer wg.Done()
				}()
			}
			wg.Wait()
		}
		//for _, repo := range repositories {
		//	for k, v := range createdrepobranches {
		//		for _, branch := range *repo.Branches {
		//			if Find(v,branch) == false {
		//				//CreateBranches(gitClient,ctx,projname,repo.Name,GetCommitIdBranch(gitClient,ctx,projname,repo.Name),*branches)
		//				branchresponse, err := gitClient.UpdateRefs(ctx, git.UpdateRefsArgs{
		//					RefUpdates: &[]git.GitRefUpdate{{
		//						IsLocked:    &islocked,
		//						Name:        &branch,
		//						OldObjectId: &oldobjectid,
		//						NewObjectId: GetCommitIdBranch(gitClient,ctx,projname,&k),
		//					},
		//					},
		//					RepositoryId: &k,
		//					Project:      projname,
		//				})
		//				if err != nil {
		//					log.Fatalf("There was some error creating the tag %v", err)
		//				}
		//				for _, branch := range (*branchresponse) {
		//					fmt.Printf("The branch %s from  repo %v was created and it's success status is  %v\n", *branch.Name,*branch.RepositoryId, *branch.UpdateStatus)
		//				}
		//			}
		//		}
		//	}
		//}
		//os.Exit(1)
		repoids := make([]PolicyRepoIdAndBranch,len(repos))
		reposLength := len(repos)
		wg.Add(reposLength)
		for i, repo := range repos {
			go CreateRepos(gitClient, ctx, username, personalAccessToken, projname, repo.Name, repo.Branches,repoids,i,&wg)
		}
		wg.Wait()
		policyClient, err := policy.NewClient(ctx, connection)
		if err != nil {
			log.Fatal(err)
		}
		for _, needs := range repoids {
			wg.Add(len(needs.Branches) * 5) //5 represents the number of times we run CreateBranchPolicy for differents policies types
			for _, branch := range needs.Branches {
				settings := SettingsPolicy{
					MinimumApproverCount: 	branchpolicies.MinimumApproverCount,
					AllowDownvotes:			branchpolicies.AllowDownvotes,
					BlockLastPusherVote: 	branchpolicies.BlockLastPusherVote,
					CreatorVoteCounts:		branchpolicies.CreatorVoteCounts,
					ResetOnSourcePush:		branchpolicies.ResetOnSourcePush,
					AllowNoFastForward:		branchpolicies.AllowNoFastForward,
					AllowRebase:			branchpolicies.AllowRebase,
					AllowRebaseMerge: 		branchpolicies.AllowRebaseMerge,
					AllowSquash: 			branchpolicies.AllowSquash,
					RequiredReviewerIds: 	branchpolicies.RequiredReviewerIds,
					Scope: []Scope{{
						RepositoryId: needs.RepoId,
						RefName:      branch,
						MatchKind:    "exact",
					}},
				}
				go CreateBranchPolicy(policyClient, ctx,
					needs.RepoId, projname,
					needs.Branches, MIN_NUMBER_OF_REWIERES_DISPLAY_NAME,
					MIN_NUMBER_OF_REWIERES_UUID, settings, true, needs.RepoName, &wg)
				go CreateBranchPolicy(policyClient, ctx,
					needs.RepoId, projname,
					needs.Branches, WORK_ITEM_LINKING_DISPLAY_NAME,
					WORK_ITEM_LINKING_DISPLAY_UUID, settings, true, needs.RepoName, &wg)
				go CreateBranchPolicy(policyClient, ctx,
					needs.RepoId, projname,
					needs.Branches, COMMENT_REQUIREMENTS_DISPLAY_NAME,
					COMMENT_REQUIREMENTS_UUID, settings, true, needs.RepoName, &wg)
				go CreateBranchPolicy(policyClient, ctx,
					needs.RepoId, projname,
					needs.Branches, REQUIRE_A_MERGE_STRATEGY_DISPLAY_NAME,
					REQUIRE_A_MERGE_STRATEGY_UUID, settings, true, needs.RepoName, &wg)
				go CreateBranchPolicy(policyClient, ctx,
					needs.RepoId, projname,
					needs.Branches, REQUIRED_REVIEWERS_DISPLAY_NAME,
					REQUIRED_REVIEWERS_UUID, settings, false, needs.RepoName, &wg)

			}
			wg.Wait()
		}
		//baserepo := "test"
		SavePoliciesStates(username, personalAccessToken)
		//gitClient.CreatePush(ctx, git.CreatePushArgs{
		//	Push:         &git.GitPush{
		//		Links:             nil,
		//		Date:              nil,
		//		PushCorrelationId: nil,
		//		PushedBy:          nil,
		//		PushId:            nil,
		//		Url:               nil,
		//		Commits:           nil,
		//		RefUpdates:        nil,
		//		Repository:        nil,
		//	},
		//	RepositoryId: &baserepo,
		//	Project:      projname,
		//})
		var files []string
		policiesfolder := "created-policies"
		err = filepath.Walk(policiesfolder, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		})
		if err != nil {
			panic(err)
		}
		createdpolicies := []*CreatedPolicies{}
		for _, file := range files {
			if file != policiesfolder {
				createdpolicies = append(createdpolicies, p.getConf(file))
			}
		}
				for _, cpol := range createdpolicies {
					wf.Add(len(*cpol))
					for _,cpl := range *cpol {
						fmt.Println(cpl.Policyid)
						settings := SettingsPolicy{
							MinimumApproverCount: branchpolicies.MinimumApproverCount,
							AllowDownvotes:       branchpolicies.AllowDownvotes,
							BlockLastPusherVote:  branchpolicies.BlockLastPusherVote,
							CreatorVoteCounts:    branchpolicies.CreatorVoteCounts,
							ResetOnSourcePush:    branchpolicies.ResetOnSourcePush,
							AllowNoFastForward:   branchpolicies.AllowNoFastForward,
							AllowRebase:          branchpolicies.AllowRebase,
							AllowRebaseMerge:     branchpolicies.AllowRebaseMerge,
							AllowSquash:          branchpolicies.AllowSquash,
							RequiredReviewerIds:  branchpolicies.RequiredReviewerIds,
							Scope: []Scope{{
								RepositoryId: cpl.Repoid,
								RefName:      cpl.Branch,
								MatchKind:    "exact",
							}},
						}
						if cpl.Typeid == REQUIRED_REVIEWERS_UUID {
							UpdateBranchPolicy(policyClient, ctx,
								cpl.Repoid, projname,
								cpl.TypeDisplayName, cpl.Typeid,
								settings, false, &cpl.Policyid, &wf)
						} else {
							UpdateBranchPolicy(policyClient, ctx,
								cpl.Repoid, projname,
								cpl.TypeDisplayName, cpl.Typeid,
								settings, true, &cpl.Policyid, &wf)
						}
					}
					wf.Wait()
				}
	}else {
		log.Fatalln("Please specify a config file for the repositories with the argument --file")
	}
}
