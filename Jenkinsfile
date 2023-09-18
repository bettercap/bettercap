pipeline {
    agent any

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Go Application') {
            steps {
                sh 'go get -u github.com/bettercap/bettercap'
                sh 'go build -o myapp main.go'
            }
        }

        stage('Run Bettercap') {
            steps {
                sh './bettercap <options>'
            }
        }

        stage('Publish Artifacts') {
            steps {
                archiveArtifacts artifacts: 'myapp'
            }
        }
    }

    post {
        success {
            echo 'Build successful!'
        }
        failure {
            echo 'Build failed!'
        }
    }
}
