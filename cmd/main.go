package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"tt-newbotsacc/internal/usecase"
	"tt-newbotsacc/pkg/browser"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("dev.env")

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/rpa_db?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Cannot unparse string:%v\n", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Cannot create pool of connections:%v\n", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Ping failed:%v\n", err)
	}

	TaskUseCase := usecase.NewTaskUseCase(pool)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	const maxCount = 4

	for i := 1; i < maxCount; i++ {
		wg.Add(1)
		go func(workerID int, ctx context.Context) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return 
				default:
				}
				func() {
					newCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					task, err := TaskUseCase.ClaimNewTask(newCtx)
					if task == nil && err == nil {
						select {
						case <-ctx.Done():
							return 
						case <-time.After(4 * time.Second): 
						}
					} else if err != nil {
						fmt.Printf("[worker:%d] Error:%v\n", workerID, err)
						select {
						case <-ctx.Done():
							return 
						case <-time.After(3 * time.Second): 
						}
					} else {
						fmt.Printf("[worker:%d] New task with ID:%v and status:%s\n", workerID, task.ID, task.Status.TaskStatus)

						var proxyAddr string
						if task.ProxyAddress.Valid {
							proxyAddr = task.ProxyAddress.String
							fmt.Printf("[worker:%d] Using proxy: %s\n", workerID, proxyAddr)
						} else {
							fmt.Printf("[worker:%d] Running without proxy\n", workerID)
						}

						b, err := browser.NewBrowser(proxyAddr)
						if err != nil {
							fmt.Printf("[worker:%d] Browser launch failed: %v", workerID, err)
							return
						}
						defer b.Close()

						err = b.Navigate("https://httpbin.org/ip")
						if err != nil {
							fmt.Printf("[worker:%d] Navigation failed: %v", workerID, err)
							return
						}

						if rand.Intn(10) < 8 {
							updateCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
							err := TaskUseCase.CompleteTask(updateCtx, task.ID)
							cancel()

							if err != nil {
								fmt.Printf("[worker:%d] Error completing task: %v\n", workerID, err)
							} else {
								fmt.Printf("[worker:%d] Successfully completed\n", workerID)
							}
						} else {
							updateCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
							err := TaskUseCase.FailTask(updateCtx, task.ID, "simulated error")
							cancel()

							if err != nil {
								fmt.Printf("[worker:%d] Error failing task: %v\n", workerID, err)
							} else {
								fmt.Printf("[worker:%d] Successfully failed\n", workerID)
							}
						}
					}
				}()
			}
		}(i, workerCtx)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Println("[Main] All workers are working. For exit press Ctrl+C")

	<-stop
	cancelWorkers()
	wg.Wait()

	fmt.Println("\n[Main] Closing DB")
}
