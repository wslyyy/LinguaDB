package initialization

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func StartHTTPServer(config Config, r *gin.Engine) (err error) {
	log.Printf("http server started: http://localhost:%d\n", config.HttpPort)
	err = r.Run(fmt.Sprintf(":%d", config.HttpPort))
	if err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}
	return nil
}
