def gitBranch = env.BRANCH_NAME
def imageName = "memphis-http-proxy"
def gitURL = "git@github.com:Memphisdev/memphis-http-proxy.git"
def repoUrlPrefix = "memphisos"

node {
  git credentialsId: 'main-github', url: gitURL, branch: gitBranch
  def versionTag = readFile "./version.conf"
  
  try{

    stage('Login to Docker Hub') {
      withCredentials([usernamePassword(credentialsId: 'docker-hub', usernameVariable: 'DOCKER_HUB_CREDS_USR', passwordVariable: 'DOCKER_HUB_CREDS_PSW')]) {
      sh 'docker login -u $DOCKER_HUB_CREDS_USR -p $DOCKER_HUB_CREDS_PSW'
      }
    }

 
    stage('Build and push image to Docker Hub') {
      sh "docker buildx use builder"
      sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}:${gitBranch} --platform linux/amd64,linux/arm64 ."
    }
	
     if (env.BRANCH_NAME ==~ /(latest)/) {
     stage('checkout to version branch'){
	    withCredentials([sshUserPrivateKey(keyFileVariable:'check',credentialsId: 'main-github')]) {
	    sh "git reset --hard origin/latest"
	    sh "GIT_SSH_COMMAND='ssh -i $check'  git checkout -b ${versionTag}"
       	    sh "GIT_SSH_COMMAND='ssh -i $check' git push --set-upstream origin ${versionTag}"
  	  }
	}
	      
	stage('Create new release') {
          sh 'sudo yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo'
          sh 'sudo yum install gh -y'
          withCredentials([string(credentialsId: 'gh_token', variable: 'GH_TOKEN')]) {
	    sh "gh release create ${versionTag} --generate-notes"
          }
	}
     }

    notifySuccessful()
  
  }
    catch (e) {
      currentBuild.result = "FAILED"
      cleanWs()
      notifyFailed()
      throw e
  }
}
 def notifySuccessful() {
  emailext (
      subject: "SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
      body: """SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':
        Check console output and connection attributes at ${env.BUILD_URL}""",
      recipientProviders: [requestor()]
    )
}
def notifyFailed() {
  emailext (
      subject: "FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
      body: """FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':
        Check console output at ${env.BUILD_URL}""",
      recipientProviders: [requestor()]
    )
}  

