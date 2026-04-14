package secretsapi

import (
	"context"
	"learn-go/k8s-api/config"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Secretsinfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	KeyCount  int    `json:"keyCount"`
	Age       string `json:"age"`
}

type Secretinfo struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Data map[string][]byte `json:"data"`
}

func GetSecrets(c *gin.Context) {

	ns := c.Query("namespace")

	secrets, err := config.Client.CoreV1().Secrets(ns).List(context.Background(), v1.ListOptions{})
	if err != nil {
		return
	}
	SercersInfoList := make([]Secretsinfo, 0)
	for _, v := range secrets.Items {
		si := Secretsinfo{
			Name:      v.Name,
			Namespace: v.Namespace,
			Type:      string(v.Type),
			KeyCount:  len(v.Data),
			Age:       v.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		// fmt.Printf("v.Name: %v\n", v.Name)
		SercersInfoList = append(SercersInfoList, si)

	}

	c.JSON(200, gin.H{
		"data": SercersInfoList,
	})

}

func GetSecret(c *gin.Context) {

	ns := c.Param("namespace")
	name := c.Param("name")

	v, err := config.Client.CoreV1().Secrets(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return
	}
	s := Secretinfo{
		Name: v.Name,
		Type: string(v.Type),
		Data: v.Data,
	}
	c.JSON(200, gin.H{

		"data": s,
	})
}
