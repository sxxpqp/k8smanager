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
)

type PodInfo struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Status     string   `json:"status"`
	Ready      string   `json:"ready"` // "就绪数/总数" 如 "1/2"
	Restarts   int32    `json:"restarts"`
	Node       string   `json:"node"`
	Age        string   `json:"age"`
	Containers []string `json:"containers"` // 容器名列表，终端/日志切换用
}

func GetPods(c *gin.Context) {
	namespace := c.Query("namespace")

	pl, err := config.Client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	podlist := make([]PodInfo, 0, len(pl.Items))
	for _, p := range pl.Items {
		// 统计重启次数 & 就绪容器数
		var restarts int32
		readyCount := 0
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
			if cs.Ready {
				readyCount++
			}
		}

		// 收集所有容器名（用于日志/终端切换）
		containers := make([]string, 0, len(p.Spec.Containers))
		for _, v := range p.Spec.Containers {
			containers = append(containers, v.Name)
			// fmt.Printf("v.Name: %v\n", v.Name)

		}

		podlist = append(podlist, PodInfo{
			Name:       p.Name,
			Namespace:  p.Namespace,
			Status:     string(p.Status.Phase),
			Ready:      fmt.Sprintf("%d/%d", readyCount, len(p.Spec.Containers)),
			Restarts:   restarts,
			Node:       p.Spec.NodeName,
			Age:        p.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Containers: containers,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": podlist})
}

func DeletePod(c *gin.Context) {

	namespace := c.Param("namespace")
	pod := c.Param("pod")

	err := config.Client.CoreV1().Pods(namespace).Delete(context.Background(), pod, metav1.DeleteOptions{})
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

	err := config.Client.CoreV1().Pods(namespace).Delete(context.Background(), pod, metav1.DeleteOptions{})
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
