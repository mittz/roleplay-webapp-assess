package kubernetesengine

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	container "cloud.google.com/go/container/apiv1"
	"github.com/mittz/roleplay-webapp-assess/utils"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type GKECluster struct {
	Name      string
	Region    string
	ProjectID string
}

type Pod struct {
	id     string
	region string
	zone   string
	cost   float64
}

func (p Pod) GetID() string {
	return p.id
}

func (p Pod) GetCost() float64 {
	return p.cost
}

func (p Pod) SetCost(cost float64) {
	p.cost = cost
}

func (p Pod) GetRegion() string {
	return p.region
}

func (p Pod) GetZone() string {
	return p.zone
}

func GetGKEClusters(projectID string) ([]GKECluster, error) {
	ctx := context.Background()
	c, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return []GKECluster{}, err
	}
	defer c.Close()

	parent := fmt.Sprintf("projects/%s/locations/-", projectID)
	req := &containerpb.ListClustersRequest{
		Parent: parent,
	}

	resp, err := c.ListClusters(ctx, req)
	if err != nil {
		return []GKECluster{}, err
	}

	var clusters []GKECluster
	for _, c := range resp.GetClusters() {
		clusters = append(clusters, GKECluster{
			Name:      c.GetName(),
			Region:    utils.GetRegionFromZone(c.GetLocation()),
			ProjectID: projectID,
		})
	}

	return clusters, nil
}

func (c GKECluster) GetPods() ([]Pod, error) {
	if err := exec.Command(
		"gcloud",
		"container",
		"clusters",
		"get-credentials",
		c.Name,
		"--region",
		c.Region,
		"--project",
		c.ProjectID,
	).Err; err != nil {
		return []Pod{}, err
	}

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return []Pod{}, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return []Pod{}, err
	}

	kubeclient := client.CoreV1().Pods("default")
	pods, err := kubeclient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []Pod{}, err
	}

	for _, pod := range pods.Items {
		log.Println(pod.String())
	}

	log.Fatal("Finish")

	return []Pod{}, nil
}

func GetPodAll(projectID string) ([]Pod, error) {
	clusters, err := GetGKEClusters(projectID)
	if err != nil {
		return []Pod{}, err
	}

	var pods []Pod
	for _, cluster := range clusters {
		p, err := cluster.GetPods()
		if err != nil {
			return []Pod{}, err
		}

		pods = append(pods, p...)
	}

	return pods, nil
}
