pipeline {
	agent any
	stages {
		stage('Checkout') {
			steps {
				dir("/home/os/Desktop/k8s/mini-k8s-2023"){
					git credentialsId: 'yhan', url:'https://gitee.com/yhan-yyds/mini-k8s-2023.git'
				}
			}
		}
		stage('Build') {
			steps {
				dir("/home/os/Desktop/k8s/mini-k8s-2023"){
					sh 'make'
				}
			}
		}
		stage('Test') {
			steps {
				dir("/home/os/Desktop/k8s/mini-k8s-2023"){
					sh 'make test'
				}
			}
		}
	}
}
