package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"runtime"
	"sync"
	"time"
	"golang.org/x/crypto/ssh/terminal"
	//"sync"
	"log"
)



func main() {
	fmt.Printf("Please type the personal access token: ")
	password, _ := terminal.ReadPassword(0)
	personalAccessToken := fmt.Sprintf("%s", password)
	var wg sync.WaitGroup
	runtime.GOMAXPROCS(4)
	reposFile := flag.String("file", "", "Add a config file yaml with all the pipelines contains")
	flag.Parse()
	if *reposFile != "" {
		var r ReposConfig
		repositories := r.getConf(reposFile).Repositories
		organizationUrl := r.getConf(reposFile).OrganizationUrl
		projname := r.getConf(reposFile).Project
		username := "mariuss2007@gmail.com"
		//personalAccessToken := "xre5yhde563fkwpawen7nlrjilocnwmdj35znbfu6uzak26v5p4q" // todo: replace value with your PAT
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

		repoids := make([]string,len(repositories))
		for i, repos := range repositories {
			if Find(existingrepos,*repos.Name) == false {
				wg.Add(1)
				go CreateRepos(gitClient, ctx, username, personalAccessToken, projname, repos.Name, repos.Branches,repoids,i, &wg)
			wg.Wait()
			}
		}
		fmt.Println(repoids)
		//policyClient, err := policy.NewClient(ctx, connection)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//isdeleted := false
		//isenabled := true
		//isblocking := false
		//policyuuid, err := uuid.Parse("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//repositoryid, err := uuid.Parse("f4da5e90-de5b-4ffa-99e5-fc4dfc36f6ae")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//t := time.Now().Format("2020-07-30T11:19:36.4966665")
		t, err := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		if err != nil {
			log.Fatal(err)
		}
		println(t.String())
		//pol, err := policyClient.CreatePolicyConfiguration(ctx,policy.CreatePolicyConfigurationArgs{
		//	Configuration:   &policy.PolicyConfiguration{
		//		Type:        &policy.PolicyTypeRef{
		//		Id:          &policyuuid,
		//		},
		//		CreatedDate: &azuredevops.Time{Time: t},
		//		IsBlocking:  &isblocking,
		//		IsDeleted:   &isdeleted,
		//		IsEnabled:   &isenabled,
		//		Settings:	Settings{
		//			CreatorVoteCounts:    	false,
		//			AllowDownvotes:			false,
		//			BlockLastPusherVote:	true,
		//			AllowNoFastForward:		false,
		//			AllowRebase: 			true,
		//			AllowRebaseMerge: 		false,
		//			AllowSquash:			true,
		//			RequiredReviewerIds: 	[]string{"4d49214c-c791-6e27-9d74-bcce48230683"},
		//			MinimumApproverCount: 	1,
		//			ResetOnSourcePush:    	true,
		//			Scope:                []Scope{{
		//				RepositoryId: repositoryid,
		//				RefName:      "refs/heads/dev",
		//				MatchKind:    "exact",
		//			}},
		//		},
		//
		//	},
		//	Project:         projname,
		//})
		//if err != nil {
		//	log.Fatal(err)
		//}
		//fmt.Printf("policy name is %s",pol.Settings)
	}else {
		log.Fatalln("Please specify a config file for the repositories with the argument --file")
	}
}


//fmt.Sprintf(`{
//	"creatorVoteCounts": False,
//	"minimumApproverCount": 1,
//	"resetOnSourcePush": True,
//    "scope": [
//      {
//        "repositoryId": %s,
//        "refName": "refs/heads/dev",
//        "matchKind": "exact"
//      }
//    ]
//  }`,"dothatshitdoit")
