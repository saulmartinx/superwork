package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

func populateUser(user User) error {
	if err := populateWorkflow(user); err != nil {
		return err
	}

	if err := populateActivityTypes(user); err != nil {
		return err
	}

	if err := populateCurrencies(user); err != nil {
		return err
	}

	return nil
}

func assignOrganization(input ModelWithOrg, user User) error {
	if input.GetOrgName() != "" && input.GetOrgID() == "" {
		org := Organization{}
		org.Name = input.GetOrgName()
		org.CompanyID = user.ActiveCompanyID
		org.OwnerID = user.ID
		if err := insertOrganization(&org); err != nil {
			return err
		}
		input.SetOrgID(org.ID)

		timeline := Timeline{
			UnderCompanyID: user.ActiveCompanyID,
			UserID:         user.ID,
			OrganizationID: org.ID,
			Action:         "created",
		}
		timeline.Name = org.Name
		if err := insertTimeline(&timeline); err != nil {
			return err
		}
	}

	return nil
}

func assignPerson(input ModelWithPerson, user User) error {
	if input.GetPersonName() != "" && input.GetPersonID() == "" {
		person := Person{}
		person.Name = input.GetPersonName()
		person.CompanyID = user.ActiveCompanyID
		person.OwnerID = user.ID
		if err := insertPerson(&person); err != nil {
			return err
		}
		input.SetPersonID(person.ID)

		timeline := Timeline{
			UnderCompanyID: user.ActiveCompanyID,
			UserID:         user.ID,
			PersonID:       person.ID,
			Action:         "created",
		}
		timeline.Name = person.Name
		if err := insertTimeline(&timeline); err != nil {
			return err
		}
	}

	return nil
}

func populateActivityTypes(user User) error {
	if has, err := hasActivityTypes(user.ID); err != nil {
		return err
	} else if has {
		return nil
	}

	at := ActivityType{}
	at.Name = "Call"
	at.KeyString = "call"
	at.OrderNr = 0
	at.CompanyID = user.ActiveCompanyID
	at.IsCustomFlag = false
	if err := insertActivityType(&at); err != nil {
		return err
	}

	at = ActivityType{}
	at.Name = "Meeting"
	at.KeyString = "meeting"
	at.OrderNr = 1
	at.CompanyID = user.ActiveCompanyID
	at.IsCustomFlag = false
	if err := insertActivityType(&at); err != nil {
		return err
	}

	at = ActivityType{}
	at.Name = "Task"
	at.KeyString = "task"
	at.OrderNr = 2
	at.CompanyID = user.ActiveCompanyID
	at.IsCustomFlag = false
	if err := insertActivityType(&at); err != nil {
		return err
	}

	at = ActivityType{}
	at.Name = "Deadline"
	at.KeyString = "deadline"
	at.OrderNr = 3
	at.CompanyID = user.ActiveCompanyID
	at.IsCustomFlag = false
	if err := insertActivityType(&at); err != nil {
		return err
	}

	at = ActivityType{}
	at.Name = "Email"
	at.KeyString = "email"
	at.OrderNr = 4
	at.CompanyID = user.ActiveCompanyID
	at.IsCustomFlag = false
	if err := insertActivityType(&at); err != nil {
		return err
	}

	at = ActivityType{}
	at.Name = "Lunch"
	at.KeyString = "lunch"
	at.OrderNr = 5
	at.CompanyID = user.ActiveCompanyID
	at.IsCustomFlag = false
	if err := insertActivityType(&at); err != nil {
		return err
	}

	return nil
}

func populateWorkflow(user User) error {
	if has, err := hasWorkflows(user.ID); err != nil {
		return err
	} else if has {
		return nil
	}

	workflow := Workflow{}
	workflow.Name = "My workflow"
	workflow.CompanyID = user.ActiveCompanyID
	if err := insertWorkflow(&workflow); err != nil {
		return err
	}
	user.ActiveWorkflowID = workflow.ID

	if err := updateUser(user); err != nil {
		return err
	}

	// Workflow stages

	stage := Stage{}
	stage.Name = "Planned"
	stage.WorkflowID = workflow.ID
	stage.OrderNr = 0
	if err := insertStage(&stage); err != nil {
		return err
	}

	stage = Stage{}
	stage.Name = "In progress"
	stage.WorkflowID = workflow.ID
	stage.OrderNr = 1
	if err := insertStage(&stage); err != nil {
		return err
	}

	stage = Stage{}
	stage.Name = "Done"
	stage.WorkflowID = workflow.ID
	stage.OrderNr = 2
	if err := insertStage(&stage); err != nil {
		return err
	}

	return nil
}

func populateCurrencies(user User) error {
	if has, err := hasCurrencies(user.ID); err != nil {
		return err
	} else if has {
		return nil
	}

	b, err := ioutil.ReadFile(filepath.Join(config.Dir, "db", "currencies.json"))
	if err != nil {
		return err
	}

	var currencies []map[string]interface{}
	if err := json.Unmarshal(b, &currencies); err != nil {
		return err
	}

	for _, c := range currencies {
		currency := Currency{}
		currency.CompanyID = user.ActiveCompanyID
		currency.Name = c["name"].(string)
		currency.Code = c["code"].(string)
		currency.DecimalPoints = int(c["decimal_points"].(float64))
		currency.Symbol = c["symbol"].(string)
		currency.IsCustomFlag = false
		if err := insertCurrency(&currency); err != nil {
			return err
		}
	}

	return nil
}
