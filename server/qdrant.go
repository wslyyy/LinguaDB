package server

import (
	"LinguaDB/model"
	"LinguaDB/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Qdrant struct {
	addrGrpc       string
	addrHttp       string
	collectionName string
	vectorSize     uint64
}

var (
	distance = pb.Distance_Cosine
)

func NewQdrant(addrGrpc string, addrHttp string, collectionName string, vectorSize uint64) *Qdrant {
	return &Qdrant{
		addrGrpc:       addrGrpc,
		addrHttp:       addrHttp,
		collectionName: collectionName,
		vectorSize:     vectorSize,
	}
}

func (q *Qdrant) CreateCollection() error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return err
	}
	defer conn.Close()

	// create grpc collection client
	collections_client := pb.NewCollectionsClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	// Check Qdrant version
	qdrantClient := pb.NewQdrantClient(conn)
	healthCheckResult, err := qdrantClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		log.Printf("Could not get health: %v", err)
		return err
	} else {
		log.Printf("Qdrant version: %s", healthCheckResult.GetVersion())
	}

	res, err := q.GetList()
	if err != nil {
		return err
	}
	for _, r := range res {
		if r.GetName() == q.collectionName {
			log.Printf("%s already exeist!", q.collectionName)
			return nil
		}
	}

	// Delete collection
	_, err = collections_client.Delete(ctx, &pb.DeleteCollection{
		CollectionName: q.collectionName,
	})
	if err != nil {
		log.Printf("Could not delete collection: %v", err)
		return err
	} else {
		log.Println("Collection", q.collectionName, "deleted")
	}

	// Create new collection
	_, err = collections_client.Create(ctx, &pb.CreateCollection{
		CollectionName: q.collectionName,
		VectorsConfig: &pb.VectorsConfig{Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     q.vectorSize,
				Distance: distance,
			},
		}},
	})
	if err != nil {
		log.Printf("Could not create collection: %v", err)
		return err
	} else {
		log.Println("Collection", q.collectionName, "created")
	}
	return nil
}

func (q *Qdrant) DeleteCollection() error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return err
	}
	defer conn.Close()

	// create grpc collection client
	collections_client := pb.NewCollectionsClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	// Check Qdrant version
	qdrantClient := pb.NewQdrantClient(conn)
	healthCheckResult, err := qdrantClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		log.Printf("Could not get health: %v", err)
		return err
	} else {
		log.Printf("Qdrant version: %s", healthCheckResult.GetVersion())
	}

	// Delete collection
	_, err = collections_client.Delete(ctx, &pb.DeleteCollection{
		CollectionName: q.collectionName,
	})
	if err != nil {
		log.Printf("Could not delete collection: %v", err)
		return err
	} else {
		log.Println("Collection", q.collectionName, "deleted")
	}
	return nil
}

func (q *Qdrant) UpsertQAEmbeddingToQdrant(chunk QAChunk) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return err
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = q.CreateCollection()
	if err != nil {
		return err
	}

	pointsClient := pb.NewPointsClient(conn)
	uuid := utils.GenUUID()

	payload := map[string]*pb.Value{
		"doc_id": {
			Kind: &pb.Value_StringValue{StringValue: fmt.Sprintf("doc_id-%s", uuid)},
		},
		"file_name": {
			Kind: &pb.Value_StringValue{StringValue: chunk.Title},
		},
		"sub_title": {
			Kind: &pb.Value_StringValue{StringValue: chunk.SubTitle},
		},
		"Q": {
			Kind: &pb.Value_StringValue{StringValue: chunk.Q},
		},
		"A": {
			Kind: &pb.Value_StringValue{StringValue: chunk.A},
		},
	}

	waitUpsert := true
	var upsertPoints []*pb.PointStruct

	upsertPoints = append(upsertPoints, &pb.PointStruct{
		Id: &pb.PointId{
			//PointIdOptions: &pb.PointId_Num{Num: 1},
			PointIdOptions: &pb.PointId_Uuid{Uuid: uuid},
		},
		Vectors: &pb.Vectors{
			VectorsOptions: &pb.Vectors_Vector{
				Vector: &pb.Vector{Data: chunk.QEmbedding},
			},
		},
		Payload: payload,
	})

	_, err = pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: q.collectionName,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})

	if err != nil {
		log.Printf("Could not upsert points: %v", err)
		return err
	} else {
		log.Println("Upsert", len(upsertPoints), "points")
	}
	return nil
}

