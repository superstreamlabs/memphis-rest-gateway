def gitBranch = env.BRANCH_NAME
def imageName = "memphis-http-proxy"
def gitURL = "git@github.com:Memphisdev/memphis-http-proxy.git"
def repoUrlPrefix = "memphisos"

node {
  git credentialsId: 'main-github', url: gitURL, branch: gitBranch
  
  try{

    stage('Login to Docker Hub') {
      withCredentials([usernamePassword(credentialsId: 'docker-hub', usernameVariable: 'DOCKER_HUB_CREDS_USR', passwordVariable: 'DOCKER_HUB_CREDS_PSW')]) {
      sh 'docker login -u $DOCKER_HUB_CREDS_USR -p $DOCKER_HUB_CREDS_PSW'
      }
    }

 
    stage('Build and push image to Docker Hub') {
      sh "docker buildx use builder"
	  sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 ."
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

