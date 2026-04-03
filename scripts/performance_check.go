package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite3", "./data/rssreader.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// 测试查询
	queries := []struct {
		name  string
		query string
		args  []interface{}
	}{
		{
			name:  "按feed_id查询文章",
			query: "SELECT COUNT(*) FROM articles WHERE feed_id = ?",
			args:  []interface{}{5},
		},
		{
			name:  "按feed_id和is_read查询文章",
			query: "SELECT COUNT(*) FROM articles WHERE feed_id = ? AND is_read = ?",
			args:  []interface{}{5, 0},
		},
		{
			name:  "按用户查询所有文章",
			query: "SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ?",
			args:  []interface{}{1},
		},
		{
			name:  "查询未读文章",
			query: "SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ? AND a.is_read = 0",
			args:  []interface{}{1},
		},
		{
			name:  "查询starred文章",
			query: "SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ? AND a.is_starred = 1",
			args:  []interface{}{1},
		},
	}

	fmt.Println("性能测试结果:")
	fmt.Println("================")

	for _, q := range queries {
		start := time.Now()
		var count int
		err := db.QueryRow(q.query, q.args...).Scan(&count)
		duration := time.Since(start)
		
		if err != nil {
			fmt.Printf("❌ %s: 错误 - %v\n", q.name, err)
		} else {
			fmt.Printf("✅ %s: %d 条记录, 耗时: %v\n", q.name, count, duration)
		}
	}

	// 测试EXPLAIN QUERY PLAN
	fmt.Println("\n查询计划分析:")
	fmt.Println("================")
	
	explainQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "按feed_id查询",
			query: "EXPLAIN QUERY PLAN SELECT * FROM articles WHERE feed_id = 5 ORDER BY published_at DESC LIMIT 50",
		},
		{
			name:  "按feed_id和is_read查询",
			query: "EXPLAIN QUERY PLAN SELECT * FROM articles WHERE feed_id = 5 AND is_read = 0 ORDER BY published_at DESC LIMIT 50",
		},
	}

	for _, eq := range explainQueries {
		fmt.Printf("\n%s:\n", eq.name)
		rows, err := db.Query(eq.query)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}
		
		for rows.Next() {
			var id, parent, notused int
			var detail string
			rows.Scan(&id, &parent, &notused, &detail)
			fmt.Printf("  %s\n", detail)
		}
		rows.Close()
	}
}