func (q *Qdrant) UpsertEmbeddingsToQdrant(embeddings [][]float32, chunks []Chunk) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return err
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = q.CreateCollection()
	if err != nil {
		return err
	}

	pointsClient := pb.NewPointsClient(conn)

	//upsertPoints := make([]*pb.PointStruct, len(embeddings))
	var upsertPoints []*pb.PointStruct

	for i, embedding := range embeddings {
		chunk := chunks[i]
		uuid := utils.GenUUID()

		payload := map[string]*pb.Value{
			"doc_id": {
				Kind: &pb.Value_StringValue{StringValue: fmt.Sprintf("doc_id-%s", uuid)},
			},
			"file_name": {
				Kind: &pb.Value_StringValue{StringValue: chunk.Title},
			},
			"sub_title": {
				Kind: &pb.Value_StringValue{StringValue: chunk.SubTitle},
			},
			"text": {
				Kind: &pb.Value_StringValue{StringValue: chunk.Text},
			},
			"extra_info": {
				Kind: &pb.Value_StringValue{StringValue: ""},
			},
		}

		upsertPoints = append(upsertPoints, &pb.PointStruct{
			Id: &pb.PointId{
				//PointIdOptions: &pb.PointId_Num{Num: 1},
				PointIdOptions: &pb.PointId_Uuid{Uuid: uuid},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{
					Vector: &pb.Vector{Data: embedding},
				},
			},
			Payload: payload,
		})
	}
	fmt.Println(len(embeddings))
	fmt.Println(len(upsertPoints))
	//fmt.Println(upsertPoints)

	//fmt.Println(upsertPoints[0])
	//fmt.Println(upsertPoints[1])
	//fmt.Printf("v1 type:%T\n", embeddings[0])
	waitUpsert := true
	maxVectorsPerRequest := 100

	for i := 0; i < len(upsertPoints); i += maxVectorsPerRequest {
		end := i + maxVectorsPerRequest
		if end > len(upsertPoints) {
			end = len(upsertPoints)
		}

		_, err = pointsClient.Upsert(ctx, &pb.UpsertPoints{
			CollectionName: q.collectionName,
			Wait:           &waitUpsert,
			Points:         upsertPoints[i:end],
		})

		if err != nil {
			log.Printf("Could not upsert points: %v", err)
			return err
		} else {
			log.Println("Upsert", len(upsertPoints), "points")
		}
	}
	return nil
}

// RetrievePointsByIds NOT USE
func (q *Qdrant) RetrievePointsByIds() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	pointsClient := pb.NewPointsClient(conn)
	// Retrieve points by ids
	pointsById, err := pointsClient.Get(ctx, &pb.GetPoints{
		CollectionName: q.collectionName,
		Ids: []*pb.PointId{
			{PointIdOptions: &pb.PointId_Uuid{Uuid: "0795e52f-4557-4201-8d59-984715f58ae8"}},
			{PointIdOptions: &pb.PointId_Uuid{Uuid: "210fe367-a2c2-4413-8e6b-2065d5d275cd"}},
		},
	})
	if err != nil {
		log.Printf("Could not retrieve points: %v", err)
	} else {
		log.Printf("Retrieved points: %s", pointsById.GetResult())
	}
}

// Search NOT USE
func (q *Qdrant) Search() []*pb.ScoredPoint {
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	pointsClient := pb.NewPointsClient(conn)
	// Unfiltered search
	unfilteredSearchResult, err := pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: q.collectionName,
		Vector:         utils.GenRandEmbedding(int(q.vectorSize)),
		Limit:          3,
		// Include all payload and vectors in the search result
		WithVectors: &pb.WithVectorsSelector{SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true}},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		log.Printf("Could not search points: %v", err)
	} else {
		log.Printf("Found points: %s", unfilteredSearchResult.GetResult())
	}
	res := unfilteredSearchResult.GetResult()
	return res
}

