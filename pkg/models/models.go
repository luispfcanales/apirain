package models

type MaintenanceTeam struct {
	ID   int    `json:"id" xmlrpc:"id"`
	Name string `json:"name" xmlrpc:"name"`
}

type MaintenanceEquipment struct {
	ID                int    `json:"id" xmlrpc:"id"`
	Name              string `json:"name" xmlrpc:"name"`
	MaintenanceTeamID any    `json:"maintenance_team_id" xmlrpc:"maintenance_team_id"`
	CategoryID        any    `json:"category_id" xmlrpc:"category_id"`
	Location          string `json:"location" xmlrpc:"location"`
}

type MaintenanceRequest struct {
	ID                any `json:"id" xmlrpc:"id"`
	Name              any `json:"name" xmlrpc:"name"`
	MaintenanceTeamID any `json:"maintenance_team_id" xmlrpc:"maintenance_team_id"`
	StageID           any `json:"stage_id" xmlrpc:"stage_id"`
	Priority          any `json:"priority" xmlrpc:"priority"`
	ScheduleDate      any `json:"schedule_date" xmlrpc:"schedule_date"`
	EquipmentID       any `json:"equipment_id" xmlrpc:"equipment_id"`
	CorrectiveDate    any `json:"corrective_date" xmlrpc:"corrective_date"`
	RepeatInterval    any `json:"repeat_interval" xmlrpc:"repeat_interval"`
	RecurrenceType    any `json:"recurrence_type" xmlrpc:"recurrence_type"`
	RecurrenceValue   any `json:"recurrence_value" xmlrpc:"recurrence_value"`
	RepeatType        any `json:"repeat_type" xmlrpc:"repeat_type"`
	RepeatUnit        any `json:"repeat_unit" xmlrpc:"repeat_unit"`
	Archive           any `json:"archive" xmlrpc:"archive"`
	RequestDate       any `json:"request_date" xmlrpc:"request_date"`
	PreventiveDate    any `json:"preventive_date" xmlrpc:"preventive_date"`
	UsedValue         any `json:"used_value" xmlrpc:"used_value"`
}

func (mr *MaintenanceRequest) GetTeamID() int {
	if mr.MaintenanceTeamID == nil {
		return 0
	}

	// Odoo returns [id, name] or false
	if teamData, ok := mr.MaintenanceTeamID.([]interface{}); ok && len(teamData) > 0 {
		switch v := teamData[0].(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func (mr *MaintenanceRequest) GetTeamName() string {
	if mr.MaintenanceTeamID == nil {
		return ""
	}

	if teamData, ok := mr.MaintenanceTeamID.([]interface{}); ok && len(teamData) > 1 {
		if teamName, ok := teamData[1].(string); ok {
			return teamName
		}
	}
	return ""
}
