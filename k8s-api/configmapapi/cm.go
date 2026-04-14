package configmapapi

import (
	"context"
	"learn-go/k8s-api/config"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetConfigMap(c *gin.Context) {

	ns := c.Query("namespace")
	configmapitem, err := config.Client.CoreV1().ConfigMaps(ns).List(context.Background(), v1.ListOptions{})
	if err != nil {
		return
	}
	// ConfigMapInfolist := make([]ConfigMapInfo)
	ConfigMapInfolist := make([]ConfigMapInfo, 0)
	for _, v := range configmapitem.Items {

		configmapinfo := ConfigMapInfo{

			Name:      v.Name,
			Namespace: v.Namespace,
			DataCount: len(v.Data),
			Age:       v.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		ConfigMapInfolist = append(ConfigMapInfolist, configmapinfo)

	}

	c.JSON(200, gin.H{

		"data": ConfigMapInfolist,
	})
}

type ConfigMapInfo struct {
	Name      string `json:"name"` // 加 tag
	Namespace string `json:"namespace"`
	DataCount int    `json:"dataCount"` // 加 tag
	Age       string `json:"age"`
}
