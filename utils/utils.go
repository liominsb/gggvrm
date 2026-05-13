package utils // Package utils 实用性

import (
	"context"
	"encoding/json"
	"fmt"
	"gggvrm/global"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

// 检查密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Setcache 设置缓存
func Setcache(ctx context.Context, key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)

	if err != nil {
		return err
	}
	a := time.Duration(rand.IntN(5) + 10)
	if err := global.RedisDB.Set(ctx, key, valueJSON, a*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// 同步点赞数和浏览数到数据库
func SyncSql(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("SyncSql 发生严重错误并恢复: %v\n", r)
			// 可以在这里做告警，并决定是否重新启动该协程
		}
	}()
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("SyncSql 任务已取消")
			return
		case <-ticker.C:
			syncFieldToDB(ctx, "likes")
			syncFieldToDB(ctx, "views")
			fmt.Println("已同步点赞数和浏览数到数据库")
		}
	}
}

func syncFieldToDB(ctx context.Context, field string) {
	var cursor uint64
	var keys []string
	var err error
	for {
		match := "article:*:" + field

		//keys只拿了点赞数的键名
		keys, cursor, err = global.RedisDB.Scan(ctx, cursor, match, 100).Result()
		if err != nil {
			fmt.Println("获取 Redis Keys 失败:", err)
			break
		}
		if len(keys) == 0 {
			if cursor == 0 {
				break
			}
		}
		v, err := global.RedisDB.MGet(ctx, keys...).Result()
		if err != nil {
			fmt.Println("批量获取 Redis 值失败:", err)
			break
		}

		updateData := make(map[int]int)

		for i, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				continue
			}

			articleIDStr := parts[1]
			articleID, err := strconv.Atoi(articleIDStr)
			if err != nil {
				continue
			}

			vstr, ok := v[i].(string)
			if !ok {
				continue
			}

			fields, err := strconv.Atoi(vstr)
			if err != nil {
				continue
			}
			updateData[articleID] = fields
		}

		if len(updateData) > 0 {
			var sqlBuilder strings.Builder
			var ids []int

			// 1. 拼接前半部分 (动态传入要更新的字段名)
			sqlBuilder.WriteString(fmt.Sprintf("UPDATE articles SET %s = CASE id ", field))

			// 2. 动态拼接 WHEN ... THEN ...
			for id, fields := range updateData {
				// 因为 id 和 fields 都是严格的 int 类型，直接 Sprintf 拼接不存在 SQL 注入风险
				sqlBuilder.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, fields))
				ids = append(ids, id)
			}

			// 3. 拼接收尾，加入 ELSE 防御和 WHERE 限制
			sqlBuilder.WriteString(fmt.Sprintf("ELSE %s END WHERE id IN (?)", field))

			// 4. 执行原生 SQL
			// global.Db.Exec 会自动把 ids 切片展开并替换掉 ?
			err := global.Db.Exec(sqlBuilder.String(), ids).Error
			if err != nil {
				fmt.Println("批量更新失败:", err)
			}
		}

		//if len(updateData) > 0 {
		//	err := global.Db.Transaction(func(tx *gorm.DB) error {
		//		for articleID, fields := range updateData {
		//			if err := tx.Model(&models.Article{}).Where("id = ?", articleID).Update(field, fields).Error; err != nil {
		//				log.Printf("更新文章 %d %s数失败: %v\n", articleID, field, err)
		//				return err
		//			}
		//		}
		//		return nil
		//	})
		//	if err != nil {
		//		fmt.Printf("批量更新数据库%s数失败: %v\n", field, err)
		//	} else {
		//		fmt.Printf("已同步%s数到数据库", field)
		//	}
		//}
		if cursor == 0 {
			break
		}
	}
}

// RandomExpiration 传入一个基础过期时间，返回增加 0~59 秒随机抖动后的时间
func RandomExpiration(baseTime time.Duration) time.Duration {
	// 使用 rand.Intn(60) 生成 0-59 的随机数，更加标准和易读
	jitter := time.Duration(rand.IntN(60)) * time.Second
	return baseTime + jitter
}

// 快速过滤掉字符串中的所有“非字母、非数字、非空格”的符号，只保留字母、数字和空白字符
func FilterSymbolsFast(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
