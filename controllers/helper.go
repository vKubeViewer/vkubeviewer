/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// this file contains some helper functions which are used by controllers of different apis

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ByteCountIEC formats byte size to human-readable-format in 2 digits float
// this brilliant code is from https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

// UpdateStatusList checks two arrays are the same or not,
// if they are different, return the second array
func UpdateStatusList(status, current []string) []string {
	if len(status) != len(current) {
		return current
	}
	for i, v := range status {
		if current[i] != v {
			return current
		}
	}
	return current

}

// ListK8sNodes connects with k8s api and get the list of nodes
func ListK8sNodes() []string {
	var curK8sNode []string
	var kubeconfig *string
	path := homedir.HomeDir() + "/.kube/config"
	kubeconfig = &path
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nodeList := clientSet.CoreV1().Nodes()
	nodes, err := nodeList.List(context.TODO(), v1.ListOptions{})

	if err != nil {
		panic(err)
	}

	for _, item := range nodes.Items {
		curK8sNode = append(curK8sNode, item.ObjectMeta.Name)
	}
	return curK8sNode
}

// stringInSlice check whether a string is in a slice or not
func stringInSlice(s string, list []string) bool {
	for _, ele := range list {
		if s == ele {
			return true
		}
	}
	return false
}
