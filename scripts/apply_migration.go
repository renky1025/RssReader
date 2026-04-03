package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run apply_migration.go <migration_file> [database_path]")
	}

	// 获取migration文件路径
	migrationPath := os.Args[1]
	
	// 获取数据库路径
	dbPath := "./data/rssreader.db"
	if len(os.Args) > 2 {
		dbPath = os.Args[2]
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// 读取migration文件
	content, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	// 执行migration
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	fmt.Printf("Migration %s applied successfully!\n", migrationPath)
	
	// 分析表以更新统计信息
	_, err = db.Exec("ANALYZE")
	if err != nil {
		log.Printf("Warning: Failed to analyze database: %v", err)
	} else {
		fmt.Println("Database analysis completed!")
	}
}