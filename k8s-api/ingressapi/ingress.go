package ingressapi

import (
	"context"
	"fmt"
	"learn-go/k8s-api/config"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetIngresses(c *gin.Context) {

	ns := c.Query("namespace")
	fmt.Printf("ns: %v\n", ns)
	ingresslist, err := config.Client.NetworkingV1().Ingresses(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	ingressInfoList := make([]IngressInfo, 0)
	for _, v := range ingresslist.Items {

		// cn := v.Spec.IngressClassNam
		v1 := IngressInfo{
			Name:      v.Name,
			Namespace: v.Namespace,
			// ClassName: ,
			Age: v.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		ingressInfoList = append(ingressInfoList, v1)
	}
	c.JSON(200, gin.H{

		"data": ingressInfoList,
	})
}

//	返回: [{
//	  "name": "mysql",
//	  "namespace": "default",
//	  "replicas": 3,
//	  "readyReplicas": 3,
//	  "image": "mysql:8.0",
//	  "containers": [{"name":"mysql","image":"mysql:8.0"}],
//	  "age": "2024-01-10"
//	}]
type IngressPath struct {
	Path    string `json:"path"`
	Backend string `json:"backend"` // "svc:port"
}

type IngressRule struct {
	Host  string        `json:"host"`
	Paths []IngressPath `json:"paths"`
}

type IngressInfo struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	ClassName string        `json:"className"`
	Rules     []IngressRule `json:"rules"`
	Age       string        `json:"age"`
}
