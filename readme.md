# OC Cluster Automator
Create and Destroy Openshift clusters with hybrid overlay on AWS and Azure

## Environment Variables
* `APP_OCSTOREPATH`         : Path to store cluster credentials and binaries 
* `APP_CLUSTERNAMEPREFIX`   : Prefix for cluster name (preferably username)
* `APP_CLUSTERPULLSECRET`   : Pull secret to install Openshift cluster
* `APP_SSHKEY`              : SSH Public key for cluster VMs

## Prerequisites 
* `openshift-install` binary as a path variable 
* AWS credentials stored and configured in `~/.aws` folder
* Azure credentials stored and configured in `~/.azure` folder

## Run Flags
* `-create` : Create cluster 
* `-destroy`: Delete cluster
* `-platfrom`: Platform to create cluster on (AWS/Azure) <br>
                eg. `-platform="azure"`
* `-dryrun` : View the command to be executed

## Run
### Create Cluster 
`go run main.go -create -platform=<aws/azure>`

### Destroy Cluster
`go run main.go -destroy -platform=<aws/azure>`

