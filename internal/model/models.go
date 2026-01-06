package model

// 所有模型的统一导入点
// 用于 AutoMigrate
var AllModels = []interface{}{
	&ChatSession{},
	&ChatMessage{},
	&Agent{},
	&KnowledgeBase{},
	&Document{},
	&DocumentChunk{},
	&IndexMapping{},
	&Tool{},
	&FAQ{},
	&KnowledgeTag{},
	&User{},
	&AuthToken{},
}
