package stsapi

import (
	"context"
	"fmt"
	"learn-go/k8s-api/config"
	"log"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Getstss(c *gin.Context) {
	ns := c.Query("namespace")
	fmt.Printf("ns: %v\n", ns)

	sts, err := config.Client.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	stsinfos := make([]StsInfo, 0)
	for _, v := range sts.Items {

		s := StsInfo{
			Name:          v.Name,
			Namespace:     v.Namespace,
			Replicas:      *v.Spec.Replicas,
			ReadyReplicas: v.Status.ReadyReplicas,
			Containers:    v.Spec.Template.Spec.Containers,
			Image:         v.Spec.Template.Spec.Containers[0].Image,
			Age:           v.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		stsinfos = append(stsinfos, s)
	}
	c.JSON(200, gin.H{
		"data": stsinfos,
	})
}

// [{
//   "name": "mysql",
//   "namespace": "default",
//   "replicas": 3,
//   "readyReplicas": 3,
//   "image": "mysql:8.0",
//   "containers": [{"name":"mysql","image":"mysql:8.0"}],
//   "age": "2024-01-10"
// }]

type StsInfo struct {
	Name          string         `json:"name"`
	Namespace     string         `json:"namespace"`
	Replicas      int32          `json:"replicas"`
	ReadyReplicas int32          `json:"readyReplicas"`
	Image         string         `json:"image"`
	Containers    []v1.Container `json:"containers"`
	Age           string         `json:"age"`
}

func PATCHStatefulSetScale(c *gin.Context) {
	// replicas
	// replicasStr := c.PostForm("replicas")
	// replicasStr := c.PostForm("replicas")
	// fmt.Printf("replicasStr: %v\n", v)
	// n, err := strconv.ParseInt(replicasStr, 10, 32) // import "strconv"
	// replicas := int32(n)
	v := Stsrs{}
	c.BindJSON(&v)
	namespace := c.Param("namespace")
	sts := c.Param("name")

	StatefulSet, err := config.Client.AppsV1().StatefulSets(namespace).Get(context.Background(), sts, metav1.GetOptions{})
	if err != nil {

		return
	}
	StatefulSet.Spec.Replicas = v.Replicas
	_, err = config.Client.AppsV1().StatefulSets(namespace).Update(context.Background(), StatefulSet, metav1.UpdateOptions{})
	if err != nil {
		return
	}
	log.Printf("StatefulSet %s/%s scale to %d", StatefulSet.Namespace, StatefulSet.Name, v.Replicas)
	c.JSON(200, gin.H{
		"message": "StatefulSet scale updated",
	})

}

type Stsrs struct {
	Replicas *int32
}
