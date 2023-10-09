package dynamodb_test

import (
	"testing"

	"github.com/ricomonster/black-flag/internal/aws/dynamodb"
	"github.com/ricomonster/black-flag/internal/config"
	"github.com/ricomonster/black-flag/internal/videos"
)

func init() {
	config.LoadEnvConfig()
}

func TestListTables(t *testing.T) {
	svc := dynamodb.NewDynamoDB[interface{}]("")
	tables, err := svc.ListTables()
	if err != nil {
		t.Errorf("TestListTables: Error %v", err)
	}

	t.Logf("TestListTables: Tables %v", tables)
}

func TestFindById(t *testing.T) {
	svc := dynamodb.NewDynamoDB[interface{}]("BlackFlag_Videos")
	item, err := svc.FindById("Id", "_Hu4GYtye5U")
	if err != nil {
		t.Errorf("TestFindById: Error %v", err)
	}

	t.Logf("TestFindById: Tables %v", item)
}

func TestUpdateItem(t *testing.T) {
	svc := dynamodb.NewDynamoDB[videos.VideoDdbAttributes]("BlackFlag_Videos")
	itemToUpdate := videos.VideoDdbAttributes{
		TestAttribute: "Made in Japan",
		// TestAttributeHana: 23,
	}

	err := svc.UpdateItem("Id", "_Hu4GYtye5U", itemToUpdate)
	if err != nil {
		t.Errorf("TestUpdateItem: Error %v", err)
	}
}
