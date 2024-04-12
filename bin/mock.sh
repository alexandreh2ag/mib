#!/bin/bash

go install go.uber.org/mock/mockgen@latest

rm -rf mock
mockgen -destination=mock/git/git_repository.go -package=mock_git github.com/alexandreh2ag/mib/git Repository
mockgen -destination=mock/git/git_worktree.go -package=mock_git github.com/alexandreh2ag/mib/git Worktree
mockgen -destination=mock/git/git_manager.go -package=mock_git github.com/alexandreh2ag/mib/git Manager
mockgen -destination=mock/types/container/image.go -package=mock_types_container github.com/alexandreh2ag/mib/types/container BuilderImage
mockgen -destination=mock/docker/client.go -package=mock_docker github.com/docker/docker/client APIClient
mockgen -destination=mock/exec/command.go -package=mock_exec github.com/alexandreh2ag/mib/exec Executable
