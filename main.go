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

	//WARNING: Dont use go get for this library use git
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"runtime"
	"sync"
	"time"
	//"sync"
	"log"
)

func main() {
	var wg sync.WaitGroup

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
		wg.Add(len(repoids))
		for _, needs := range repoids {
			go CreateMinReviewersPolicy(policyClient,ctx,needs.RepoId,projname,needs.Branches,&wg)
		}
		wg.Wait()
		//minnrofreviewersuuid, err := uuid.Parse(MIN_NUMBER_OF_REWIERES_UUID)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//minnrofreviewerdn := MIN_NUMBER_OF_REWIERES_DISPLAY_NAME
		//commentrequuid, err := uuid.Parse("c6a1889d-b943-4856-b76f-9e46bb6b0df2")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//commentreqdn := "Comment requirements"
		//requireviewersuuid, err := uuid.Parse("fd2167ab-b0be-447a-8ec8-39368250530e")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//
		//requireviewersdn := "Required reviewers"
		//requiremergestrategyuuid, err := uuid.Parse("fa4e907d-c16b-4a4c-9dfa-4916e5d171ab")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//requireviewersdn := "Require a merge strategy"
		//workitemlinkinguuid, err := uuid.Parse("40e92b44-2fe1-4dd6-b3d8-74a9c21d0c6e")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//workitemlinkdn := "Work item linking"

		//repositoryid, err := uuid.Parse("f4da5e90-de5b-4ffa-99e5-fc4dfc36f6ae")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//t := time.Now().Format("2020-07-30T11:19:36.4966665")
		t, err := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(t.String())
		//for _, repoid := range repoids {
		//	if repoid != "" {
		//		repid, err := uuid.Parse(repoid)
		//		if err != nil {
		//			log.Fatal(err)
		//		}
		//
		//		pol, err := policyClient.CreatePolicyConfiguration(ctx, policy.CreatePolicyConfigurationArgs{
		//			Configuration: &policy.PolicyConfiguration{
		//				Type: &policy.PolicyTypeRef{
		//					DisplayName: &minnrofreviewerdn,
		//					Id: &minnrofreviewersuuid,
		//				},
		//				CreatedDate: &azuredevops.Time{Time: t},
		//				IsBlocking:  &isblocking,
		//				IsDeleted:   &isdeleted,
		//				IsEnabled:   &isenabled,
		//				Settings: SettingsMinNrReviewers{
		//					AllowDownvotes:       false,
		//					BlockLastPusherVote:  true,
		//					CreatorVoteCounts:    false,
		//					//AllowNoFastForward:   "false",
		//					//AllowRebaseMerge:	  "false",
		//					//AllowRebase:          "true",
		//					//AllowSquash:          "true",
		//					//RequiredReviewerIds:  []string{"4d49214c-c791-6e27-9d74-bcce48230683"},
		//					MinimumApproverCount: 1,
		//					ResetOnSourcePush:    true,
		//					Scope: []Scope{{
		//						RepositoryId: repid,
		//						RefName:      "refs/heads/dev",
		//						MatchKind:    "exact",
		//					}},
		//				},
		//			},
		//			Project: projname,
		//		})
		//		if err != nil {
		//			log.Fatal(err)
		//		}
		//		fmt.Printf("policy name is %v\n", pol)
		//	}
		//}
	}else {
		log.Fatalln("Please specify a config file for the repositories with the argument --file")
	}

}

func RemoveIndex(s []Repo, index int) []Repo {
	return append(s[:index], s[index+1:]...)
}
