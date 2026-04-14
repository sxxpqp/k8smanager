package deploymentsapi

import (
	"context"
	"fmt"
	"learn-go/k8s-api/config"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ContainerInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type DeployInfo struct {
	Name          string          `json:"name"`
	Namespace     string          `json:"namespace"`
	Replicas      int32           `json:"replicas"`
	ReadyReplicas int32           `json:"readyReplicas"`
	Image         string          `json:"image"`      // 第一个容器镜像（兼容旧逻辑）
	Containers    []ContainerInfo `json:"containers"` // 所有容器（镜像更新用）
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
		// 收集所有容器信息
		containers := make([]ContainerInfo, 0, len(v.Spec.Template.Spec.Containers))
		for _, c := range v.Spec.Template.Spec.Containers {
			containers = append(containers, ContainerInfo{
				Name:  c.Name,
				Image: c.Image,
			})
		}

		image := ""
		if len(containers) > 0 {
			image = containers[0].Image
		}

		var replicas int32
		if v.Spec.Replicas != nil {
			replicas = *v.Spec.Replicas
		}

		deployinfo := DeployInfo{
			Name:          v.Name,
			Namespace:     v.Namespace,
			Replicas:      replicas,
			ReadyReplicas: v.Status.ReadyReplicas,
			Image:         image,
			Containers:    containers,
		}
		deploymentlist = append(deploymentlist, deployinfo)
	}

	c.JSON(200, gin.H{
		"data": deploymentlist,
	})

}

func PATCHDeploymentScale(c *gin.Context) {
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

	deployment, err := config.Client.AppsV1().Deployments(namespace).Get(context.Background(), deploy, v1.GetOptions{})
	if err != nil {

		return
	}
	deployment.Spec.Replicas = &v.Replicas
	_, err = config.Client.AppsV1().Deployments(namespace).Update(context.Background(), deployment, v1.UpdateOptions{})
	if err != nil {
		return
	}
	log.Printf("deployment %s/%s scale to %d", deployment.Namespace, deployment.Name, v.Replicas)
	c.JSON(200, gin.H{
		"message": "Deployment scale updated",
	})

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

type DeployImage struct {
	Container string `json:"container"`
	Image     string `json:"image"`
}

func PatchDeplayUpdateimage(c *gin.Context) {

	ns := c.Param("namespace")
	name := c.Param("deploy")
	var di DeployImage
	err := c.BindJSON(&di)
	if err != nil {
		c.JSON(200, "json绑定失败")
		return
	}

	deployment, err := config.Client.AppsV1().Deployments(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		c.JSON(200, gin.H{"message": err.Error()})
		return
	}
	// ✅ 用下标修改，range 遍历的是副本，直接改不生效
	for i := range deployment.Spec.Template.Spec.Containers {
		if deployment.Spec.Template.Spec.Containers[i].Name == di.Container {
			deployment.Spec.Template.Spec.Containers[i].Image = di.Image
		}
	}
	_, err = config.Client.AppsV1().Deployments(ns).Update(context.Background(), deployment, v1.UpdateOptions{})
	if err != nil {
		c.JSON(200, gin.H{"message": err.Error()})
		return
	}
	//data := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, di.Container, di.Image)
	//
	//_, err = config.Client.AppsV1().Deployments(ns).Patch(context.Background(), name, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})
	//if err != nil {
	//	return
	//}

	c.JSON(200, gin.H{

		"message": "镜像更新成功",
	})
}
