package model

type QueryPayload struct {
	TopK        int       `json:"top"`
	Vector      []float32 `json:"vector"`
	Space       string    `json:"space"`
	WithPayLoad bool      `json:"with_payload"`
}

type QdrantResp struct {
	Result []struct {
		Version int     `json:"version"`
		Score   float32 `json:"score"`
		Payload struct {
			Doc_id     string `json:"doc_id"`
			Extra_info string `json:"extra_info"`
			File_name  string `json:"file_name""`
			Sub_title  string `json:"sub_title"`
			Text       string `json:"text"`
		} `json:"payload"`
		Vector any `json:"vector"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float32 `json:"time"`
}

type QdrantQAResp struct {
	Result []struct {
		Version int     `json:"version"`
		Score   float32 `json:"score"`
		Payload struct {
			Doc_id    string `json:"doc_id"`
			File_name string `json:"file_name""`
			Sub_title string `json:"sub_title"`
			Q         string `json:"Q"`
			A         string `json:"A"`
		} `json:"payload"`
		Vector any `json:"vector"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float32 `json:"time"`
}

type Query struct {
	Question string `json:"question" binding:"required"`
	UseCache bool   `json:"useCache"`
	DbName   string `json:"dbName" binding:"required"`
}

type Insert struct {
	DbName  string `json:"dbName" binding:"required"`
	DirName string `json:"dirName" binding:"required"`
}

type DeleteDB struct {
	DbName string `json:"dbName" binding:"required"`
}
