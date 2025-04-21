// pingpong.go
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MatchLog struct {
	Time        string `bson:"time" json:"time"`
	Player      string `bson:"player" json:"player"`
	Power       int    `bson:"power" json:"power"`
	Goroutine   string `bson:"goroutine" json:"goroutine"`
	MatchNumber int    `bson:"match_number" json:"match_number"`
	Duration    int64  `bson:"duration_ms" json:"duration_ms"`
}

var (
	matchNumber     int
	logMutex        sync.Mutex
	matchCollection *mongo.Collection
	mongoClient     *mongo.Client
	redisClient     *redis.Client
	playerAChan     chan int
	playerBChan     chan int
	gameOverChan    chan struct{}
	wg              sync.WaitGroup
	csvFile         *os.File
	csvWriter       *csv.Writer
)

func init() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)

}

func connectMongoDB() error {
	uri := "mongodb+srv://singha20032546:singha2003@cluster0.jmwhrnj.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("ข้อผิดพลาดในการเชื่อมต่อ MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("Ping ไปยัง MongoDB ล้มเหลว: %v", err)
	}

	mongoClient = client
	matchCollection = client.Database("pingpongdb").Collection("matches")
	fmt.Println("✅ เชื่อมต่อ MongoDB สำเร็จ")
	return nil
}

func connectRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("ข้อผิดพลาดในการเชื่อมต่อ Redis: %v", err)
	}

	fmt.Println("✅ เชื่อมต่อ Redis สำเร็จ")
	return nil
}

func initCSVFile() error {
	var err error

	if _, err = os.Stat("match_log.csv"); os.IsNotExist(err) {
		csvFile, err = os.Create("match_log.csv")
		if err != nil {
			return fmt.Errorf("ไม่สามารถสร้างไฟล์ CSV ได้: %v", err)
		}

		csvWriter = csv.NewWriter(csvFile)
		headers := []string{"เวลา", "ผู้เล่น", "พลัง", "Goroutine", "หมายเลขการแข่งขัน", "ระยะเวลา (มิลลิวินาที)"}
		if err := csvWriter.Write(headers); err != nil {
			return fmt.Errorf("ไม่สามารถเขียนหัวข้อใน CSV ได้: %v", err)
		}
		csvWriter.Flush()
	} else {
		csvFile, err = os.OpenFile("match_log.csv", os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("ไม่สามารถเปิดไฟล์ CSV ได้: %v", err)
		}
		csvWriter = csv.NewWriter(csvFile)

		file, err := os.Open("match_log.csv")
		if err != nil {
			return fmt.Errorf("ไม่สามารถเปิดไฟล์ CSV เพื่ออ่าน: %v", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("ไม่สามารถอ่านข้อมูลจาก CSV ได้: %v", err)
		}

		maxMatch := 0
		for i, record := range records {
			if i == 0 {
				continue
			}
			if len(record) >= 5 {
				var currentMatch int
				if _, err := fmt.Sscanf(record[4], "%d", &currentMatch); err == nil {
					if currentMatch > maxMatch {
						maxMatch = currentMatch
					}
				}
			}
		}
		matchNumber = maxMatch
	}

	return nil
}

func logToCSV(entry MatchLog) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	record := []string{
		entry.Time,
		entry.Player,
		fmt.Sprintf("%d", entry.Power),
		entry.Goroutine,
		fmt.Sprintf("%d", entry.MatchNumber),
		fmt.Sprintf("%d", entry.Duration),
	}

	if err := csvWriter.Write(record); err != nil {
		return fmt.Errorf("ไม่สามารถเขียนข้อมูลลงใน CSV ได้: %v", err)
	}
	csvWriter.Flush()
	return nil
}

func logToMongoDB(entry MatchLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := matchCollection.InsertOne(ctx, entry); err != nil {
		log.Printf("⚠️ การบันทึกลง MongoDB ล้มเหลว: %v", err)
	}
}

