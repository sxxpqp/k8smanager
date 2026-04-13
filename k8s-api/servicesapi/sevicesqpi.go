package servicesapi

import (
	"context"
	"fmt"
	"learn-go/k8s-api/config"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetServicesList(c *gin.Context) {
	namespace := c.Query("namespace")

	serverlist, err := config.Client.CoreV1().Services(namespace).List(context.Background(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(502, gin.H{

			"data": "api 502",
		})
		return
	}
	var serverInfolist []ServiceInfo

	for _, v := range serverlist.Items {
		ports := make([]string, 0, len(v.Spec.Ports))
		for _, p := range v.Spec.Ports {
			if p.NodePort > 0 {
				// NodePort 类型显示映射关系
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
			}
		}
		var serverINfo ServiceInfo
		serverINfo = ServiceInfo{
			Name:      v.Name,
			Namespace: v.Namespace,
			Type:      string(v.Spec.Type),
			ClusterIP: v.Spec.ClusterIP,
			Ports:     ports,
			Selector:  v.Spec.Selector, // 直接就是 map[string]string
			Age:       v.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		serverInfolist = append(serverInfolist, serverINfo)

	}

	c.JSON(200, gin.H{

		"data": serverInfolist,
	})

}

type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"clusterIP"`
	Ports     []string          `json:"ports"`
	Selector  map[string]string `json:"selector"`
	Age       string            `json:"age"`
}
