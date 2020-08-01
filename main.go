package main


//For azure-devops-go-api library please clone the repo into src/github.com/microsoft/ and change branch
//from default to azuredevops/v5.1.0-b1. The main branch at the time of writing this client didn't contain
//the changes to fix time issue when creating the branch policies for azure repos
import (
	"context"
	"flag"
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
	"golang.org/x/crypto/ssh/terminal"

	//"golang.org/x/crypto/ssh/terminal"
	//WARNING: Dont use go get for this library use git
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"runtime"
	"sync"
	"log"
)

func main() {
	var wg sync.WaitGroup
	var wf sync.WaitGroup
	runtime.GOMAXPROCS(4)
	reposFile := flag.String("file", "", "Add a config file yaml with all the pipelines contains")
	user := flag.String("user", "", "Add username(email) that you want to connect with")
	flag.Parse()
	if *reposFile != "" {
		var r ReposConfig
		fmt.Printf("Please type the PAT for the current user: ")
		password, _ := terminal.ReadPassword(0)
		personalAccessToken := fmt.Sprintf("%s", password)
		repositories := r.getConf(reposFile).Repositories
		organizationUrl := r.getConf(reposFile).OrganizationUrl
		projname := r.getConf(reposFile).Project
		username := ""
		if *user != "" {
			username = *user //"mariuss2007@gmail.com"
		} else {
			log.Fatalln("Please provide a username to connect to azure repos using the argument: --user")
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
		repos := repositories
		for i, repo := range repositories {
			if Find(existingrepos,*repo.Name) == true {
					repos = append(repositories[:i], repositories[i+1:]...)
			}
		}
		fmt.Printf("This is %v with username %s",repos,username)
		repoids := make([]PolicyRepoIdAndBranch,len(repos))
		reposLength := len(repos)
		wg.Add(reposLength)
		for i, repo := range repos {
			go CreateRepos(gitClient, ctx, username, personalAccessToken, projname, repo.Name, repo.Branches,repoids,i,&wg)
		}
		wg.Wait()
		fmt.Println(repoids)
		policyClient, err := policy.NewClient(ctx, connection)
		if err != nil {
			log.Fatal(err)
		}

		wg.Add(len(repoids)*5) //5 represents the number of times we run CreateBranchPolicy for differents policies types
		for _, needs := range repoids {
			wf.Add(len(needs.Branches))
			for _,branch := range needs.Branches {
				settings := SettingsPolicy{
					AllowDownvotes:      false,
					BlockLastPusherVote: true,
					CreatorVoteCounts:   false,
					MinimumApproverCount: 1,
					ResetOnSourcePush:    	true,
					AllowNoFastForward:		false,
					AllowRebase:			true,
					AllowRebaseMerge: 		false,
					AllowSquash: 			true,
					RequiredReviewerIds: 	[]string{"4d49214c-c791-6e27-9d74-bcce48230683"},
					Scope: []Scope{{
						RepositoryId: needs.RepoId,
						RefName:      branch,
						MatchKind:    "exact",
					}},
				}
				go CreateBranchPolicy(policyClient, ctx,
									needs.RepoId, projname,
									needs.Branches,MIN_NUMBER_OF_REWIERES_DISPLAY_NAME,
									MIN_NUMBER_OF_REWIERES_UUID,settings,true, &wg)
				go CreateBranchPolicy(policyClient, ctx,
									needs.RepoId, projname,
									needs.Branches,WORK_ITEM_LINKING_DISPLAY_NAME,
									WORK_ITEM_LINKING_DISPLAY_UUID,settings,true, &wg)
				go CreateBranchPolicy(policyClient, ctx,
									needs.RepoId, projname,
									needs.Branches,COMMENT_REQUIREMENTS_DISPLAY_NAME,
									COMMENT_REQUIREMENTS_UUID,settings,true, &wg)
				go CreateBranchPolicy(policyClient, ctx,
									needs.RepoId, projname,
									needs.Branches,REQUIRE_A_MERGE_STRATEGY_DISPLAY_NAME,
									REQUIRE_A_MERGE_STRATEGY_UUID,settings,true, &wg)
				go CreateBranchPolicy(policyClient, ctx,
									needs.RepoId, projname,
									needs.Branches,REQUIRED_REVIEWERS_DISPLAY_NAME,
									REQUIRED_REVIEWERS_UUID,settings,false, &wg)
				wf.Done()
			}
			wf.Wait()
		}
		wg.Wait()
	}else {
		log.Fatalln("Please specify a config file for the repositories with the argument --file")
	}

}

func RemoveIndex(s []Repo, index int) []Repo {
	return append(s[:index], s[index+1:]...)
}
