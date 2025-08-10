package testutils

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"palu-wiki/internal/models"
)

// TestDB 测试数据库助手
type TestDB struct {
	DB   *gorm.DB
	Mock sqlmock.Sqlmock
}

// SetupTestDB 创建测试用的数据库连接和mock
func SetupTestDB(t *testing.T) *TestDB {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 禁用日志输出
	})
	require.NoError(t, err)

	return &TestDB{
		DB:   gormDB,
		Mock: mock,
	}
}

// Close 关闭测试数据库连接
func (tdb *TestDB) Close() {
	if sqlDB, err := tdb.DB.DB(); err == nil {
		sqlDB.Close()
	}
}

// ExpectationsWereMet 验证所有mock期望都已满足
func (tdb *TestDB) ExpectationsWereMet(t *testing.T) {
	require.NoError(t, tdb.Mock.ExpectationsWereMet())
}

// MockSelectUser 模拟查询用户的SQL期望
func (tdb *TestDB) MockSelectUser(user *models.User) {
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "nickname", "role", "status",
		"created_at", "updated_at",
	}).AddRow(
		user.ID, user.Username, user.Email, user.Password, user.Nickname,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)

	tdb.Mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(rows)
}

// MockCreateUser 模拟创建用户的SQL期望
func (tdb *TestDB) MockCreateUser(user *models.User) {
	tdb.Mock.ExpectBegin()
	tdb.Mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(
			user.Username, user.Email, user.Password, user.Nickname,
			user.Role, user.Status, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(user.ID))
	tdb.Mock.ExpectCommit()
}

// MockSelectArticle 模拟查询文章的SQL期望
func (tdb *TestDB) MockSelectArticle(article *models.Article) {
	rows := sqlmock.NewRows([]string{
		"id", "title", "content", "summary", "author_id", "category_id",
		"status", "view_count", "created_at", "updated_at",
	}).AddRow(
		article.ID, article.Title, article.Content, article.Summary,
		article.AuthorID, article.CategoryID, article.Status,
		article.ViewCount, article.CreatedAt, article.UpdatedAt,
	)

	tdb.Mock.ExpectQuery(`SELECT \* FROM "articles"`).
		WillReturnRows(rows)
}

// MockCreateArticle 模拟创建文章的SQL期望
func (tdb *TestDB) MockCreateArticle(article *models.Article) {
	tdb.Mock.ExpectBegin()
	tdb.Mock.ExpectQuery(`INSERT INTO "articles"`).
		WithArgs(
			article.Title, article.Content, article.Summary,
			article.AuthorID, article.CategoryID, article.Status,
			article.ViewCount, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(article.ID))
	tdb.Mock.ExpectCommit()
}

// MockUpdateArticle 模拟更新文章的SQL期望
func (tdb *TestDB) MockUpdateArticle(article *models.Article) {
	tdb.Mock.ExpectBegin()
	tdb.Mock.ExpectExec(`UPDATE "articles" SET`).
		WithArgs(
			article.Title, article.Content, article.Summary,
			article.CategoryID, article.Status, sqlmock.AnyArg(),
			article.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	tdb.Mock.ExpectCommit()
}

// MockDeleteArticle 模拟删除文章的SQL期望
func (tdb *TestDB) MockDeleteArticle(id uint) {
	tdb.Mock.ExpectBegin()
	tdb.Mock.ExpectExec(`DELETE FROM "articles"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	tdb.Mock.ExpectCommit()
}

// AnyTime 用于匹配任意时间的helper
type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}