func (q *Qdrant) SearchHttp(vector []float32) (*model.QdrantResp, error) {
	err := q.CreateCollection()
	if err != nil {
		return nil, err
	}
	qdEndpoint := fmt.Sprintf("http://%s/collections/%s/points/search", q.addrHttp, q.collectionName)

	query := model.QueryPayload{
		TopK:        2,
		Vector:      vector,
		Space:       "cosine",
		WithPayLoad: true,
	}
	payload, err := json.Marshal(query)
	if err != nil {
		log.Printf("Error encoding query: %s\n", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", qdEndpoint, bytes.NewReader(payload))
	if err != nil {
		log.Printf("Error creating request: %s\n", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {

		if err != nil {
			log.Printf("Error decoding response: %s\n", err)
			return nil, err
		}
		var r model.QdrantResp
		err = json.Unmarshal(respBytes, &r)
		//fmt.Println(r)
		//fmt.Println(r.Result[0].Score)

		var out bytes.Buffer
		err := json.Indent(&out, respBytes, "", "  ")
		return &r, err
	} else {
		log.Printf("Error response: %s\n", resp.Status)
		return nil, errors.New(fmt.Sprintf("http error code %d", resp.StatusCode))
	}
}

func (q *Qdrant) SearchQAQdrantHttp(vector []float32) (*model.QdrantQAResp, error) {
	err := q.CreateCollection()
	if err != nil {
		return nil, err
	}

	qdEndpoint := fmt.Sprintf("http://%s/collections/%s/points/search", q.addrHttp, q.collectionName)

	query := model.QueryPayload{
		TopK:        2,
		Vector:      vector,
		Space:       "cosine",
		WithPayLoad: true,
	}
	payload, err := json.Marshal(query)
	if err != nil {
		log.Printf("Error encoding query: %s\n", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", qdEndpoint, bytes.NewReader(payload))
	if err != nil {
		log.Printf("Error creating request: %s\n", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {

		if err != nil {
			log.Printf("Error decoding response: %s\n", err)
			return nil, err
		}
		var r model.QdrantQAResp
		err = json.Unmarshal(respBytes, &r)
		//fmt.Println(r)
		//fmt.Println(r.Result[0].Score)

		var out bytes.Buffer
		err := json.Indent(&out, respBytes, "", "  ")
		return &r, err
	} else {
		log.Printf("Error response: %s\n", resp.Status)
		return nil, errors.New(fmt.Sprintf("http error code %d", resp.StatusCode))
	}
}

func (q *Qdrant) DeletePoints(dirname string) error {
	qdEndpoint := fmt.Sprintf("http://%s/collections/%s/points/delete", q.addrHttp, q.collectionName)

	var mustList []model.Must
	mustList = append(mustList, model.Must{
		Key:   "file_name",
		Match: model.Match{
			Value: dirname,
		},
	})

	filter := model.MyFilter{
		Filter: model.Filter{
			Must: mustList,
		},
	}
	filterJson, err := json.Marshal(filter)

	fmt.Println(string(filterJson))

	if err != nil {
		log.Printf("Error encoding query: %s\n", err)
		return err
	}
	req, err := http.NewRequest("POST", qdEndpoint, bytes.NewReader(filterJson))
	if err != nil {
		log.Printf("Error creating request: %s\n", err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %s\n", err)
		return err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respBytes))

	if resp.StatusCode != http.StatusOK {
	    return errors.New(fmt.Sprintf("删除文件夹(%s)内容失败, 返回内容为: %s", dirname, string(respBytes)))
	}
	return nil
}

func (q *Qdrant) GetList() ([]*pb.CollectionDescription, error) {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(q.addrGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
		log.Printf("did not connect: %v", err)
	}
	defer conn.Close()

	collections_client := pb.NewCollectionsClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := collections_client.List(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		return nil, err
		log.Printf("could not get collections: %v", err)
	}
	log.Printf("List of collections: %s", r.GetCollections())
	return r.GetCollections(), nil
}
