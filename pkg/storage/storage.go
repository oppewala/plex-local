package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	//"github.com/oppewala/plex-local-dl/pkg/plex"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	//"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

type Storage struct {
	client *aztables.Client
}

type Entry struct {
	Category string
	DBId     string
	Title    string
	PlexKey  uint
}

type DuplicateEntryError struct {
	partitionKey string
	rowKey       string
	entry        Entry
}

func (e *DuplicateEntryError) Error() string {
	return fmt.Sprintf("Entry with partition key '%v' and row key '%v' already exists: %v", e.partitionKey, e.rowKey, e.entry)
}

func ConnectStorage(connectionString string) *Storage {
	sc, err := aztables.NewServiceClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Fatalf("Failed to connect to Azure Table: %v", err)
	}

	c := sc.NewClient("watch")

	log.Printf("[Storage] Azure client connected")
	return &Storage{
		client: c,
	}
}

func (s *Storage) Add(entry Entry) error {
	exists, err := s.Exists(entry.Category, entry.DBId)
	if err != nil {
		return err
	}

	if exists == true {
		return &DuplicateEntryError{
			partitionKey: entry.Category,
			rowKey:       entry.DBId,
			entry:        entry,
		}
	}

	e := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: entry.Category,
			RowKey:       entry.DBId,
		},
		Properties: map[string]interface{}{
			"Title":   entry.Title,
			"PlexKey": entry.PlexKey,
		},
	}
	j, err := json.Marshal(e)
	if err != nil {
		err = fmt.Errorf("failed to marshal entity: %w", err)
		return err
	}

	_, err = s.client.AddEntity(context.TODO(), j, nil)

	return err
}

func (s *Storage) Exists(category string, dbid string) (bool, error) {
	_, err := s.client.GetEntity(context.TODO(), category, dbid, nil)
	if err == nil {
		log.Printf("[Storage] Entry already exists in azure table for partition key '%v' and row key '%v'", category, dbid)
		return true, nil
	}

	formattedErr := fmt.Errorf("failed to get entity with partition key '%v' and row key '%v' from azure table: %w", category, dbid, err)
	var dat map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(err.Error()), &dat); err != nil {
		err = fmt.Errorf("could not handle error response from azure table: %v - %w", formattedErr, err)
		return false, err
	}

	if dat["odata.error"]["code"] == "ResourceNotFound" {
		return false, nil
	}
	return false, formattedErr
}

func (s *Storage) Remove(category string, dbid string) error {
	exists, err := s.Exists(category, dbid)
	if err != nil {
		return err
	}

	if exists == false {
		log.Printf("[Storage] No entry found to remove with partition key '%v' and row key '%v' from azure table", category, dbid)
		return nil
	}

	_, err = s.client.DeleteEntity(context.TODO(), category, dbid, nil)
	return err
}

// ForceRemove is raw input to delete entities, should only be used to clean up bad data
func (s *Storage) ForceRemove(partition string, row string) error {
	_, err := s.client.DeleteEntity(context.TODO(), partition, row, nil)
	return err
}

func (s *Storage) Get(category string, dbid string) (Entry, error) {
	e, err := s.client.GetEntity(context.TODO(), category, dbid, nil)
	if err != nil {
		err = fmt.Errorf("failed to get entity: %w", err)
		return Entry{}, err
	}

	var entity aztables.EDMEntity
	err = json.Unmarshal(e.Value, &entity)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal entity: %w", err)
		return Entry{}, err
	}

	return Entry{
		Category: entity.PartitionKey,
		DBId:     entity.RowKey,
		Title:    entity.Properties["Title"].(string),
		PlexKey:  entity.Properties["PlexKey"].(uint),
	}, nil
}

func (s *Storage) List() ([]Entry, error) {
	//filter := fmt.Sprintf("PartitionKey eq 'movie' and RowKey eq '1234'")
	//opt := &aztables.ListEntitiesOptions{
	//	Filter: &filter,
	//	Select: to.StringPtr("RowKey,Value,Product,Available"),
	//	Top: to.Int32Ptr(20),
	//}

	entries := make([]Entry, 0)

	pager := s.client.List(nil)
	for pager.NextPage(context.TODO()) {
		resp := pager.PageResponse()
		fmt.Printf("Received: %v entities\n", len(resp.Entities))

		for _, e := range resp.Entities {
			var entity aztables.EDMEntity
			err := json.Unmarshal(e, &entity)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal entity: %w", err)
				return nil, err
			}

			entries = append(entries, Entry{
				Category: entity.PartitionKey,
				DBId:     entity.RowKey,
				Title:    entity.Properties["Title"].(string),
				PlexKey:  entity.Properties["PlexKey"].(uint),
			})
		}
	}

	return entries, nil
}
