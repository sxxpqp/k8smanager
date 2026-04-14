package main

import (
	"learn-go/k8s-api/configmapapi"
	"learn-go/k8s-api/deploymentsapi"
	"learn-go/k8s-api/ingressapi"
	"learn-go/k8s-api/nodeapi"
	"learn-go/k8s-api/nsapi"
	"learn-go/k8s-api/podapi"
	"learn-go/k8s-api/secretsapi"
	"learn-go/k8s-api/servicesapi"
	"learn-go/k8s-api/stsapi"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	// Define a simple GET endpoint
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/api/namespaces", nsapi.GetNamespaces)
	r.GET("/api/pods", podapi.GetPods)
	r.DELETE("/api/pods/:namespace/:pod", podapi.DeletePod)
	r.POST("/api/pods/:namespace/:pod/restart", podapi.PostPodRestart)
	r.GET("/api/deployments", deploymentsapi.GetDeployments)

	r.GET("/api/pods/:namespace/:pod/logs", podapi.Getlogs)

	r.PATCH("/api/deployments/:namespace/:deploy/scale", deploymentsapi.PATCHDeploymentScale)
	r.POST("/api/deployments/:namespace/:deploy/restart", deploymentsapi.RestartDeployment)
	r.GET("/api/services", servicesapi.GetServicesList)
	r.GET("/api/pods/:namespace/:pod/events", podapi.GetPodEvents)
	r.GET("/api/pods/:namespace/:pod/exec", podapi.ExecPod)
	r.PATCH("/api/deployments/:namespace/:deploy/image", deploymentsapi.PatchDeplayUpdateimage)
	r.GET("/api/configmaps", configmapapi.GetConfigMap)
	r.GET("/api/nodes", nodeapi.GetNodes)
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)

	//	GET /api/configmaps/:namespace/:name
	r.GET("/api/configmaps/:namespace/:name", configmapapi.GetConfigMapDescribe)
	//PUT /api/configmaps/:namespace/:name

	r.PUT("/api/configmaps/:namespace/:name", configmapapi.PutConfigMap)

	r.GET("/api/secrets", secretsapi.GetSecrets)
	r.GET("/api/secrets/:namespace/:name", secretsapi.GetSecret)
	///api/statefulsets?namespace=default
	r.GET("/api/statefulsets", stsapi.Getstss)
	r.PATCH("/api/statefulsets/:namespace/:name/scale", stsapi.PATCHStatefulSetScale)

	r.GET("/api/ingresses", ingressapi.GetIngresses)
	r.Run("0.0.0.0:8080")
}
