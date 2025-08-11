package main

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const testDB = "superwork_test"

func TestMain(m *testing.M) {
	if err := envconfig.Process("superwork", &config); err != nil {
		log.Fatal(err)
	}
	config.Env = "test"
	if err := recreateDB(testDB); err != nil {
		log.Fatal(err)
	}
	if err := connectDB(testDB); err != nil {
		log.Fatal(err)
	}
	retCode := m.Run()
	os.Exit(retCode)
}

func TestUser(t *testing.T) {
	user := User{
		Email: "someone@somewhere.com",
	}

	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}
	user.ActiveCompanyID = company.ID

	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	companyUser := CompanyUser{
		CompanyID: company.ID,
		UserID:    user.ID,
	}
	if err := insertCompanyUser(&companyUser); err != nil {
		t.Fatal(err)
	}

	ID, err := selectFirstCompanyByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ID != company.ID {
		t.Fatal("invalid company")
	}

	user.Name = "blah"
	if err := updateUser(user); err != nil {
		t.Fatal(err)
	}

	user.PasswordHash = "voodoo"
	if err := updateUser(user); err != nil {
		t.Fatal(err)
	}

	model, err := selectUserByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model.ID != user.ID {
		t.Fatal("Invalid user ID")
	}

	model, err = selectUserByEmail(user.Email)
	if err != nil {
		t.Fatal(err)
	}
	if model.ID != user.ID {
		t.Fatal("Invalid user ID")
	}

	if err := deleteUser(model.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCompany(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone1@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	companyUser := CompanyUser{
		CompanyID: company.ID,
		UserID:    user.ID,
	}
	if err := insertCompanyUser(&companyUser); err != nil {
		t.Fatal(err)
	}

	belongs, err := companyBelongsToUser(company.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !belongs {
		t.Fatal("company should belong to user")
	}

	companyUsers, err := selectCompanyUsersByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(companyUsers) != 1 {
		t.Fatal("company user not found")
	}
	if companyUsers[0].UserID != user.ID {
		t.Fatal("invalid company user")
	}

	model, err := selectCompanyUserByUserAndCompany(user.ID, company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model == nil {
		t.Fatal("company user not found")
	}
	if model.ID != companyUser.ID {
		t.Fatal("invalid company user ID")
	}

	if err := updateCompany(company); err != nil {
		t.Fatal(err)
	}

	companyUsers, err = selectCompanyUsersByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(companyUsers) != 1 {
		t.Fatal("company user not found")
	}
	cu := companyUsers[0]
	if cu.CompanyID != company.ID {
		t.Fatal("invalid company ID")
	}

	cu2, err := selectCompanyUserByID(companyUsers[0].ID)
	if err != nil {
		t.Fatal("company user not found")
	}
	if cu2.UserID != user.ID {
		t.Fatal("invalid company user")
	}

	c, err := selectCompanyByID(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("company not found")
	}
	if c.ID != company.ID {
		t.Fatal("invalid company found")
	}

	cu.IsAdmin = true
	if err := updateCompanyUser(cu); err != nil {
		t.Fatal(err)
	}

	if err := deleteCompany(company.ID); err != nil {
		t.Fatal(err)
	}

	if err := deleteCompanyUser(cu.ID); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflow(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone2@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	workflow := Workflow{}
	workflow.Name = "voodoo"
	workflow.CompanyID = company.ID
	if err := insertWorkflow(&workflow); err != nil {
		t.Fatal(err)
	}

	workflows, err := selectWorkflowsByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(workflows) == 0 {
		t.Fatal("Workflows not found")
	}

	workflow.Name = "chill"
	if err := updateWorkflow(workflow); err != nil {
		t.Fatal(err)
	}

	model, err := selectWorkflowByID(workflow.ID)
	if err != nil {
		t.Fatal(err)
	}

	ID, _, err := selectFirstWorkflowByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ID != workflow.ID {
		t.Fatal("invalid workflow")
	}

	if err := deleteWorkflow(model.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPerson(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone3@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	person := Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	p, err := selectPersonByID(person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != person.ID {
		t.Fatal("person not found")
	}

	// Select persons by company

	persons, err := selectPersonsByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) == 0 {
		t.Fatal("persons not found")
	}

	// Select persons by organization

	org := Organization{}
	org.CompanyID = company.ID
	org.OwnerID = user.ID
	org.Name = "Shitfuck Inc"
	if err := insertOrganization(&org); err != nil {
		t.Fatal(err)
	}

	person = Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	person.OrgID = org.ID
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	persons, err = selectPersonsByOrganization(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) == 0 {
		t.Fatal("persons not found")
	}

	// update person

	person.FirstName = "Andres"
	person.Name = "Lainela"
	if err := updatePerson(person); err != nil {
		t.Fatal(err)
	}

	if err := deletePerson(person.ID); err != nil {
		t.Fatal(err)
	}
}

func TestOrganization(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone4@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	org := Organization{}
	org.CompanyID = company.ID
	org.OwnerID = user.ID
	org.Name = "Shitfuck Inc"
	if err := insertOrganization(&org); err != nil {
		t.Fatal(err)
	}

	organizations, err := selectOrganizationsByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(organizations) == 0 {
		t.Fatal("organizations not found")
	}

	model, err := selectOrganizationByID(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model.ID != org.ID {
		t.Fatal("organization not found")
	}

	org.Name = "italian mafia"
	if err := updateOrganization(org); err != nil {
		t.Fatal(err)
	}

	if err := deleteOrganization(org.ID); err != nil {
		t.Fatal(err)
	}
}

func TestActivation(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone111@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	activation := Activation{
		UserID: user.ID,
	}
	if err := insertActivation(&activation); err != nil {
		t.Fatal(err)
	}

	model, err := selectActivationByID(activation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model.ID != activation.ID {
		t.Fatal("invalid activation")
	}

	if err := deleteActivation(activation.ID); err != nil {
		t.Fatal(err)
	}
}

func TestActivity(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone5@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	org := Organization{}
	org.CompanyID = company.ID
	org.OwnerID = user.ID
	org.Name = "Shitfuck Inc"
	if err := insertOrganization(&org); err != nil {
		t.Fatal(err)
	}

	act := Activity{}
	act.Name = "Work"
	act.UserID = user.ID
	act.OrgID = org.ID
	act.CompanyID = company.ID
	if err := insertActivity(&act); err != nil {
		t.Fatal(err)
	}

	activities, err := selectActivitiesByOrganization(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(activities) != 1 {
		t.Fatal("activities not found")
	}

	a, err := selectActivityByID(act.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != act.ID {
		t.Fatal("activity not found")
	}

	activities, err = selectActivitiesByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(activities) == 0 {
		t.Fatal("activities not found")
	}

	workflow := Workflow{}
	workflow.Name = "voodoo"
	workflow.CompanyID = company.ID
	if err := insertWorkflow(&workflow); err != nil {
		t.Fatal(err)
	}

	stage := Stage{}
	stage.WorkflowID = workflow.ID
	stage.Name = "first stage"
	if err := insertStage(&stage); err != nil {
		t.Fatal(err)
	}

	task := Task{}
	task.CreatorUserID = user.ID
	task.UserID = user.ID
	task.WorkflowID = workflow.ID
	task.StageID = stage.ID
	task.Name = "blah"
	task.CompanyID = company.ID
	if err := insertTask(&task); err != nil {
		t.Fatal(err)
	}

	act.TaskID = task.ID
	if err := updateActivity(act); err != nil {
		t.Fatal(err)
	}

	activities, err = selectActivitiesByTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(activities) != 1 {
		t.Fatal("activity not found by task")
	}

	person := Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	act.PersonID = person.ID
	if err := updateActivity(act); err != nil {
		t.Fatal(err)
	}

	activities, err = selectActivitiesByPerson(person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(activities) != 1 {
		t.Fatal("activities not found")
	}

	act.Name = "updated name"
	if err := updateActivity(act); err != nil {
		t.Fatal(err)
	}

	if err := deleteActivity(act.ID); err != nil {
		t.Fatal(err)
	}
}

func TestNote(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone6@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	person := Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	note := Note{}
	note.Name = "Work"
	note.UserID = user.ID
	note.CompanyID = company.ID
	note.PersonID = person.ID
	if err := insertNote(&note); err != nil {
		t.Fatal(err)
	}

	notes, err := selectNotesByPerson(person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 {
		t.Fatal("notes not found")
	}

	// Select notes by task

	workflow := Workflow{}
	workflow.Name = "voodoo"
	workflow.CompanyID = company.ID
	if err := insertWorkflow(&workflow); err != nil {
		t.Fatal(err)
	}

	stage := Stage{}
	stage.WorkflowID = workflow.ID
	stage.Name = "first stage"
	if err := insertStage(&stage); err != nil {
		t.Fatal(err)
	}

	model, err := selectStageByID(stage.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model.ID != stage.ID {
		t.Fatal("invalid stage loaded")
	}

	task := Task{}
	task.CreatorUserID = user.ID
	task.UserID = user.ID
	task.WorkflowID = workflow.ID
	task.StageID = stage.ID
	task.Name = "blah"
	task.CompanyID = company.ID
	if err := insertTask(&task); err != nil {
		t.Fatal(err)
	}

	note = Note{}
	note.Name = "Work"
	note.UserID = user.ID
	note.CompanyID = company.ID
	note.TaskID = task.ID
	if err := insertNote(&note); err != nil {
		t.Fatal(err)
	}

	notes, err = selectNotesByTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 {
		t.Fatal("notes not found")
	}

	// update note

	note.Name = "updated note name"
	if err := updateNote(note); err != nil {
		t.Fatal(err)
	}

	n, err := selectNoteByID(note.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n == nil {
		t.Fatal("note not found")
	}
	if note.ID != n.ID {
		t.Fatal("invalid note found")
	}

	org := Organization{}
	org.CompanyID = company.ID
	org.OwnerID = user.ID
	org.Name = "Shitfuck Inc"
	if err := insertOrganization(&org); err != nil {
		t.Fatal(err)
	}

	note.OrgID = org.ID
	if err := updateNote(note); err != nil {
		t.Fatal(err)
	}

	notes, err = selectNotesByOrganization(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 {
		t.Fatal("note not found by org")
	}

	if err := deleteNote(model.ID); err != nil {
		t.Fatal(err)
	}
}

func TestContact(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone7@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	person := Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	model := Contact{}
	model.Name = "Work"
	model.PersonID = person.ID
	model.Type = "phone"
	if err := insertContact(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated contact"
	if err := updateContact(model); err != nil {
		t.Fatal(err)
	}

	contact, err := selectContactByID(model.ID)
	if err != nil {
		t.Fatal(err)
	}
	if contact.ID != model.ID {
		t.Fatal("contact not found")
	}

	contacts, err := selectContactsByPerson(person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(contacts) != 1 {
		t.Fatal("contacts not found")
	}

	if err := deleteContact(model.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPersonField(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone8@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := PersonField{}
	model.Name = "perse"
	model.CompanyID = company.ID
	if err := insertPersonField(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "perseke"
	if err := updatePersonField(model); err != nil {
		t.Fatal(err)
	}

	if err := deletePersonField(model); err != nil {
		t.Fatal(err)
	}
}

func TestCategory(t *testing.T) {
	model := Category{}
	model.Name = "something"
	if err := insertCategory(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated name"
	if err := updateCategory(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteCategory(model); err != nil {
		t.Fatal(err)
	}
}

func TestFile(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone9@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := File{}
	model.UserID = user.ID
	if err := insertFile(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated name"
	if err := updateFile(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteFile(model); err != nil {
		t.Fatal(err)
	}
}

func TestStage(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone10@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	workflow := Workflow{}
	workflow.Name = "voodoo"
	workflow.CompanyID = company.ID
	if err := insertWorkflow(&workflow); err != nil {
		t.Fatal(err)
	}

	model := Stage{}
	model.WorkflowID = workflow.ID
	model.Name = "first stage"
	if err := insertStage(&model); err != nil {
		t.Fatal(err)
	}

	// select stages by workflow

	stages, err := selectStagesByWorkflow(workflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) == 0 {
		t.Fatal("stages not found")
	}

	// update stage

	model.Name = "second stage"
	if err := updateStage(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteStage(model.ID); err != nil {
		t.Fatal(err)
	}
}

func TestTask(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone11@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	workflow := Workflow{}
	workflow.Name = "voodoo"
	workflow.CompanyID = company.ID
	if err := insertWorkflow(&workflow); err != nil {
		t.Fatal(err)
	}

	stage := Stage{}
	stage.WorkflowID = workflow.ID
	stage.Name = "first stage"
	if err := insertStage(&stage); err != nil {
		t.Fatal(err)
	}

	model := Task{}
	model.CreatorUserID = user.ID
	model.UserID = user.ID
	model.WorkflowID = workflow.ID
	model.StageID = stage.ID
	model.Name = "bad task"
	model.CompanyID = company.ID
	if err := insertTask(&model); err != nil {
		t.Fatal(err)
	}

	tasks, err := selectTasksByWorkflow(workflow.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) == 0 {
		t.Fatal("tasks not found")
	}

	now := time.Now()
	model.WonTime = &now
	if err := updateTask(model); err != nil {
		t.Fatal(err)
	}

	tasks, err = selectTasksByWorkflow(workflow.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Fatal("tasks found")
	}

	tasks, err = selectTasksByWorkflow(workflow.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) == 0 {
		t.Fatal("tasks not found")
	}

	org := Organization{}
	org.CompanyID = company.ID
	org.OwnerID = user.ID
	org.Name = "Shitfuck Inc"
	if err := insertOrganization(&org); err != nil {
		t.Fatal(err)
	}

	model = Task{}
	model.CreatorUserID = user.ID
	model.UserID = user.ID
	model.WorkflowID = workflow.ID
	model.StageID = stage.ID
	model.OrgID = org.ID
	model.Name = "good task, the best actually"
	model.CompanyID = company.ID
	if err := insertTask(&model); err != nil {
		t.Fatal(err)
	}

	tasks, err = selectTasksByOrganization(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatal("tasks not found")
	}

	// Select tasks by person

	person := Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	model = Task{}
	model.CreatorUserID = user.ID
	model.UserID = user.ID
	model.WorkflowID = workflow.ID
	model.StageID = stage.ID
	model.PersonID = person.ID
	model.Name = "another great deal"
	model.CompanyID = company.ID
	if err := insertTask(&model); err != nil {
		t.Fatal(err)
	}

	models, err := selectTasksByPerson(person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 {
		t.Fatal("task not found by person")
	}

	// update

	model.Name = "Updated name"
	if err := updateTask(model); err != nil {
		t.Fatal(err)
	}

	task, err := selectTaskByID(model.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := deleteTask(task.ID); err != nil {
		t.Fatal(err)
	}
}

func TestOrganizationField(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone12@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	org := Organization{}
	org.CompanyID = company.ID
	org.OwnerID = user.ID
	org.Name = "Shitfuck Inc"
	if err := insertOrganization(&org); err != nil {
		t.Fatal(err)
	}

	model := OrganizationField{}
	model.CompanyID = company.ID
	model.Name = "bah"
	if err := insertOrganizationField(&model); err != nil {
		t.Fatal(err)
	}

	// update organization field

	model.Name = "updated bah"
	if err := updateOrganizationField(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteOrganizationField(model); err != nil {
		t.Fatal(err)
	}
}

func TestDealField(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone13@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := TaskField{}
	model.CompanyID = company.ID
	if err := insertTaskField(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated deal field"
	if err := updateTaskField(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteTaskField(model); err != nil {
		t.Fatal(err)
	}
}

func TestActivityField(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone14@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := ActivityField{}
	model.CompanyID = company.ID
	if err := insertActivityField(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated field"
	if err := updateActivityField(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteActivityField(model); err != nil {
		t.Fatal(err)
	}
}

func TestActivityType(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone15@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := ActivityType{}
	model.CompanyID = company.ID
	if err := insertActivityType(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated field"
	if err := updateActivityType(model); err != nil {
		t.Fatal(err)
	}

	at, err := selectActivityTypeByID(model.ID)
	if err != nil {
		t.Fatal(err)
	}
	if at.ID != model.ID {
		t.Fatal("invalid activity type")
	}

	types, err := selectActivityTypesByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 1 {
		t.Fatal("types not found")
	}

	if err := deleteActivityType(model); err != nil {
		t.Fatal(err)
	}
}

func TestCurrency(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone16@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := Currency{}
	model.CompanyID = company.ID
	if err := insertCurrency(&model); err != nil {
		t.Fatal(err)
	}

	currencies, err := selectCurrenciesByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(currencies) != 1 {
		t.Fatal("currencies not found")
	}

	model.Name = "updated field"
	if err := updateCurrency(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteCurrency(model); err != nil {
		t.Fatal(err)
	}
}

func TestFilter(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone17@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := Filter{}
	model.CompanyID = company.ID
	model.UserID = user.ID
	if err := insertFilter(&model); err != nil {
		t.Fatal(err)
	}

	model.Name = "updated field"
	if err := updateFilter(model); err != nil {
		t.Fatal(err)
	}

	if err := deleteFilter(model); err != nil {
		t.Fatal(err)
	}
}

func TestGoal(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone20@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	goal := Goal{}
	goal.CompanyID = company.ID
	goal.Name = "important goal"
	goal.UserID = user.ID
	if err := insertGoal(&goal); err != nil {
		t.Fatal(err)
	}

	goal.Name = "Updated name"
	if err := updateGoal(goal); err != nil {
		t.Fatal(err)
	}

	if err := deleteGoal(goal); err != nil {
		t.Fatal(err)
	}
}

func TestNoteField(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone21@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	nf := NoteField{}
	nf.CompanyID = company.ID
	nf.Name = "blah"
	if err := insertNoteField(&nf); err != nil {
		t.Fatal(err)
	}

	nf.Name = "updated"
	if err := updateNoteField(nf); err != nil {
		t.Fatal(err)
	}

	if err := deleteNoteField(nf); err != nil {
		t.Fatal(err)
	}
}

func TestOrganizationRelationship(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone22@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	or := OrganizationRelationship{}
	or.CompanyID = company.ID
	or.Name = "blah"
	if err := insertOrganizationRelationship(&or); err != nil {
		t.Fatal(err)
	}

	or.Name = "updated"
	if err := updateOrganizationRelationship(or); err != nil {
		t.Fatal(err)
	}

	if err := deleteOrganizationRelationship(or); err != nil {
		t.Fatal(err)
	}
}

func TestProduct(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone23@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	p := Product{}
	p.CompanyID = company.ID
	p.Name = "blah"
	p.OwnerID = user.ID
	if err := insertProduct(&p); err != nil {
		t.Fatal(err)
	}

	p.Name = "updated"
	if err := updateProduct(p); err != nil {
		t.Fatal(err)
	}

	// price

	price := Price{}
	price.ProductID = p.ID
	if err := insertPrice(&price); err != nil {
		t.Fatal(err)
	}

	price.Price = 123
	if err := updatePrice(price); err != nil {
		t.Fatal(err)
	}

	if err := deletePrice(price); err != nil {
		t.Fatal(err)
	}

	// delete product

	if err := deleteProduct(p); err != nil {
		t.Fatal(err)
	}
}

func TestProductField(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone24@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	p := ProductField{}
	p.CompanyID = company.ID
	p.Name = "blah"
	if err := insertProductField(&p); err != nil {
		t.Fatal(err)
	}

	if err := updateProductField(p); err != nil {
		t.Fatal(err)
	}

	if err := deleteProductField(p); err != nil {
		t.Fatal(err)
	}
}

func TestPushNotification(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone25@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	p := PushNotification{}
	p.CompanyID = company.ID
	p.Name = "blah"
	p.UserID = user.ID
	if err := insertPushNotification(&p); err != nil {
		t.Fatal(err)
	}

	if err := updatePushNotification(p); err != nil {
		t.Fatal(err)
	}

	if err := deletePushNotification(p); err != nil {
		t.Fatal(err)
	}
}

func TestTimeline(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone26@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	model := Timeline{
		UnderCompanyID: company.ID,
		UserID:         user.ID,
		Action:         "created",
	}
	model.Name = "something"
	if err := insertTimeline(&model); err != nil {
		t.Fatal(err)
	}

	models, err := selectTimelineByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) == 0 {
		t.Fatal("timeline not found")
	}
}

func TestDeletedObjects(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone50@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	timeEntry := TimeEntry{
		UserID:    user.ID,
		CompanyID: company.ID,
	}
	timeEntry.Name = "zaajets"
	if err := insertTimeEntry(&timeEntry); err != nil {
		t.Fatal(err)
	}

	if err := deleteTimeEntry(timeEntry.ID); err != nil {
		t.Fatal(err)
	}

	deleted, err := selectDeletedObjectsByCompany(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted) != 1 {
		t.Fatal("1 deleted object expected")
	}
	if deleted[0].Type != "time_entries" {
		t.Fatal("deleted time entry expected")
	}
	if deleted[0].ID != timeEntry.ID {
		t.Fatal("invalid time entry ID")
	}
	if deleted[0].Name != timeEntry.Name {
		t.Fatal("invalid time entry name")
	}
}

func TestTimeEntry(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone30@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	timeEntry := TimeEntry{
		UserID:    user.ID,
		CompanyID: company.ID,
	}
	if err := insertTimeEntry(&timeEntry); err != nil {
		t.Fatal(err)
	}

	model, err := selectTimeEntryByID(timeEntry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model == nil {
		t.Fatal("time entry not found")
	}

	timeEntry.Name = "foo"
	if err := updateTimeEntry(timeEntry); err != nil {
		t.Fatal(err)
	}

	models, err := selectTimeEntriesByUserAndCompany(user.ID, company.ID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 {
		t.Fatal("time entry not found")
	}

	if err := deleteTimeEntry(timeEntry.ID); err != nil {
		t.Fatal(err)
	}
}

func TestStats(t *testing.T) {
	_, err := selectStats()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPopulateUser(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone31@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	if err := populateUser(user); err != nil {
		t.Fatal(err)
	}
}

func TestAssignOrganization(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone32@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	person := Person{}
	person.CompanyID = company.ID
	person.FirstName = "John"
	person.OwnerID = user.ID
	person.Name = "Smith"
	if err := insertPerson(&person); err != nil {
		t.Fatal(err)
	}

	person.OrgName = "new organization"
	if err := assignOrganization(&person, user); err != nil {
		t.Fatal(err)
	}
	if person.OrgID == "" {
		t.Fatal("Organization should have been created")
	}
}

func TestAssignPerson(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone33@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	activity := Activity{}
	activity.Name = "Work"
	activity.UserID = user.ID
	activity.CompanyID = company.ID
	if err := insertActivity(&activity); err != nil {
		t.Fatal(err)
	}

	activity.PersonName = "a new person"
	if err := assignPerson(&activity, user); err != nil {
		t.Fatal(err)
	}
}

func TestUndelete(t *testing.T) {
	user := User{
		Email: "someone40@somewhere.com",
	}

	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}
	user.ActiveCompanyID = company.ID

	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	companyUser := CompanyUser{
		CompanyID: company.ID,
		UserID:    user.ID,
	}
	if err := insertCompanyUser(&companyUser); err != nil {
		t.Fatal(err)
	}

	do := DeletedObject{
		Type: "company_users",
	}
	do.ID = companyUser.ID

	if err := deleteCompanyUser(companyUser.ID); err != nil {
		t.Fatal(err)
	}

	if err := undelete(do); err != nil {
		t.Fatal(err)
	}

	model, err := selectCompanyUserByID(do.ID)
	if err != nil {
		t.Fatal(err)
	}
	if model.DeletedAt != nil {
		t.Fatal("model should be un-deleted")
	}
}

func TestCompanyUser(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone41@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	companyUser := CompanyUser{
		CompanyID: company.ID,
		UserID:    user.ID,
	}
	if err := insertCompanyUser(&companyUser); err != nil {
		t.Fatal(err)
	}

	companies, err := selectCompaniesByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(companies) != 1 {
		t.Fatal(err)
	}
}

func TestUserEvent(t *testing.T) {
	company := Company{}
	company.Name = "supercompany"
	if err := insertCompany(&company); err != nil {
		t.Fatal(err)
	}

	user := User{
		Email:           "someone300@somewhere.com",
		ActiveCompanyID: company.ID,
	}
	if err := insertUser(&user); err != nil {
		t.Fatal(err)
	}

	userEvent := UserEvent{
		UserID: user.ID,
	}
	userEvent.Name = "addPerson"
	if err := insertUserEvent(&userEvent); err != nil {
		t.Fatal(err)
	}

	models, err := selectUserEventsByUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 {
		t.Fatal("user event not found")
	}
}