func logToRedis(entry MatchLog) {
	ctx := context.Background()
	pipe := redisClient.Pipeline()

	pipe.Incr(ctx, "pingpong:total_matches")
	pipe.ZIncrBy(ctx, "pingpong:player_scores", 1, entry.Player)
	pipe.LPush(ctx, "pingpong:recent_hits", fmt.Sprintf("%s|%d|%s|%d", entry.Player, entry.Power, entry.Time, entry.MatchNumber))
	pipe.LTrim(ctx, "pingpong:recent_hits", 0, 9)
	pipe.Publish(ctx, "pingpong:updates", fmt.Sprintf("%s ยิงพลัง %d ในการแข่งขัน %d", entry.Player, entry.Power, entry.MatchNumber))

	matchData, _ := json.Marshal(entry)
	pipe.Set(ctx, fmt.Sprintf("pingpong:match:%d", entry.MatchNumber), matchData, 0)
	pipe.Set(ctx, "pingpong:last_match", matchData, 0)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("⚠️ Redis pipeline ล้มเหลว: %v", err)
	}

	currentMax, err := redisClient.Get(ctx, "pingpong:max_power").Int()
	if err == redis.Nil || entry.Power > currentMax {
		redisClient.Set(ctx, "pingpong:max_power", entry.Power, 0)
	}
}

func table(power int) int {
	return power * (rand.Intn(21) + 70) / 100
}

func playerA() {
	defer wg.Done()

	for {
		select {
		case power, ok := <-playerBChan:
			if !ok {
				return
			}
			start := time.Now()
			result := table(power)
			duration := time.Since(start).Milliseconds()

			entry := MatchLog{start.Format(time.RFC3339Nano), "A", result, "playerA", matchNumber, duration}
			logToCSV(entry)
			logToMongoDB(entry)
			logToRedis(entry)

			fmt.Printf("🎾 A ยิงด้วยพลัง %d (รับจาก B: %d)\n", result, power)
			playerAChan <- result

		case <-gameOverChan:
			return
		}
	}
}

func playerB() {
	defer wg.Done()

	for {
		select {
		case power, ok := <-playerAChan:
			if !ok {
				return
			}
			start := time.Now()
			myPower := rand.Intn(51) + 80
			duration := time.Since(start).Milliseconds()

			entry := MatchLog{start.Format(time.RFC3339Nano), "B", myPower, "playerB", matchNumber, duration}
			logToCSV(entry)
			logToMongoDB(entry)
			logToRedis(entry)

			fmt.Printf("🏓 B รับพลัง %d และตอบกลับด้วยพลัง %d\n", power, myPower)

			if myPower > power {
				playerBChan <- table(myPower)
			} else {
				fmt.Println("❌ B แพ้การแข่งขัน")
				closeChannels()
				return
			}
		case <-gameOverChan:
			return
		}
	}
}

func closeChannels() {
	close(playerAChan)
	close(playerBChan)
	close(gameOverChan)
}

func startGame() {
	logMutex.Lock()
	matchNumber++
	logMutex.Unlock()

	initial := rand.Intn(51) + 100
	fmt.Printf("\n🔔 เริ่มการแข่งขัน %d ด้วยพลังเริ่มต้น %d\n", matchNumber, initial)
	playerBChan <- initial
}

func resetGame() {
	playerAChan = make(chan int, 1)
	playerBChan = make(chan int, 1)
	gameOverChan = make(chan struct{})

	wg.Add(2)
	go playerA()
	go playerB()

	startGame()
	wg.Wait()
}

func playerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "🏓 เซิร์ฟเวอร์ผู้เล่น (Port 8888) กำลังทำงาน")
}

func tableHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "🎯 เซิร์ฟเวอร์โต๊ะ (Port 8889) กำลังทำงาน")
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	totalMatches, _ := redisClient.Get(ctx, "pingpong:total_matches").Result()
	maxPower, _ := redisClient.Get(ctx, "pingpong:max_power").Result()
	recentHits, _ := redisClient.LRange(ctx, "pingpong:recent_hits", 0, -1).Result()
	playerScores, _ := redisClient.ZRevRangeWithScores(ctx, "pingpong:player_scores", 0, -1).Result()

	fmt.Fprintf(w, "📊 สถิติการแข่งขัน:\nจำนวนการแข่งขันทั้งหมด: %s\nพลังสูงสุด: %s\n\n", totalMatches, maxPower)
	fmt.Fprintf(w, "🏅 คะแนนผู้เล่น:\n")
	for _, p := range playerScores {
		fmt.Fprintf(w, "- %s: %.0f ครั้ง\n", p.Member, p.Score)
	}
	fmt.Fprintf(w, "\n🕒 การยิงล่าสุด:\n")
	for _, hit := range recentHits {
		fmt.Fprintf(w, "- %s\n", hit)
	}
}

func matchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()

	lastMatchData, err := redisClient.Get(ctx, "pingpong:last_match").Result()
	if err != nil {

		if err == redis.Nil {

			opts := options.FindOne().SetSort(bson.D{{"match_number", -1}})
			var result MatchLog
			if err := matchCollection.FindOne(ctx, bson.D{}, opts).Decode(&result); err != nil {
				http.Error(w, fmt.Sprintf("ไม่พบข้อมูลแมตช์ล่าสุด: %v", err), http.StatusNotFound)
				return
			}
			resultJSON, _ := json.Marshal(result)
			w.Write(resultJSON)
			return
		}
		http.Error(w, fmt.Sprintf("เกิดข้อผิดพลาดในการดึงข้อมูล: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(lastMatchData))
}

func matchIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "URL ไม่ถูกต้อง", http.StatusBadRequest)
		return
	}
	matchID := parts[2]

	matchNum, err := strconv.Atoi(matchID)
	if err != nil {
		http.Error(w, "รูปแบบ ID ไม่ถูกต้อง", http.StatusBadRequest)
		return
	}

	matchData, err := redisClient.Get(ctx, fmt.Sprintf("pingpong:match:%d", matchNum)).Result()
	if err != nil && err != redis.Nil {
		http.Error(w, fmt.Sprintf("เกิดข้อผิดพลาดในการดึงข้อมูลจาก Redis: %v", err), http.StatusInternalServerError)
		return
	}

	if err != redis.Nil {
		w.Write([]byte(matchData))
		return
	}

	filter := bson.D{{"match_number", matchNum}}
	var result []MatchLog
	cursor, err := matchCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("เกิดข้อผิดพลาดในการค้นหาข้อมูล: %v", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result); err != nil {
		http.Error(w, fmt.Sprintf("เกิดข้อผิดพลาดในการอ่านข้อมูล: %v", err), http.StatusInternalServerError)
		return
	}

	if len(result) == 0 {
		http.Error(w, fmt.Sprintf("ไม่พบข้อมูลแมตช์หมายเลข %d", matchNum), http.StatusNotFound)
		return
	}

	resultJSON, _ := json.Marshal(result)
	w.Write(resultJSON)
}

func main() {
	if err := connectMongoDB(); err != nil {
		log.Fatalf("❌ ข้อผิดพลาด MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	if err := connectRedis(); err != nil {
		log.Fatalf("❌ ข้อผิดพลาด Redis: %v", err)
	}
	defer redisClient.Close()

	if err := initCSVFile(); err != nil {
		log.Fatalf("❌ ข้อผิดพลาด CSV: %v", err)
	}
	defer func() {
		csvWriter.Flush()
		csvFile.Close()
	}()

	http.HandleFunc("/player", playerHandler)
	http.HandleFunc("/table", tableHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/match", matchHandler)
	http.HandleFunc("/match/", matchIDHandler)

	go func() {
		log.Println("🚀 เซิร์ฟเวอร์ผู้เล่นกำลังทำงานที่ http://localhost:8888")
		log.Fatal(http.ListenAndServe(":8887", nil))
	}()
	go func() {
		log.Println("📡 เซิร์ฟเวอร์โต๊ะกำลังทำงานที่ http://localhost:8889")
		log.Fatal(http.ListenAndServe(":8886", nil))
	}()

	for {
		resetGame()
		fmt.Println("\n➡️ กด Enter เพื่อเริ่มการแข่งขันใหม่...")
		fmt.Scanln()
	}
}
