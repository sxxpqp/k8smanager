package configmapapi

import (
	"context"
	"learn-go/k8s-api/config"
	"log"
	"net/http"

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

func GetConfigMapDescribe(c *gin.Context) {
	//返回: {
	//	"data": {
	//		"name": "nginx-config",
	//			"namespace": "default",
	//			"data": { "key1": "value1", "key2": "value2" }  // cm.Data
	//	}
	//}
	ns := c.Param("namespace")
	name := c.Param("name")
	configMapList, err := config.Client.CoreV1().ConfigMaps(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return
	}
	v := ConfigMapDescribe{

		Name:      configMapList.Name,
		Namespace: configMapList.Namespace,
		Data:      configMapList.Data,
	}
	c.JSON(200, gin.H{
		"data": v,
	})

}

type ConfigMapDescribe struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
}

// PUT /api/configmaps/:namespace/:name
// Body: { "data": { "key1": "new value" } }
// 返回: { "message": "更新成功" }
//
// // Go 关键代码
// cm, _ := client.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
// cm.Data = body.Data
// client.CoreV1().ConfigMaps(ns).Update(ctx, cm, metav1.UpdateOptions{})
func PutConfigMap(c *gin.Context) {
	ns := c.Param("namespace")
	name := c.Param("name")
	var body Body
	err2 := c.BindJSON(&body)
	if err2 != nil {
		return
	}
	log.Println(body)
	cm, err := config.Client.CoreV1().ConfigMaps(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return
	}
	cm.Data = body.Data
	log.Println(cm.Data)
	cm, err = config.Client.CoreV1().ConfigMaps(ns).Update(context.Background(), cm, v1.UpdateOptions{})
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return

	}
	c.JSON(200, gin.H{
		"message": "更新成功",
	})

}

type Body struct {
	Data map[string]string `json:"data"`
}
