package podapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"learn-go/k8s-api/config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodInfo struct {
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Status    string  `json:"status"`
	Ready     bool    `json:"ready"`
	Restarts  int32   `json:"restarts"`
	Node      string  `json:"node"`
	Age       v1.Time `json:"age"`
}

func GetPods(c *gin.Context) {
	// firstname := c.DefaultQuery("firstname", "Guest")
	namespace := c.Query("namespace") // shortcut for c.Request.URL.Query().Get("lastname")

	pl, err := config.Client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	var podlist = []PodInfo{}
	for _, v := range pl.Items {

		// podinfo.Ready = v.Status.ContainerStatuses
		v2, err := config.Client.CoreV1().Pods(namespace).Get(context.Background(), v.Name, v1.GetOptions{})
		if err != nil {
			return
		}

		for _, v1 := range v2.Status.ContainerStatuses {
			podinfo := PodInfo{
				Name:      v2.ObjectMeta.Name,
				Namespace: namespace,
				Status:    string(v2.Status.Phase),
				Ready:     v1.Ready,
				Restarts:  v1.RestartCount,
				Node:      v.Spec.NodeName,
				Age:       v.ObjectMeta.CreationTimestamp,
			}
			podlist = append(podlist, podinfo)

		}

	}

	c.JSON(200, gin.H{
		"data": podlist,
	})

}

func DeletePod(c *gin.Context) {

	namespace := c.Param("namespace")
	pod := c.Param("pod")

	err := config.Client.CoreV1().Pods(namespace).Delete(context.Background(), pod, v1.DeleteOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod 删除成功"})

}

func Getlogs(c *gin.Context) {
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	tailStr := c.Query("tail")

	// string → int64
	var tailLines int64 = 200
	if n, err := strconv.ParseInt(tailStr, 10, 64); err == nil {
		tailLines = n
	}

	// 1. 构造请求
	req := config.Client.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{
		TailLines: &tailLines,
	})

	// 2. 建立流
	stream, err := req.Stream(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer stream.Close()

	// 3. 读取全部内容
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, stream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": buf.String()})

}
func PostPodRestart(c *gin.Context) {

	namespace := c.Param("namespace")
	pod := c.Param("pod")

	err := config.Client.CoreV1().Pods(namespace).Delete(context.Background(), pod, v1.DeleteOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod 已删除，控制器将自动重建"})
}

func GetPodEvents(c *gin.Context) {

	namespace := c.Param("namespace")
	pod := c.Param("pod")

	list, err := config.Client.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod),
	})
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	result := make([]EventInfo, 0)
	for _, e := range list.Items {
		result = append(result, EventInfo{
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Count:     e.Count,
			FirstTime: e.FirstTimestamp.Format("2006-01-02 15:04:05"),
			LastTime:  e.LastTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(200, gin.H{"data": result})

}

type EventInfo struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Count     int32  `json:"count"`
	FirstTime string `json:"firstTime"`
	LastTime  string `json:"lastTime"`
}
