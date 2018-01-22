## Pull Requests
If you are working directly on this repository instead of forking it, please make sure
that you follow the [git-flow](http://nvie.com/posts/a-successful-git-branching-model/)
naming convention. 
Your branch should start with `feature/your_feature_name`. When there is a lot of development 
happening on the `develop` branch, you should `rebase` your feature branch onto develop 
regularly.

If opening a PR against Ethermint, please use branch `develop` as the base branch
for changes. 
`unstable` should only be used to integrate multiple PRs that might have effects on each other.

Please make sure that all tests run with `make test` pass.



