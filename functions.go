package main

import (
	"context"
	"fmt"
	gt "github.com/go-git/go-git"
	"github.com/go-git/go-git/config"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/go-git/go-git/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
	"github.com/otiai10/copy"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	STATEPATH = "created-policies"
	PATH  = "repo"
	REMOTENAME = "origin"
	ADDPATH    = "asset.yml"
	STATECOMMITMSG = "Adding created policies states"
	COMMITMSG  = "Adding asset.yml"
	MIN_NUMBER_OF_REWIERES_DISPLAY_NAME = "Minimum number of reviewers"
	MIN_NUMBER_OF_REWIERES_UUID = "fa4e907d-c16b-4a4c-9dfa-4906e5d171dd"
	WORK_ITEM_LINKING_DISPLAY_NAME = "Work item linking"
	WORK_ITEM_LINKING_DISPLAY_UUID = "40e92b44-2fe1-4dd6-b3d8-74a9c21d0c6e"
	COMMENT_REQUIREMENTS_DISPLAY_NAME = "Comment requirements"
	COMMENT_REQUIREMENTS_UUID = "c6a1889d-b943-4856-b76f-9e46bb6b0df2"
	REQUIRE_A_MERGE_STRATEGY_DISPLAY_NAME = "Require a merge strategy"
	REQUIRE_A_MERGE_STRATEGY_UUID =	 "fa4e907d-c16b-4a4c-9dfa-4916e5d171ab"
	REQUIRED_REVIEWERS_DISPLAY_NAME = "Required reviewers"
	REQUIRED_REVIEWERS_UUID = "fd2167ab-b0be-447a-8ec8-39368250530e"
)

func InitAllRepos(remoteUrl string,username string, password string,i int) {
	err := copy.Copy(PATH,fmt.Sprintf("%s%d",PATH,i))
	if err != nil {
		panic(err)
	}
	_, err = gt.PlainInit(fmt.Sprintf("%s%d",PATH,i),false)
	if err != nil {
		panic(err)
	}
	d, err := gt.PlainOpen(fmt.Sprintf("%s%d",PATH,i))
	if err != nil {
		panic(err)
	}
	_, err = d.CreateRemote(&config.RemoteConfig{
		Name: REMOTENAME,
		URLs: []string{remoteUrl},
	})
	w, err := d.Worktree()
	if err != nil {
		panic(err)
	}
	err = w.AddWithOptions(&gt.AddOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}
	_, err = w.Status()
	if err != nil {
		panic(err)
	}
	//fmt.Println(status)
	_, err = w.Commit(COMMITMSG, &gt.CommitOptions{
		Author: &object.Signature{
			Email: username,
			When:  time.Now(),
		},
	})
	err = d.Push(&gt.PushOptions{
		RemoteName: REMOTENAME,
		RefSpecs:   nil,
		Auth:       &http.BasicAuth{
			Username: username,
			Password: password,
		},
		Progress:   nil,
		Prune:      false,
		Force:      false,
	})
	if err != nil {
		panic(err)
	}
	re, err := d.Remotes()
	if err != nil {
		panic(err)
	}
	for _, remote := range re {
		err = d.DeleteRemote(remote.Config().Name)
		if err != nil {
			panic(err)
		}
	}
	err = os.RemoveAll(fmt.Sprintf("%s%d",PATH,i))
	if err != nil {
		panic(err)
	}
}

func SavePoliciesStates(username string, password string) {
	d, err := gt.PlainOpen(".")
	if err != nil {
		panic(err)
	}
	w, err := d.Worktree()
	if err != nil {
		panic(err)
	}
	err = w.AddWithOptions(&gt.AddOptions{
		All: true,
		Glob: STATEPATH,
	})
	if err != nil {
		panic(err)
	}
	_, err = w.Status()
	if err != nil {
		panic(err)
	}
	//fmt.Println(status)
	_, err = w.Commit(STATECOMMITMSG, &gt.CommitOptions{
		Author: &object.Signature{
			Email: username,
			When:  time.Now(),
		},
	})
	err = d.Push(&gt.PushOptions{
		RemoteName: REMOTENAME,
		RefSpecs:   nil,
		Auth:       &http.BasicAuth{
			Username: username,
			Password: password,
		},
		Progress:   nil,
		Prune:      false,
		Force:      false,
	})
	if err != nil {
		panic(err)
	}
}

func GetAllRepos(client git.Client,ctx context.Context,
	project *string)  []string {
	responseValue, err := client.GetRepositories(ctx, git.GetRepositoriesArgs{Project: project})

	if err != nil {
		log.Fatal(err)
	}
	existingrepos := make([]string,len(*responseValue))
	for i, getrepos := range (*responseValue) {
		existingrepos[i] =  *getrepos.Name
	}
	return existingrepos
}


