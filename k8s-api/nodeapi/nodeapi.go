package nodeapi

import (
	"context"
	"learn-go/k8s-api/config"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeInfo struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Roles   []string `json:"roles"`
	IP      string   `json:"ip"`
	OS      string   `json:"os"`
	Kernel  string   `json:"kernel"`
	Runtime string   `json:"runtime"`
	CPU     string   `json:"cpu"`
	Memory  string   `json:"memory"`
	Age     string   `json:"age"` // 建议存储格式化后的字符串
}

func GetNodes(c *gin.Context) {

	nodes, err := config.Client.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return
	}
	nodeinfolist := make([]NodeInfo, 0)
	// roles := make([]string, 0)
	// for k := range nodes.Items.Labels {
	// 	if strings.HasPrefix(k, "node-role.kubernetes.io/") {
	// 		roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
	// 	}
	// }
	//
	for _, v := range nodes.Items {

		nodeinfo := NodeInfo{

			Name:   v.Name,
			Status: string(v.Status.Phase),
			Age:    v.CreationTimestamp.Format("2006-01-02 15:04:05"),
			// IP:      string(v.Status.Addresses[0].Address),
			OS:      v.Status.NodeInfo.OSImage,
			Kernel:  v.Status.NodeInfo.KernelVersion,
			Runtime: v.Status.NodeInfo.ContainerRuntimeVersion,
			CPU:     v.Status.Capacity.Cpu().String(),
			Memory:  v.Status.Capacity.Memory().String(),
		}

		for _, v := range v.Status.Conditions {
			if v.Type == "Ready" {
				// "True"→Ready
				nodeinfo.Status = string(v.Status)
			}

		}
		for _, v := range v.Status.Addresses {

			if v.Type == "InternalIP" {
				nodeinfo.IP = v.Address
			}

		}
		roles := make([]string, 0)
		for k, _ := range v.Labels {
			if k == "node-role.kubernetes.io/control-plane" {
				roles = append(roles, "control-plane")

			}

		}
		nodeinfo.Roles = roles

		nodeinfolist = append(nodeinfolist, nodeinfo)

	}

	c.JSON(200, gin.H{
		"data": nodeinfolist,
	})

}
