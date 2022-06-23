package route

import (
	"github.com/gin-gonic/gin"
	"openDevops/src/modules/server/web/controller"
)

func configRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/nodeAdd", controller.NodePathAdd)
		api.GET("/nodePathQuery", controller.NodePathQuery)
		api.POST("/resourceMount", controller.ResourceMount)
		api.POST("/resource-unmount", controller.ResourceUnMount)

		/*
		 matcher1 = {

		          "key": "stree_app",
		          "value": "kafka",
		          "type": 1
		      }

		      matcher2 = {
		          "key": "name",
		          "value": "genMockResourceHost_host_3",
		          "type": 1
		      }
		      matcher3 = {
		          "key": "private_ip",
		          "value": "8.*.5.*",
		          "type": 3
		      }
		      matcher4 = {
		          "key": "os",
		          "value": "amd64",
		          "type": 2
		      }

		      matcher5 = {

		          "key": "stree_app",
		          "value": "kafka|es",
		          "type": 3
		      }
		      matcher6 = {

		          "key": "stree_group",
		          "value": "inf",
		          "type": 1
		      }

		{
		    "resource_type": "resource_host",
		    "labels": [
		            {
		                "key": "stree_app",
		                "value": "kafka|es",
		                "type": 3
		            },
		            {
		                "key": "stree_group",
		                "value": "inf",
		                "type": 1
		            }
		    ],
		    "target_label": "cluster"
		}
		*/
		api.POST("/resourceQuery", controller.ResourceQuery)
	}
}