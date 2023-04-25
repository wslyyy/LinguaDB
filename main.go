package main

import (
	"LinguaDB/initialization"
	"LinguaDB/lingua"
	"LinguaDB/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"log"
	"net/http"
)

// question := "司龄4年，工龄25年，我有多少天年假"
// question2 := "我迟到了半小时会怎么样"
// question3 := "中午几点点外卖比较合适"
// question4 := "公司可以抽烟吗"
// question5 := "内部分享是什么"
// question6 := "我想联系HR或者行政，他们叫什么名字，怎么联系"
// question7 := "什么时候发工资"
// question8 := "公司的组织架构"
// question9 := "公司有哪些福利呀"
// question10 := "产假怎么休？"
// question11 := "应届生刚工作如何适应"
// question12 := "我入职5个月了，我的团建费攒多少了"
var config *initialization.Config

func main() {
	var (
		cfg = pflag.StringP("config", "c", "./config.yaml", "config file path")
	)
	pflag.Parse()
	config = initialization.LoadConfig(*cfg)

	//lingua.LoadDOC(*config)

	r := gin.Default()
	r.GET("/ping", PingHandler)
	r.POST("/insert", InsertHandler)
	r.POST("/query", QueryHandler)
	r.POST("/deleteDB", DeleteDBHandler)

	err := initialization.StartHTTPServer(*config, r)
	if err != nil {
		log.Printf("failed to start server: %v", err)
	}
}

func PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func QueryHandler(context *gin.Context) {
	var query model.Query
	query.UseCache = false
	err := context.ShouldBindJSON(&query)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}
	fmt.Printf("info: %#v\n", query)
	question := query.Question
	useCache := query.UseCache
	dbName := query.DbName
	answer, err := lingua.Query(*config, question, useCache, dbName)
	if err != nil {
		log.Printf("回答失败：%v", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"answer": answer,
	})
}

func InsertHandler(context *gin.Context) {
	var insert model.Insert
	err := context.ShouldBindJSON(&insert)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}
	fmt.Printf("info: %#v\n", insert)
	dbName := insert.DbName
	dirName := insert.DirName
	err = lingua.LoadDOC(*config, dbName, dirName)
	if err != nil {
		log.Printf("插入失败：%v", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"info": "文档插入成功",
	})
}

func DeleteDBHandler(context *gin.Context) {
	var deleteDB model.DeleteDB
	err := context.ShouldBindJSON(&deleteDB)
	if err != nil {
		fmt.Println(err.Error())
		context.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}
	fmt.Printf("info: %#v\n", deleteDB)
	dbName := deleteDB.DbName
	err = lingua.DeleteDB(*config, dbName)
	if err != nil {
		log.Printf("删除目标库失败：%v", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"info": "删除目标库成功",
	})
}
