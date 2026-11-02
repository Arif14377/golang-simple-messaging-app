package repository

import (
	"context"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"go.elastic.co/apm/v2"
)

// startMySQLSpan membuat span "db" untuk query MySQL/GORM.
func startMySQLSpan(ctx context.Context, name, action string) (context.Context, *apm.Span) {
	instance := env.GetEnv("DB_NAME", "simple_messaging_app")
	spanCtx, span := tracing.StartSpan(ctx, name, "db")
	tracing.SetDatabase(span, "mysql", instance, action)

	return spanCtx, span
}

// startMongoSpan membuat span "db" untuk operasi MongoDB.
func startMongoSpan(ctx context.Context, name, action string) (context.Context, *apm.Span) {
	instance := database.MessageCollectionName()
	spanCtx, span := tracing.StartSpan(ctx, name, "db")
	tracing.SetDatabase(span, "mongodb", instance, action)

	return spanCtx, span
}
