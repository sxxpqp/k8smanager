package deploymentsapi

import (
	"context"
	"fmt"
	"learn-go/k8s-api/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type DeployInfo struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Replicas      int32  `json:"replicas"`
	ReadyReplicas int32  `json:"readyReplicas"`
	Image         string `json:"image"`
}

type deployrs struct {
	Replicas int32 `json:"replicas"`
}

func GetDeployments(c *gin.Context) {
	// firstname := c.DefaultQuery("firstname", "Guest")
	namespace := c.Query("namespace") // shortcut for c.Request.URL.Query().Get("lastname")

	pl, err := config.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	var deploymentlist = []DeployInfo{}
	for _, v := range pl.Items {

		// podinfo.Ready = v.Status.ContainerStatuses
		deployinfo := DeployInfo{
			Name:          v.Name,
			Namespace:     namespace,
			Replicas:      *v.Spec.Replicas,
			ReadyReplicas: v.Status.ReadyReplicas,
			Image:         v.Spec.Template.Spec.Containers[0].Image,
		}
		deploymentlist = append(deploymentlist, deployinfo)

	}

	c.JSON(200, gin.H{
		"data": deploymentlist,
	})

}

func PATCHDeploymentScale(c *gin.Context) {
	ctx := context.Background()
	// replicas
	// replicasStr := c.PostForm("replicas")
	// replicasStr := c.PostForm("replicas")
	// fmt.Printf("replicasStr: %v\n", v)
	// n, err := strconv.ParseInt(replicasStr, 10, 32) // import "strconv"
	// replicas := int32(n)
	v := deployrs{}
	c.BindJSON(&v)
	namespace := c.Param("namespace")
	deploy := c.Param("deploy")

	scale, err := config.Client.AppsV1().Deployments(namespace).GetScale(
		ctx, deploy, metav1.GetOptions{},
	)
	scale.Spec.Replicas = v.Replicas
	fmt.Printf("v.Replicas: %v\n", v.Replicas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	config.Client.AppsV1().Deployments(namespace).UpdateScale(context.Background(), deploy, scale, v1.UpdateOptions{})

}

func RestartDeployment(c *gin.Context) {
	// patch pod template annotation 触发滚动重启（上面已给过代码）
	namespace := c.Param("namespace")
	deploy := c.Param("deploy")
	// 给 pod template 加 restartedAt 注解，触发滚动重启
	patch := fmt.Sprintf(`{
        "spec": {
            "template": {
                "metadata": {
                    "annotations": {
                        "kubectl.kubernetes.io/restartedAt": "%s"
                    }
                }
            }
        }
    }`, time.Now().Format(time.RFC3339))

	_, err := config.Client.AppsV1().Deployments(namespace).Patch(
		context.Background(),
		deploy,
		types.MergePatchType,
		[]byte(patch),
		metav1.PatchOptions{},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deployment 滚动重启中"})
}
