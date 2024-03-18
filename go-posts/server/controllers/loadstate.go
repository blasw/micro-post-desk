package controllers

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
)

func GetLoadState() gin.HandlerFunc {
	return func(c *gin.Context) {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			log.Error("Unable to get cpu load", "err", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		log.Debug("CPU load", "percent", percent)
		c.JSON(200, percent)
	}
}
