package storage

import (
	"LinguaDB/initialization"
	"LinguaDB/model"
	"LinguaDB/server"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

// 定义一个初始化数据库的函数
func initDB() (err error) {
	// DSN:Data Source Name
	dsn := "user:password@tcp(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True"
	// 不会校验账号密码是否正确
	// 注意！！！这里不要使用:=，我们是给全局变量赋值
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 尝试与数据库建立连接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

type QAStorage struct {
	Q          string
	QEmbedding []float32
	A          string
	AEmbedding []float32
	Title      string
	SubTitle   string
	DbName     string
}

func (QA *QAStorage) SaveToMysqlStorage() {
	err := initDB()
	if err != nil {
		fmt.Printf("init db failed,err:%v\n", err)
		return
	}
	// 预处理插入示例
	sqlStr := fmt.Sprintf("insert into %s(Q, QEmbedding, A, AEmbedding, Title, SubTitle) values (?,?,?,?,?,?)", QA.DbName)
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		fmt.Printf("prepare failed, err:%v\n", err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(QA.Q, QA.QEmbedding, QA.A, QA.AEmbedding, QA.Title, QA.SubTitle)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	fmt.Println("insert success.")

}

func (QA *QAStorage) createQdrantStorage(qdrant *server.Qdrant) error {
	err := qdrant.CreateCollection()
	if err != nil {
		return err
	}
	return nil
}

func (QA *QAStorage) SaveToQdrantStorage(config initialization.Config) error {
	qdrant := server.NewQdrant(config.QdrantAddrGrpc, config.QdrantAddrHttp, QA.DbName, uint64(1536))
	err := QA.createQdrantStorage(qdrant)
	if err != nil {
		return err
	}
	qachunk := server.QAChunk{
		Q:          QA.Q,
		A:          QA.A,
		QEmbedding: QA.QEmbedding,
		Title:      QA.Title,
		SubTitle:   QA.SubTitle,
	}
	qdrant.UpsertQAEmbeddingToQdrant(qachunk)
	return nil
}

func (QA *QAStorage) GetQAStorage(config initialization.Config) (*model.QdrantQAResp, error) {
	qdrant := server.NewQdrant(config.QdrantAddrGrpc, config.QdrantAddrHttp, QA.DbName, uint64(1536))
	err := QA.createQdrantStorage(qdrant)
	if err != nil {
		return nil, err
	}
	res, err := qdrant.SearchQAQdrantHttp(QA.QEmbedding)
	if err != nil {
		log.Fatalf("get topK QA  from qdrant failed: %v", err)
		return nil, err
	}
	if len(res.Result) > 0 {
		if res.Result[0].Score >= 0.90 {
			log.Printf("命中缓存QA，相似度为：%v", res.Result[0].Score)
			log.Printf("最为相似的问题为：%v", res.Result[0].Payload.Q)
			return res, nil
		} else {
			log.Printf("未命中缓存QA，相似度为%v", res.Result[0].Score)
		}
	}
	return nil, nil
}
