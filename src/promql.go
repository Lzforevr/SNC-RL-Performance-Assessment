package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	Host = "http://localhost:9090"
  StorePath = "/home/user/output/09/Future"
)

type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type Standardized struct {
	Metric  string
	Pod2Val map[string]float64
}

func main() {
  // 定义错误输出日志
  logPath, err:= os.OpenFile(StorePath+"/promql.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
  if err != nil {
    log.Fatal(err)
  }
  defer logPath.Close()
  log.SetOutput(logPath)
  
	// 定义要查询的指标和 Prometheus 查询
	queries := map[string]string{
		"PodMemoryUsageBytes":   `sum(container_memory_working_set_bytes{container!="",pod!=""}) by (pod)`,
		"PodMemoryUsageRequest": `sum(container_memory_working_set_bytes{container!="",pod!=""}) by (pod) / sum(kube_pod_container_resource_requests{container!="", resource="memory", unit="byte"}) by (pod) * 100`,
		"PodNetworkIn":          `sum(avg(rate(container_network_receive_bytes_total{pod!=""}[1m])) by (pod)) by (pod)`,
		"PodNetworkOut":         `sum(avg(rate(container_network_transmit_bytes_total{pod!=""}[1m])) by (pod)) by (pod)`,
		"PodCPUCoreUsage":       `sum(rate(container_cpu_usage_seconds_total{container!="",pod!=""}[1m])) by (pod)`,
		"PodCPUUsageRequest":    `sum(rate(container_cpu_usage_seconds_total{container!="",pod!=""}[1m])) by (pod) / (sum(kube_pod_container_resource_requests{container!="",pod!="",resource="cpu"}) by (pod)) * 100`,
	}

	// 创建一个 channel 来收集查询结果
	resultChan := make(chan Standardized, len(queries))

	var wg sync.WaitGroup
	// 使用 context 来处理进程取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用 Ticker 来每 10 秒执行一次查询
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 设置信号通道
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 命令参数读取
	Epoch := 5
	if len(os.Args) > 1 {
		epoch, _ := strconv.ParseInt(os.Args[1], 10, 8)
		Epoch = int(epoch)
		if Epoch < 1 || Epoch > 128 {
			fmt.Println("")
		}
	}
	for cnt := 0; cnt < Epoch; cnt++ {
		select {
		case <-ticker.C:
			time = time.Now().Format(time.RFC3339)
			timestamp := time[11:len(time)-1]
      date := time.Now().Format("2006-01-02")
      newDir(StorePath)
      newDir(StorePath+"/"+date)
			for metric, query := range queries {
				wg.Add(1)
				go PromQuery(ctx, Host, metric, query, &wg, resultChan)
			}

			wg.Wait()

			for i := 0; i < len(queries); i++ {
				resultData := <-resultChan
				metric := resultData.Metric

				outputFilePath := StorePath + "/" + date + "/" + metric + ".csv"
				headers := []string{"timestamp", "pod", "value"}
				if err := ensureHeader(outputFilePath, headers); err != nil {
					log.Fatalf("Error ensuring header for %v: %v", outputFilePath, err)
				}

				outputFile, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_RDWR, 0666)
				if err != nil {
					log.Fatalf("Try to open %v.CSV error: %v", metric, err)
				}

				writer := csv.NewWriter(outputFile)
				for pod, value := range resultData.Pod2Val {
					writeCSV(writer, timestamp, pod, value)
				}
				writer.Flush()
				outputFile.Close()
			}
		case <-signalChan:
			fmt.Println("Received interrupt, cleaning up and exiting...")
			cancel()          // 发送取消信号给所有协程
			wg.Wait()         // 等待所有协程退出
			close(resultChan) // 关闭 channel
			return
		}
		fmt.Printf("Round %v finished;", cnt+1)
	}
}

// 检查文件是否存在头部并创建
func ensureHeader(filePath string, headers []string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// 写入文件头
		writer := csv.NewWriter(file)
		defer writer.Flush()
		return writer.Write(headers)
	}
	return nil
}

// ParseToFloat64 将值（字符串）转换为 float64。
func ParseToFloat64(value interface{}) float64 {
	parsedValue, err := strconv.ParseFloat(value.(string), 64)
	if err != nil {
		return -1 // 如果转换失败，返回 -1
	}
	return parsedValue
}

// PromQuery 查询 Prometheus 并返回标准化的结果。
func PromQuery(ctx context.Context, promServer, metric, query string, wg *sync.WaitGroup, resultChan chan<- Standardized) {
	defer wg.Done()

	// 对查询进行 URL 编码
	queryEncoded := url.QueryEscape(query)
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", promServer, queryEncoded)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for %s: %v", metric, err)
		return
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Fetch HTTP Response error for %s: %v", metric, err)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading Prometheus error for %s: %v", metric, err)
		return
	}

	// 将响应解析为结构体
	var result PrometheusQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Parsing Prometheus error for %s: %v", metric, err)
		return
	}

	// 创建一个标准化的结果 map
	NormResult := Standardized{Metric: metric, Pod2Val: make(map[string]float64)}
	if result.Status == "success" {
		for _, r := range result.Data.Result {
			val := ParseToFloat64(r.Value[1])
			if val == -1 {
				continue
			}
			NormResult.Pod2Val[r.Metric["pod"]] = val
		}
	}

	// 将结果发送回 channel
	select {
	case resultChan <- NormResult:
	case <-ctx.Done():
		log.Printf("Context canceled, dropping result for %s", metric)
	}
}

// writeCSV 向 CSV 文件写入一行数据，并在首次写入时写入标题行。
func writeCSV(writer *csv.Writer, timestamp string, pod string, value float64) {

	row := []string{timestamp, pod, fmt.Sprintf("%f", value)}
	if err := writer.Write(row); err != nil {
		log.Fatalf("Writing CSV rows error: %v", err)
	}
}

// newDir 以日期为间隔，分隔不同数据集
func newDir(dirName string) {
  _,err := os.Stat(dirName)
  if os.IsNotExist(err) {
    err := os.MkdirAll(dirName,0777)
    if err != nil {
      log.Fatalf("Creating Directory error: %v",err)
      }
    }
}