func CreateRepos(client git.Client,ctx context.Context,
	username string, password string,
	project *string, name *string,branches *[]string,
	reposids []PolicyRepoIdAndBranch, i int,
	wg *sync.WaitGroup) {
	defer wg.Done()
	repos, err := client.CreateRepository(ctx, git.CreateRepositoryArgs{
		GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
			Name: name,
		},
		Project: project,
	})
	if err != nil {
		log.Fatalf("There was some error creating the repo %v", err)
	}
	reposids[i] = PolicyRepoIdAndBranch{
		RepoName: *repos.Name,
		RepoId: *repos.Id,
		Branches: *branches,
	}
	InitAllRepos(*repos.RemoteUrl,username,password,i)
	CreateBranch(client,ctx,project,name,GetCommitIdBranch(client,ctx,project,name),*branches)
	fmt.Printf("The repo %s  was created with the url for clone is %s\n", *repos.Name, *repos.SshUrl)
}

func CreateBranch(client git.Client,ctx context.Context,
					project *string,repo *string,newobjectid *string,
					branches []string)  {
	oldobjectid := "0000000000000000000000000000000000000000"
	islocked := false
	for _, branch := range  branches {
		branchresponse, err := client.UpdateRefs(ctx, git.UpdateRefsArgs{
			RefUpdates: &[]git.GitRefUpdate{{
				IsLocked:    &islocked,
				Name:        &branch,
				OldObjectId: &oldobjectid,
				NewObjectId: newobjectid,
			},
			},
			RepositoryId: repo,
			Project:      project,
		})
		if err != nil {
			log.Fatalf("There was some error creating the tag %v", err)
		}
		for _, branch := range (*branchresponse) {
			fmt.Printf("The branch %s from  repo %v was created and it's success status is  %v\n", *branch.Name,*branch.RepositoryId, *branch.UpdateStatus)
		}
	}
}


func GetCommitIdBranch(client git.Client,ctx context.Context,
					   project *string,repo *string)  *string{
	branchmodel := "master"
	branch, err := client.GetBranch(ctx,git.GetBranchArgs{
		RepositoryId:          repo,
		Name:                  &branchmodel,
		Project:               project,
	})
	if err != nil {
		log.Fatalf("There was some error creating the tag %v", err)
	}
	fmt.Printf("The branch %s  was created and it's last commit  is  %v\n", *branch.Name,*branch.Commit.CommitId)
	return branch.Commit.CommitId
}

func CreateBranchPolicy(client policy.Client,ctx context.Context,
					repoid uuid.UUID, project *string,
					branches []string,typedn string,
					typeuuid string,settings SettingsPolicy,
					isBlocking bool,reponame string,wg *sync.WaitGroup)  {
	defer wg.Done()
	isdeleted := false
	isenabled := true
	tuuid, err := uuid.Parse(typeuuid)
	if err != nil {
		log.Fatal(err)
	}

	t, err := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	if err != nil {
		log.Fatal(err)
	}
	pol, err := client.CreatePolicyConfiguration(ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: &policy.PolicyConfiguration{
			Type: &policy.PolicyTypeRef{
				DisplayName: &typedn,
				Id:          &tuuid,
			},
			CreatedDate: &azuredevops.Time{Time: t},
			IsBlocking:  &isBlocking,
			IsDeleted:   &isdeleted,
			IsEnabled:   &isenabled,
			Settings: settings,
		},
		Project: project,
	})
	if err != nil {
		log.Fatal(err)
	}
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(fmt.Sprintf("created-policies/%s.yaml",reponame), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//polbranches := fmt.Sprintf("[%s]",strings.Join(branches,","))
	if _, err := f.Write([]byte(fmt.Sprintf("- policyid: %d\n" +
													"  repoid: %v\n" +
													"  branch: %s\n" +
													"  typeid: %v\n" +
													"  typedisplayname: %s\n",
													*pol.Id,
													settings.Scope[0].RepositoryId,
													settings.Scope[0].RefName,
													tuuid,
													typedn)));
	err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("policy with name %s and id %d was created\n", *pol.Type.DisplayName,*pol.Id)
}

func UpdateBranchPolicy(client policy.Client,ctx context.Context,
	repoid uuid.UUID, project *string,
	typedn string,typeuuid string,
	settings SettingsPolicy,isBlocking bool,
	policyid *int, wg *sync.WaitGroup)  {
	defer wg.Done()
	isenabled := true
	tuuid, err := uuid.Parse(typeuuid)
	if err != nil {
		log.Fatal(err)
	}
	pol, err := client.UpdatePolicyConfiguration(ctx, policy.UpdatePolicyConfigurationArgs{
		Configuration: &policy.PolicyConfiguration{
			Type: &policy.PolicyTypeRef{
				Id:          &tuuid,
			},
			IsBlocking:  &isBlocking,
			IsEnabled:   &isenabled,
			Settings: settings,
		},
		Project: project,
		ConfigurationId: policyid,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("policy with name %s and id %d was updated\n", *pol.Type.DisplayName,*pol.Id)
}



func (c *ReposConfig) getConf(ReposFile *string) *ReposConfig {

	yamlFile, err := ioutil.ReadFile(*ReposFile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func (p *CreatedPolicies) getConf(ReposFile string) *CreatedPolicies {

	yamlFile, err := ioutil.ReadFile(ReposFile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, p)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return p
}

func Find(slice []string, val string)  bool{
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
