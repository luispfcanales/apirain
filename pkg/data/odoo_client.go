package data

import (
	"fmt"

	"github.com/luispfcanales/apirain/pkg/config"
	"github.com/luispfcanales/apirain/pkg/models"

	"github.com/kolo/xmlrpc"
)

type OdooClient struct {
	cfg      *config.Config
	uid      int
	password string
}

func NewOdooClient(cfg *config.Config) (*OdooClient, error) {
	commonClient, err := xmlrpc.NewClient(cfg.OdooURL+"/xmlrpc/2/common", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating common client: %w", err)
	}
	defer commonClient.Close()

	var uid int
	err = commonClient.Call("authenticate", []interface{}{
		cfg.OdooDB,
		cfg.OdooUsername,
		cfg.OdooPassword,
		map[string]interface{}{},
	}, &uid)
	if err != nil {
		return nil, fmt.Errorf("error in authentication: %w", err)
	}

	if uid == 0 {
		return nil, fmt.Errorf("authentication failed: verify credentials")
	}

	return &OdooClient{
		cfg:      cfg,
		uid:      uid,
		password: cfg.OdooPassword,
	}, nil
}

func (c *OdooClient) call(model, method string, args []interface{}, kwargs map[string]interface{}, reply interface{}) error {
	modelsClient, err := xmlrpc.NewClient(c.cfg.OdooURL+"/xmlrpc/2/object", nil)
	if err != nil {
		return fmt.Errorf("error creating models client: %w", err)
	}
	defer modelsClient.Close()

	return modelsClient.Call("execute_kw", []interface{}{
		c.cfg.OdooDB,
		c.uid,
		c.password,
		model,
		method,
		args,
		kwargs,
	}, reply)
}

func (c *OdooClient) GetMaintenanceTeams() ([]models.MaintenanceTeam, error) {
	var teams []models.MaintenanceTeam
	err := c.call("maintenance.team", "search_read", []interface{}{[]interface{}{}}, map[string]interface{}{
		"fields": []string{"name", "id"},
	}, &teams)
	return teams, err
}

func (c *OdooClient) GetMaintenanceRequests() ([]models.MaintenanceRequest, error) {
	var requests []models.MaintenanceRequest
	err := c.call("maintenance.request", "search_read", []interface{}{[]interface{}{}}, map[string]interface{}{
		"fields": []string{
			"name",
			"maintenance_team_id",
			"stage_id",
			"priority",
			"schedule_date",
			"equipment_id",
			"corrective_date",
			"repeat_interval",
			"recurrence_type",
			"recurrence_value",
			"repeat_type",
			"repeat_unit",
		},
		"limit": 100,
	}, &requests)
	return requests, err
}

func (c *OdooClient) GetMaintenanceEquipment() ([]models.MaintenanceEquipment, error) {
	var equipment []models.MaintenanceEquipment
	err := c.call("maintenance.equipment", "search_read", []interface{}{[]interface{}{}}, map[string]interface{}{
		"fields": []string{"name", "maintenance_team_id", "category_id", "location"},
		"limit":  100,
	}, &equipment)
	return equipment, err
}
