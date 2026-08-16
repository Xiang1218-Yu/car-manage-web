// Command carmanageweb 启动停车场管理系统 Web 服务。
//
// 用法：
//
//	carmanageweb [-addr :8080] [-db parking.db]
//
// 数据库默认为当前目录下的 parking.db，不存在会自动创建并建表。
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"carmanageweb/internal/server"
	"carmanageweb/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath := flag.String("db", "parking.db", "SQLite 数据库文件路径")
	seed := flag.Bool("seed", false, "首次启动时写入演示数据（仅在数据库为空时生效）")
	flag.Parse()

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer st.Close()

	if *seed {
		if err := seedDemo(st); err != nil {
			log.Printf("写入演示数据失败: %v", err)
		}
	}

	srv, err := server.New(st)
	if err != nil {
		log.Fatalf("初始化服务失败: %v", err)
	}

	httpSrv := &http.Server{
		Addr:              *addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 优雅关闭
	go func() {
		log.Printf("停车场管理系统已启动：http://localhost%s", *addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务退出: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("强制关闭: %v", err)
	}
	log.Println("已退出")
}
