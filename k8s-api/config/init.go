package config

import (
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// getClient 本地开发：读取 ~/.kube/config
func getClient() (*kubernetes.Clientset, error) {
	// 使用推荐的路径变量代替手动拼接
	kubeconfig := clientcmd.RecommendedHomeFile
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// getInClusterClient 在 Pod 里运行（Operator 部署在 K8s 里时）
func getInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

var Client     *kubernetes.Clientset
var RestConfig *rest.Config  // exec 终端需要用到

func init() {
	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("加载 kubeconfig 失败: %s", err.Error())
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("创建 client 失败: %s", err.Error())
	}

	fmt.Println("成功连接到 Kubernetes 集群！")
	Client     = client
	RestConfig = cfg  // 保存原始 config，exec 时用
}
