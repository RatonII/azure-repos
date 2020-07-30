package main

import (
	"context"
	"fmt"
	gt "github.com/go-git/go-git"
	"github.com/go-git/go-git/config"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/go-git/go-git/plumbing/transport/http"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	PATH  = "repo"
	REMOTENAME = "origin"
	ADDPATH    = "README.md"
	COMMITMSG  = "Adding README.md"
)

func InitAllRepos(remoteUrl string,username string, password string) {
	_, err := gt.PlainInit(PATH,false)
	if err != nil {
		panic(err)
	}
	d, err := gt.PlainOpen(PATH)
	if err != nil {
		panic(err)
	}
	//https://mariuss2007@dev.azure.com/mariuss2007/test/_git/YamlPipelineTest4
	_, err = d.CreateRemote(&config.RemoteConfig{
		Name: REMOTENAME,
		URLs: []string{remoteUrl},
	})
	w, err := d.Worktree()
	if err != nil {
		panic(err)
	}
	_, err = w.Add(ADDPATH)
	if err != nil {
		panic(err)
	}
	status, err := w.Status()
	if err != nil {
		panic(err)
	}
	fmt.Println(status)
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
	err = os.RemoveAll(fmt.Sprintf("%s/.git",PATH))
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

func CreateBranch(client git.Client,ctx context.Context,
					project *string,repo *string,newobjectid *string)  {
	branches := []string{"refs/heads/dev","refs/heads/test"}
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
			fmt.Printf("The branch %s  was created and it's success is  %v\n", *branch.Name, *branch.UpdateStatus)
		}
	}
}

func CreateRepos(client git.Client,ctx context.Context,
				username string, password string,
				project *string, name *string,wg *sync.WaitGroup) {
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
	InitAllRepos(*repos.RemoteUrl,username,password)
	CreateBranch(client,ctx,project,name,GetCommitIdBranch(client,ctx,project,name))
	fmt.Printf("The repo %s  was created with the url for clone is %s\n", *repos.Name, *repos.SshUrl)

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

func Find(slice []string, val string)  bool{
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
