package nsapi

import (
	"context"
	"fmt"
	"learn-go/k8s-api/config"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaces(c *gin.Context) {
	// Return JSON response

	var optslist metav1.ListOptions
	namespacelist, err := config.Client.CoreV1().Namespaces().List(context.Background(), optslist)
	if err != nil {
		return
	}
	var list []string
	for _, v := range namespacelist.Items {
		list = append(list, v.Name)
	}
	fmt.Printf("list: %v\n", list)
	c.JSON(200, gin.H{

		"data": list,
	})

}
